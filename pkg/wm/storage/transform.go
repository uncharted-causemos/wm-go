package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

var populationRegionalLookupCache map[int]map[string]float64 = make(map[int]map[string]float64)

func getPopulationDataAvailableYears() []int {
	return []int{2020}
}

func getPopulationDatacubeParams() wm.DatacubeParams {
	return wm.DatacubeParams{
		DataID:          "430d621b-c4cd-4cb2-8d21-1c590522602c",
		RunID:           "indicator",
		Feature:         "Population Count",
		Resolution:      "year",
		TemporalAggFunc: "sum",
		SpatialAggFunc:  "sum",
	}
}

// TransformOutputTimeseriesByRegion returns transformed timeseries data
func (s *Storage) TransformOutputTimeseriesByRegion(timeseries []*wm.TimeseriesValue, config wm.TransformConfig) ([]*wm.TimeseriesValue, error) {
	// op := "Storage.TransformOutputTimeseriesByRegion"
	switch config.Transform {
	case wm.TransformPerCapita:
		config.ScaleFactor = 1
		return s.transformPerCapitaTimeseries(timeseries, config)
	case wm.TransformPerCapita1K:
		config.ScaleFactor = 1_000
		return s.transformPerCapitaTimeseries(timeseries, config)
	case wm.TransformPerCapita1M:
		config.ScaleFactor = 1_000_000
		return s.transformPerCapitaTimeseries(timeseries, config)
	case wm.TransformNormalization:
		return s.normalizeRegionalTimeseries(timeseries, config)
	default:
		return timeseries, nil
	}
}

// TransformOutputQualifierTimeseriesByRegion returns transformed qualifier timeseries data
func (s *Storage) TransformOutputQualifierTimeseriesByRegion(data []*wm.ModelOutputQualifierTimeseries, config wm.TransformConfig) ([]*wm.ModelOutputQualifierTimeseries, error) {
	// op := "Storage.TransformOutputQualifierTimeseriesByRegion"
	result := make([]*wm.ModelOutputQualifierTimeseries, 0)
	for _, qSeries := range data {
		series, err := s.TransformOutputTimeseriesByRegion(qSeries.Timeseries, config)
		if err != nil {
			return nil, err
		}
		result = append(result, &wm.ModelOutputQualifierTimeseries{
			Name: qSeries.Name, Timeseries: series})
	}
	return result, nil
}

// TransformRegionAggregationByAdminLevel returns transformed regional data for given admin level at ONE timestamp
func (s *Storage) TransformRegionAggregationByAdminLevel(data *wm.ModelOutputRegional, config wm.TransformConfig) (*wm.ModelOutputRegional, error) {
	// op := "Storage.TransformRegionAggregationByAdminLevel"
	switch config.Transform {
	case wm.TransformNormalization:
		return s.normalizeRegionAggregationByAdminLevel(data, config)
	default:
		return data, nil
	}
}

// TransformRegionAggregation returns transformed regional data for ALL admin regions at ONE timestamp
func (s *Storage) TransformRegionAggregation(data *wm.ModelOutputRegionalAdmins, timestamp string, config wm.TransformConfig) (*wm.ModelOutputRegionalAdmins, error) {
	// op := "Storage.TransformRegionAggregation"

	switch config.Transform {
	case wm.TransformPerCapita:
		config.ScaleFactor = 1
		return s.transformPerCapitaRegionAggregation(data, timestamp, config.ScaleFactor)
	case wm.TransformPerCapita1K:
		config.ScaleFactor = 1_000
		return s.transformPerCapitaRegionAggregation(data, timestamp, config.ScaleFactor)
	case wm.TransformPerCapita1M:
		config.ScaleFactor = 1_000_000
		return s.transformPerCapitaRegionAggregation(data, timestamp, config.ScaleFactor)
	case wm.TransformNormalization:
		return s.normalizeRegionAggregation(data)
	default:
		return data, nil
	}
}

