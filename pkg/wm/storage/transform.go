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

// TransformOutputQualifierTimeseriesByRegion returns transformed qualifier timeseries data
func (s *Storage) TransformOutputQualifierTimeseriesByRegion(data []*wm.ModelOutputQualifierTimeseries, config wm.TransformConfig) ([]*wm.ModelOutputQualifierTimeseries, error) {
	// op := "Storage.TransformOutputQualifierTimeseriesByRegion"
	var result []*wm.ModelOutputQualifierTimeseries
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

// TransformRegionAggregation returns transformed regional data for ALL admin regions at ONE timestamp
func (s *Storage) TransformRegionAggregation(data *wm.ModelOutputRegionalAdmins, timestamp string, config wm.TransformConfig) (*wm.ModelOutputRegionalAdmins, error) {
	op := "Storage.TransformRegionAggregation"

	switch config.Transform {
	case wm.TransformPerCapita:
		return s.transformPerCapitaRegionAggregation(data, timestamp)
	case wm.TransformNormalization:
		return nil, &wm.Error{Op: op, Err: fmt.Errorf("Not yet implemented")}
	default:
		return data, nil
	}
}

// TransformQualifierRegional returns transformed qualifier regional data for ALL admin regions at ONE timestamp
func (s *Storage) TransformQualifierRegional(data *wm.ModelOutputRegionalQualifiers, timestamp string, config wm.TransformConfig) (*wm.ModelOutputRegionalQualifiers, error) {
	op := "Storage.TransformQualifierRegional"

	switch config.Transform {
	case wm.TransformPerCapita:
		return s.transformPerCapitaQualifierRegional(data, timestamp)
	case wm.TransformNormalization:
		return nil, &wm.Error{Op: op, Err: fmt.Errorf("Not yet implemented")}
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

	result := &wm.ModelOutputRegionalAdmins{
		Country: []wm.ModelOutputAdminData{},
		Admin1:  []wm.ModelOutputAdminData{},
		Admin2:  []wm.ModelOutputAdminData{},
		Admin3:  []wm.ModelOutputAdminData{},
	}

	// Calculate per capita value
	for _, v := range data.Country {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Country = append(result.Country, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / p})
		}
	}
	for _, v := range data.Admin1 {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Admin1 = append(result.Admin1, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / p})
		}
	}
	for _, v := range data.Admin2 {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Admin2 = append(result.Admin2, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / p})
		}
	}
	for _, v := range data.Admin3 {
		if p, ok := pLookup[v.ID]; ok && p != 0 {
			result.Admin3 = append(result.Admin3, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / p})
		}
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
