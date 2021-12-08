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

func (a *api) getDataOutputRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRegional"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	transform := getTransform(r)
	data, err := a.dataOutput.GetRegionAggregation(params, timestamp)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	if transform != "" {
		data, err = a.dataOutput.TransformRegionAggregation(data, timestamp, wm.TransformConfig{Transform: transform})
		if err != nil {
			return &wm.Error{Op: op, Err: err}
		}
	}
	render.Render(w, r, &modelOutputRegionalData{data})
	return nil
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