// TransformQualifierRegional returns transformed qualifier regional data for ALL admin regions at ONE timestamp
func (s *Storage) TransformQualifierRegional(data *wm.ModelOutputRegionalQualifiers, timestamp string, config wm.TransformConfig) (*wm.ModelOutputRegionalQualifiers, error) {
	// op := "Storage.TransformQualifierRegional"

	switch config.Transform {
	case wm.TransformPerCapita:
		config.ScaleFactor = 1
		return s.transformPerCapitaQualifierRegional(data, timestamp, config.ScaleFactor)
	case wm.TransformPerCapita1K:
		config.ScaleFactor = 1_000
		return s.transformPerCapitaQualifierRegional(data, timestamp, config.ScaleFactor)
	case wm.TransformPerCapita1M:
		config.ScaleFactor = 1_000_000
		return s.transformPerCapitaQualifierRegional(data, timestamp, config.ScaleFactor)
	case wm.TransformNormalization:
		return s.normalizeQualifierRegional(data)
	default:
		return data, nil
	}
}

func (s *Storage) normalizeRegionalTimeseries(timeseries []*wm.TimeseriesValue, config wm.TransformConfig) ([]*wm.TimeseriesValue, error) {
	op := "Storage.normalizeRegionalTimeseries"

	params := config.DatacubeParams

	// Get admin level from region id
	adminLevels := []wm.AdminLevel{wm.AdminLevelCountry, wm.AdminLevel1, wm.AdminLevel2, wm.AdminLevel3}
	regionId := config.RegionID
	adminLevelNum := len(strings.Split(string(regionId), "__")) - 1

	// Fetch min max from precomputed extrema file and get min and max value across the region and timestamp
	min, max, err := s.getRegionalMinMaxFromS3(params, adminLevels[adminLevelNum])
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	var result []*wm.TimeseriesValue
	for _, v := range timeseries {
		result = append(result, &wm.TimeseriesValue{Timestamp: v.Timestamp, Value: normalize(v.Value, min, max)})
	}

	return result, nil
}

func (s *Storage) transformPerCapitaTimeseries(timeseries []*wm.TimeseriesValue, config wm.TransformConfig) ([]*wm.TimeseriesValue, error) {
	op := "Storage.transformPerCapitaTimeseries"
	populationTimeseries, err := s.GetOutputTimeseriesByRegion(getPopulationDatacubeParams(), config.RegionID)
	if err != nil {
		if wm.ErrorCode(err) == wm.ENOTFOUND {
			// if population data is not found, raise an internal server error
			return nil, &wm.Error{Code: wm.EINTERNAL, Op: op, Err: err}
		}
		return nil, &wm.Error{Op: op, Err: err}
	}

	var scaleFactor float64 = 1
	if math.Abs(config.ScaleFactor) > 0 {
		scaleFactor = config.ScaleFactor
	}

	// Calculate Per capita with given timeseries and population data
	result := make([]*wm.TimeseriesValue, 0)
	for _, v := range timeseries {
		year := time.UnixMilli(v.Timestamp).UTC().Year()
		var population float64
		// find the population of matching or last available year
		for i, s := range populationTimeseries {
			py := time.UnixMilli(s.Timestamp).UTC().Year()
			if year == py {
				// found the matching year
				population = populationTimeseries[i].Value
				break
			} else if year < py {
				// could not find the matching year. use the last available year
				population = populationTimeseries[int(math.Max(0, float64(i-1)))].Value
				break
			} else if i == len(populationTimeseries)-1 && py < year {
				// if given year is greater than available population data year, use the last available year's value
				population = populationTimeseries[len(populationTimeseries)-1].Value
			}
		}
		valuePercapita := (v.Value / population) * scaleFactor
		result = append(result, &wm.TimeseriesValue{Timestamp: v.Timestamp, Value: valuePercapita})
	}
	return result, nil
}

func (s *Storage) transformPerCapitaRegionAggregation(data *wm.ModelOutputRegionalAdmins, timestamp string, scaleFactor float64) (*wm.ModelOutputRegionalAdmins, error) {
	op := "Storage.transformPerCapitaRegionAggregation"

	pLookup, err := s.getRegionalPopulation(timestamp)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	if scaleFactor == 0 {
		scaleFactor = 1
	}

	resultAdminDataList := [4][]wm.ModelOutputAdminData{}
	for i, d := range [][]wm.ModelOutputAdminData{data.Country, data.Admin1, data.Admin2, data.Admin3} {
		for _, v := range d {
			if p, ok := pLookup[v.ID]; ok && p != 0 {
				resultAdminDataList[i] = append(resultAdminDataList[i], wm.ModelOutputAdminData{ID: v.ID, Value: (v.Value / p) * scaleFactor})
			}
		}
	}
	result := &wm.ModelOutputRegionalAdmins{
		Country: resultAdminDataList[0],
		Admin1:  resultAdminDataList[1],
		Admin2:  resultAdminDataList[2],
		Admin3:  resultAdminDataList[3],
	}
	return result, nil
}

