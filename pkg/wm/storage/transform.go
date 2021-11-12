package storage

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// TransformOutputTimeseriesByRegion returns transformed timeseries data
func (s *Storage) TransformOutputTimeseriesByRegion(timeseries []*wm.TimeseriesValue, config wm.TransformConfig) ([]*wm.TimeseriesValue, error) {
	if config.Transform == "percapita" {
		return s.transformPerCapitaTimeseries(timeseries, config)
	}
	var series []*wm.TimeseriesValue
	return series, nil
}

// TransformRegionAggregation returns transformed regional data for ALL admin regions at ONE timestamp
func (s *Storage) TransformRegionAggregation(data *wm.ModelOutputRegionalAdmins, timestamp string, config wm.TransformConfig) (*wm.ModelOutputRegionalAdmins, error) {
	if config.Transform == "percapita" {
		return s.transformPerCapitaRegionAggregation(data, timestamp)
	}
	var result *wm.ModelOutputRegionalAdmins
	return result, nil
}

func (s *Storage) transformPerCapitaTimeseries(timeseries []*wm.TimeseriesValue, config wm.TransformConfig) ([]*wm.TimeseriesValue, error) {
	op := "Storage.transformPerCapitaTimeseries"
	populationTimeseries, err := s.GetOutputTimeseriesByRegion(wm.DatacubeParams{
		DataID:          "test",
		RunID:           "indicator",
		Feature:         "test",
		Resolution:      "year",
		TemporalAggFunc: "sum",
		SpatialAggFunc:  "sum",
	}, config.RegionID)
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
		year := time.UnixMilli(v.Timestamp).Year()
		var population float64
		// find population of matching or last available year
		for i, s := range populationTimeseries {
			py := time.UnixMilli(s.Timestamp).Year()
			if year == py {
				// found matching year
				population = populationTimeseries[i].Value
				break
			} else if year < py {
				// could not found maching year. use previous available year
				population = populationTimeseries[int(math.Max(0, float64(i-1)))].Value
				break
			} else if i == len(populationTimeseries)-1 && py < year {
				// if given year is greater than available population data year, use last year's value
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
		result.Country = append(result.Country, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / pLookup[v.ID]})
	}
	for _, v := range data.Admin1 {
		result.Admin1 = append(result.Admin1, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / pLookup[v.ID]})
	}
	for _, v := range data.Admin2 {
		result.Admin2 = append(result.Admin2, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / pLookup[v.ID]})
	}
	for _, v := range data.Admin3 {
		result.Admin3 = append(result.Admin3, wm.ModelOutputAdminData{ID: v.ID, Value: v.Value / pLookup[v.ID]})
	}

	return result, nil
}

func (s *Storage) getRegionalPopulation(timestamp string) (map[string]float64, error) {
	op := "Storage.getRegionalPopulation"
	ts, err := strconv.Atoi(timestamp)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	year := time.UnixMilli(int64(ts)).Year()

	// get available year for the population data
	// TODO: make this a helper function
	availableYears := [5]int{2000, 2005, 2010, 2015, 2020}
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

	// TODO: cache population lookup data for specific year into memory
	pTimestamp := time.Date(pYear, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	regionalPopulation, err := s.GetRegionAggregation(wm.DatacubeParams{
		DataID:          "test",
		RunID:           "indicator",
		Feature:         "test",
		Resolution:      "year",
		TemporalAggFunc: "sum",
		SpatialAggFunc:  "sum",
	}, fmt.Sprintf("%d", pTimestamp))

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

	return regionPopLookup, nil
}
