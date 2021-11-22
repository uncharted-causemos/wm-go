package storage

import (
	"fmt"
	"math"
	"strconv"
	"time"

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
		return s.transformPerCapitaTimeseries(timeseries, config)
	default:
		return timeseries, nil
	}
}

// TransformRegionAggregation returns transformed regional data for ALL admin regions at ONE timestamp
func (s *Storage) TransformRegionAggregation(data *wm.ModelOutputRegionalAdmins, timestamp string, config wm.TransformConfig) (*wm.ModelOutputRegionalAdmins, error) {
	// op := "Storage.TransformRegionAggregation"

	switch config.Transform {
	case wm.TransformPerCapita:
		return s.transformPerCapitaRegionAggregation(data, timestamp)
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
		return s.transformPerCapitaQualifierRegional(data, timestamp)
	case wm.TransformNormalization:
		return s.normalizeQualifierRegional(data)
	default:
		return data, nil
	}
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

	// Calculate Per capita with given timeseries and population data
	var result []*wm.TimeseriesValue
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
		valuePercapita := v.Value / population
		result = append(result, &wm.TimeseriesValue{Timestamp: v.Timestamp, Value: valuePercapita})
	}
	return result, nil
}

func (s *Storage) transformPerCapitaRegionAggregation(data *wm.ModelOutputRegionalAdmins, timestamp string) (*wm.ModelOutputRegionalAdmins, error) {
	op := "Storage.transformPerCapitaRegionAggregation"

	pLookup, err := s.getRegionalPopulation(timestamp)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	resultAdminDataList := [4][]wm.ModelOutputAdminData{}
	for i, d := range [][]wm.ModelOutputAdminData{data.Country, data.Admin1, data.Admin2, data.Admin3} {
		for _, v := range d {
			if p, ok := pLookup[v.ID]; ok && p != 0 {
				resultAdminDataList[i] = append(resultAdminDataList[i], wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / p})
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

func (s *Storage) transformPerCapitaQualifierRegional(data *wm.ModelOutputRegionalQualifiers, timestamp string) (*wm.ModelOutputRegionalQualifiers, error) {
	op := "Storage.transformPerCapitaQualifierRegional"

	pLookup, err := s.getRegionalPopulation(timestamp)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	calcPerCapita := func(values map[string]float64, population float64) map[string]float64 {
		newValues := make(map[string]float64)
		for k, v := range values {
			newValues[k] = v / population
		}
		return newValues
	}

	result := &wm.ModelOutputRegionalQualifiers{
		Country: []wm.ModelOutputRegionQualifierBreakdown{},
		Admin1:  []wm.ModelOutputRegionQualifierBreakdown{},
		Admin2:  []wm.ModelOutputRegionQualifierBreakdown{},
		Admin3:  []wm.ModelOutputRegionQualifierBreakdown{},
	}

	// Calculate per capita value
	for _, v := range data.Country {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Country = append(result.Country, wm.ModelOutputRegionQualifierBreakdown{ID: v.ID, Values: calcPerCapita(v.Values, p)})
		}
	}
	for _, v := range data.Admin1 {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Admin1 = append(result.Admin1, wm.ModelOutputRegionQualifierBreakdown{ID: v.ID, Values: calcPerCapita(v.Values, p)})
		}
	}
	for _, v := range data.Admin2 {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Admin2 = append(result.Admin2, wm.ModelOutputRegionQualifierBreakdown{ID: v.ID, Values: calcPerCapita(v.Values, p)})
		}
	}
	for _, v := range data.Admin3 {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Admin3 = append(result.Admin3, wm.ModelOutputRegionQualifierBreakdown{ID: v.ID, Values: calcPerCapita(v.Values, p)})
		}
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

	for _, v := range regionalPopulation.Country {
		regionPopLookup[v.ID] = v.Value
	}
	for _, v := range regionalPopulation.Admin1 {
		regionPopLookup[v.ID] = v.Value
	}
	for _, v := range regionalPopulation.Admin2 {
		regionPopLookup[v.ID] = v.Value
	}
	for _, v := range regionalPopulation.Admin3 {
		regionPopLookup[v.ID] = v.Value
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