func (s *Storage) transformPerCapitaQualifierRegional(data *wm.ModelOutputRegionalQualifiers, timestamp string, scaleFactor float64) (*wm.ModelOutputRegionalQualifiers, error) {
	op := "Storage.transformPerCapitaQualifierRegional"

	pLookup, err := s.getRegionalPopulation(timestamp)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	if scaleFactor == 0 {
		scaleFactor = 1
	}

	calcPerCapita := func(values map[string]float64, population float64, scaleFactor float64) map[string]float64 {
		newValues := make(map[string]float64)
		for k, v := range values {
			newValues[k] = (v / population) * scaleFactor
		}
		return newValues
	}

	resultAdminDataList := [4][]wm.ModelOutputRegionQualifierBreakdown{}
	for i, d := range [][]wm.ModelOutputRegionQualifierBreakdown{data.Country, data.Admin1, data.Admin2, data.Admin3} {
		for _, v := range d {
			if p, ok := pLookup[v.ID]; ok && p != 0 {
				resultAdminDataList[i] = append(resultAdminDataList[i], wm.ModelOutputRegionQualifierBreakdown{ID: v.ID, Values: calcPerCapita(v.Values, p, scaleFactor)})
			}
		}
	}
	result := &wm.ModelOutputRegionalQualifiers{
		Country: resultAdminDataList[0],
		Admin1:  resultAdminDataList[1],
		Admin2:  resultAdminDataList[2],
		Admin3:  resultAdminDataList[3],
	}
	return result, nil
}

// getRegionalPopulation returns a lookup table that maps region id to population of the region for the year that matches with given timestamp
func (s *Storage) getRegionalPopulation(timestamp string) (map[string]float64, error) {
	op := "Storage.getRegionalPopulation"

	year, err := getAvailablePopulationDataYear(timestamp)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	// Check in memory cache for the data and return it if exists
	if data, ok := populationRegionalLookupCache[year]; ok {
		return data, nil
	}

	pTimestamp := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	regionalPopulation, err := s.GetRegionAggregation(getPopulationDatacubeParams(), fmt.Sprintf("%d", pTimestamp))
	if err != nil {
		if wm.ErrorCode(err) == wm.ENOTFOUND {
			// if population data is not found, raise an internal server error
			return nil, &wm.Error{Code: wm.EINTERNAL, Op: op, Err: err}
		}
		return nil, &wm.Error{Op: op, Err: err}
	}

	// Store population by region id
	regionPopLookup := make(map[string]float64)

	for _, d := range [][]wm.ModelOutputAdminData{regionalPopulation.Country, regionalPopulation.Admin1, regionalPopulation.Admin2, regionalPopulation.Admin3} {
		for _, v := range d {
			regionPopLookup[v.ID] = v.Value
		}
	}

	// cache population lookup data by year into memory
	populationRegionalLookupCache[year] = regionPopLookup
	return regionPopLookup, nil
}

// getAvailablePopulationDataYear returns matching or last available year of the population data for given timstamp
func getAvailablePopulationDataYear(timestamp string) (int, error) {
	op := "getAvailablePopulationDataYear"

	ts, err := strconv.Atoi(timestamp)
	if err != nil {
		return 0, &wm.Error{Op: op, Err: err}
	}
	year := time.UnixMilli(int64(ts)).UTC().Year()
	availableYears := getPopulationDataAvailableYears()
	// get available year for the population data
	pYear := availableYears[len(availableYears)-1]
	for i, y := range availableYears {
		if year == y {
			pYear = y
			break
		} else if i > 0 && year < y {
			pYear = availableYears[i-1]
			break
		}
	}
	return pYear, nil
}

func (s *Storage) normalizeRegionAggregationByAdminLevel(data *wm.ModelOutputRegional, config wm.TransformConfig) (*wm.ModelOutputRegional, error) {
	op := "Storage.normalizeRegionAggregationByAdminLevel"

	result := make(wm.ModelOutputRegional)

	for adminLevel := range *data {
		min, max, err := s.getRegionalMinMaxFromS3(config.DatacubeParams, adminLevel)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		for _, d := range (*data)[adminLevel] {
			result[adminLevel] = append(result[adminLevel], wm.ModelOutputAdminData{ID: d.ID, Value: normalize(d.Value, min, max)})
		}
	}
	return &result, nil
}

