package api

import (
	"net/http"

	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type outputStatsResponse struct {
	*wm.OutputStatWithZoom
}

// Render allows to satisfy the render.Renderer interface.
func (msr *outputStatsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelRegionalOutputStatsResponse struct {
	*wm.ModelRegionalOutputStat
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelRegionalOutputStatsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputRegionalTimeseries struct {
	*wm.ModelOutputRegionalTimeSeries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputRegionalTimeseries) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputTimeseriesValue struct {
	*wm.TimeseriesValue
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputTimeseriesValue) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputBulkRegionalData struct {
	*wm.ModelOutputBulkAggregateRegionalAdmins
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputBulkRegionalData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputRegionalData struct {
	*wm.ModelOutputRegionalAdmins
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputRegionalData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputRawDataPoint struct {
	*wm.ModelOutputRawDataPoint
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputRawDataPoint) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputQualifierTimeseriesResponse struct {
	*wm.ModelOutputQualifierTimeseries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputQualifierTimeseriesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputQualifierValuesResponse struct {
	*wm.ModelOutputQualifierBreakdown
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputQualifierValuesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getRegionalDataOutputStats(w http.ResponseWriter, r *http.Request) error {
	op := "api.getRegionalDataOutputStats"
	params := getDatacubeParams(r)
	stats, err := a.dataOutput.GetRegionalOutputStats(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.Render(w, r, &modelRegionalOutputStatsResponse{stats})
	return nil
}

func (a *api) getDataOutputStats(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputStats"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	stats, err := a.dataOutput.GetOutputStats(params, timestamp)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, stat := range stats {
		list = append(list, &outputStatsResponse{stat})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getBulkDataOutputTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getBulkDataOutputTimeseries"
	regionIDs, err := getRegionIDsFromBody(r)
	params := getDatacubeParams(r)
	transform := getTransform(r)
	var regionalTimeSeries []*wm.ModelOutputRegionalTimeSeries

	if len(regionIDs) == 0 {
		data, err := a.dataOutput.GetOutputTimeseries(params)
		if err != nil {
			return &wm.Error{Op: op, Err: err}
		}
		regionalTimeSeries = []*wm.ModelOutputRegionalTimeSeries{
			{RegionID: "", Timeseries: data},
		}
	} else {
		regionalTimeSeries = make([]*wm.ModelOutputRegionalTimeSeries, len(regionIDs))
		for i := 0; i < len(regionIDs); i++ {
			regionalTS, err := a.getTimeSeries(regionIDs[i], params, transform)
			if err != nil {
				if wm.ErrorCode(err) == wm.ENOTFOUND {
					regionalTimeSeries[i] = &wm.ModelOutputRegionalTimeSeries{
						RegionID:   regionIDs[i],
						Timeseries: []*wm.TimeseriesValue{},
					}
					continue
				}
				return &wm.Error{Op: op, Err: err}
			}
			regionalTimeSeries[i] = &wm.ModelOutputRegionalTimeSeries{
				RegionID:   regionIDs[i],
				Timeseries: regionalTS,
			}
		}
	}
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}

	list := []render.Renderer{}
	for _, timeseries := range regionalTimeSeries {
		list = append(list, &modelOutputRegionalTimeseries{timeseries})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputTimeseries"
	params := getDatacubeParams(r)
	regionID := getRegionID(r)
	transform := getTransform(r)
	var timeseries []*wm.TimeseriesValue
	var err error

	timeseries, err = a.getTimeSeries(regionID, params, transform)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, point := range timeseries {
		list = append(list, &modelOutputTimeseriesValue{point})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getTimeSeries(regionID string, params wm.DatacubeParams, transform wm.Transform) ([]*wm.TimeseriesValue, error) {
	var timeseries []*wm.TimeseriesValue
	var err error

	if regionID == "" {
		timeseries, err = a.dataOutput.GetOutputTimeseries(params)
		if err != nil {
			return nil, err
		}
		return timeseries, nil
	}
	timeseries, err = a.dataOutput.GetOutputTimeseriesByRegion(params, regionID)
	if err != nil {
		return nil, err
	}
	if transform != "" {
		timeseries, err = a.dataOutput.TransformOutputTimeseriesByRegion(timeseries, wm.TransformConfig{Transform: transform, RegionID: regionID})
		if err != nil {
			return nil, err
		}
	}

	return timeseries, nil
}

func (a *api) getBulkDataOutputRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getBulkDataOutputRegional"
	timestamps, err := getTimestamps(r)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	params := getDatacubeParams(r)
	transform := getTransform(r)
	aggForSelect := getAggForSelect(r)
	aggForAll := getAggForAll(r)
	bulkData := wm.ModelOutputBulkAggregateRegionalAdmins{}

	bulkRegionalData := make([]wm.ModelOutputBulkRegionalAdmins, len(timestamps.Timestamps))
	var totalBulkRegionalData []wm.ModelOutputBulkRegionalAdmins

	for i, timestamp := range timestamps.Timestamps {
		data, err := a.getRegionAggregation(params, timestamp, transform)
		bulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
			Timestamp:                 timestamp,
			ModelOutputRegionalAdmins: *data,
		}
		if err != nil {
			return &wm.Error{Op: op, Err: err}
		}
	}
	bulkData.ModelOutputBulkRegionalAdmins = bulkRegionalData
	if len(timestamps.AllTimestamps) != 0 {
		// avoid looking through subset of timestamps an unnecessary amount of times
		subsetTimeStamps := map[string]wm.ModelOutputBulkRegionalAdmins{}
		for _, timestamp := range bulkRegionalData {
			subsetTimeStamps[timestamp.Timestamp] = timestamp
		}
		// hold only data that wouldn't already be in bulkRegionalData
		totalBulkRegionalData = make([]wm.ModelOutputBulkRegionalAdmins, len(timestamps.AllTimestamps))
		for i, timestamp := range timestamps.AllTimestamps {
			val, ok := subsetTimeStamps[timestamp]
			if !ok {
				data, err := a.getRegionAggregation(params, timestamp, transform)
				totalBulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
					Timestamp:                 timestamp,
					ModelOutputRegionalAdmins: *data,
				}
				if err != nil {
					return &wm.Error{Op: op, Err: err}
				}
			} else {
				// copy it over to apply whatever aggregation we will need to
				totalBulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
					Timestamp:                 val.Timestamp,
					ModelOutputRegionalAdmins: val.ModelOutputRegionalAdmins,
				}
			}
		}
	}

	if aggForSelect == "mean" {
		bulkData.SelectAgg = calculateMeanAggregationForBulkRegionalData(bulkRegionalData)
	}

	if aggForAll == "mean" && len(timestamps.AllTimestamps) != 0 {
		bulkData.AllAgg = calculateMeanAggregationForBulkRegionalData(totalBulkRegionalData)
	}

	render.Render(w, r, &modelOutputBulkRegionalData{&bulkData})
	return nil
}

func calculateMeanAggregationForBulkRegionalData(bulkRegionalData []wm.ModelOutputBulkRegionalAdmins) wm.ModelOutputRegionalAdmins {
	aggData := wm.ModelOutputRegionalAdmins{}
	countryAgg := map[string]*[2]float64{}
	admin1Agg := map[string]*[2]float64{}
	admin2Agg := map[string]*[2]float64{}
	admin3Agg := map[string]*[2]float64{}
	for _, regionalData := range bulkRegionalData {
		sumAggregation(regionalData.Country, countryAgg)
		sumAggregation(regionalData.Admin1, admin1Agg)
		sumAggregation(regionalData.Admin2, admin2Agg)
		sumAggregation(regionalData.Admin3, admin3Agg)
	}
	aggData.Country = applyMeanAggregation(countryAgg)
	aggData.Admin1 = applyMeanAggregation(admin1Agg)
	aggData.Admin2 = applyMeanAggregation(admin2Agg)
	aggData.Admin3 = applyMeanAggregation(admin3Agg)
	return aggData
}

func sumAggregation(regionProperty []wm.ModelOutputAdminData, aggDict map[string]*[2]float64) {
	for _, property := range regionProperty {
		val, ok := aggDict[property.ID]
		if ok {
			val[0] += property.Value
			val[1]++
		} else {
			aggDict[property.ID] = &[2]float64{property.Value, 1}
		}
	}
}

func applyMeanAggregation(countryAgg map[string]*[2]float64) []wm.ModelOutputAdminData {
	aggregations := make([]wm.ModelOutputAdminData, len(countryAgg))
	i := 0
	for key, val := range countryAgg {
		aggregations[i] = wm.ModelOutputAdminData{ID: key, Value: val[0] / val[1]}
		i++
	}
	return aggregations
}

func (a *api) getDataOutputRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRegional"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	transform := getTransform(r)
	data, err := a.getRegionAggregation(params, timestamp, transform)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.Render(w, r, &modelOutputRegionalData{data})
	return nil
}

func (a *api) getRegionAggregation(params wm.DatacubeParams, timestamp string, transform wm.Transform) (*wm.ModelOutputRegionalAdmins, error) {
	data, err := a.dataOutput.GetRegionAggregation(params, timestamp)
	if err != nil {
		return nil, err
	}
	if transform != "" {
		data, err = a.dataOutput.TransformRegionAggregation(data, timestamp, wm.TransformConfig{Transform: transform})
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (a *api) getDataOutputRaw(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRaw"
	params := getDatacubeParams(r)
	data, err := a.dataOutput.GetRawData(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, point := range data {
		list = append(list, &modelOutputRawDataPoint{point})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputRegionLists(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRegionLists"
	params := getRegionListsParams(r)
	data, err := a.dataOutput.GetRegionLists(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getDataOutputQualifierCounts(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierCounts"
	params := getQualifierInfoParams(r)
	data, err := a.dataOutput.GetQualifierCounts(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getDataOutputQualifierLists(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierLists"
	params := getQualifierInfoParams(r)
	qualifiers := getQualifierNames(r)
	data, err := a.dataOutput.GetQualifierLists(params, qualifiers)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getDataOutputQualifierTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierTimeseries"
	params := getDatacubeParams(r)
	regionID := getRegionID(r)
	qualifier := getQualifierName(r)
	qualifierOptions := getQualifierOptions(r)
	transform := getTransform(r)

	var data []*wm.ModelOutputQualifierTimeseries
	var err error
	if regionID == "" {
		data, err = a.dataOutput.GetQualifierTimeseries(params, qualifier, qualifierOptions)
	} else {
		data, err = a.dataOutput.GetQualifierTimeseriesByRegion(params, qualifier, qualifierOptions, regionID)
		if transform != "" {
			data, err = a.dataOutput.TransformOutputQualifierTimeseriesByRegion(data, wm.TransformConfig{Transform: transform, RegionID: regionID})
			if err != nil {
				return &wm.Error{Op: op, Err: err}
			}
		}
	}

	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, timeseries := range data {
		list = append(list, &modelOutputQualifierTimeseriesResponse{timeseries})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputQualifierData(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierData"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	qualifiers := getQualifierNames(r)
	data, err := a.dataOutput.GetQualifierData(params, timestamp, qualifiers)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, value := range data {
		list = append(list, &modelOutputQualifierValuesResponse{value})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputQualifierRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierRegional"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	qualifier := getQualifierName(r)
	transform := getTransform(r)
	data, err := a.dataOutput.GetQualifierRegional(params, timestamp, qualifier)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	if transform != "" {
		data, err = a.dataOutput.TransformQualifierRegional(data, timestamp, wm.TransformConfig{Transform: transform})
		if err != nil {
			return &wm.Error{Op: op, Err: err}
		}
	}
	render.JSON(w, r, data)
	return nil
}

func (a *api) getDataOutputPipelineResults(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputPipelineResults"
	params := getPipelineResultParams(r)
	data, err := a.dataOutput.GetPipelineResults(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, data)
	return nil
}