func (s *Storage) normalizeRegionAggregation(data *wm.ModelOutputRegionalAdmins) (*wm.ModelOutputRegionalAdmins, error) {
	// op := "Storage.normalizeRegionAggregation"

	resultAdminDataList := [4][]wm.ModelOutputAdminData{}
	for i, d := range [][]wm.ModelOutputAdminData{data.Country, data.Admin1, data.Admin2, data.Admin3} {
		min, max := getMinMax(d)
		for _, v := range d {
			resultAdminDataList[i] = append(resultAdminDataList[i], wm.ModelOutputAdminData{ID: v.ID, Value: normalize(v.Value, min, max)})
		}
	}
	result := &wm.ModelOutputRegionalAdmins{
		Country: resultAdminDataList[0],
		Admin1:  resultAdminDataList[1],
		Admin2:  resultAdminDataList[2],
		Admin3:  resultAdminDataList[3],
	}
	return result, nil
}

func (s *Storage) normalizeQualifierRegional(data *wm.ModelOutputRegionalQualifiers) (*wm.ModelOutputRegionalQualifiers, error) {
	// op := "Storage.normalizeQualifierRegional"

	resultAdminDataList := [4][]wm.ModelOutputRegionQualifierBreakdown{}
	for i, d := range [][]wm.ModelOutputRegionQualifierBreakdown{data.Country, data.Admin1, data.Admin2, data.Admin3} {
		minMax := getMinMaxQualifierBreakdown(d)
		for _, v := range d {
			resultAdminDataList[i] = append(resultAdminDataList[i], wm.ModelOutputRegionQualifierBreakdown{ID: v.ID, Values: normalizeValues(v.Values, minMax)})
		}
	}
	result := &wm.ModelOutputRegionalQualifiers{
		Country: resultAdminDataList[0],
		Admin1:  resultAdminDataList[1],
		Admin2:  resultAdminDataList[2],
		Admin3:  resultAdminDataList[3],
	}
	return result, nil
}

// getRegionalMinMaxFromS3 fetches regional min max values from precomputed extrema file from s3
func (s *Storage) getRegionalMinMaxFromS3(params *wm.DatacubeParams, adminLevel wm.AdminLevel) (float64, float64, error) {
	op := "Storage.getRegionalMinMaxFromS3"
	key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/stats/default/extrema.json",
		params.DataID, params.RunID, params.Resolution, params.Feature, adminLevel)
	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return 0, 0, &wm.Error{Op: op, Err: err}
	}

	var extrema wm.RegionalExtrema
	err = json.Unmarshal(buf, &extrema)
	if err != nil {
		return 0, 0, &wm.Error{Op: op, Err: err}
	}

	aggFuncKey := fmt.Sprintf("s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
	min := extrema.Min[aggFuncKey][0].Value
	max := extrema.Max[aggFuncKey][0].Value
	return min, max, nil
}

// getMinMax gets local min and max values from provided admin data
func getMinMax(adminData []wm.ModelOutputAdminData) (float64, float64) {
	if len(adminData) == 0 {
		return 0, 0
	}
	min := adminData[0].Value
	max := min
	for _, v := range adminData {
		max = math.Max(max, v.Value)
		min = math.Min(min, v.Value)
	}
	return min, max
}

func getMinMaxQualifierBreakdown(data []wm.ModelOutputRegionQualifierBreakdown) map[string]struct {
	min float64
	max float64
} {
	result := make(map[string]struct {
		min float64
		max float64
	})
	if len(data) == 0 {
		return result
	}
	for _, d := range data {
		for key, val := range d.Values {
			min := result[key].min
			max := result[key].max
			if _, ok := result[key]; !ok {
				min, max = val, val
			}
			result[key] = struct {
				min float64
				max float64
			}{min: math.Min(min, val), max: math.Max(max, val)}
		}
	}
	return result
}

func normalize(val, min, max float64) float64 {
	if min == 0 && max == 0 {
		return 0
	} else if min == max {
		return 1
	}
	return (val - min) / (max - min)
}

func normalizeValues(vals map[string]float64, minMax map[string]struct {
	min float64
	max float64
}) map[string]float64 {
	result := make(map[string]float64)
	for k, v := range vals {
		result[k] = normalize(v, minMax[k].min, minMax[k].max)
	}
	return result
}
