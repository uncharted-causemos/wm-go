package api

import (
	"net/http"

	"github.com/go-chi/chi"
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

type oldModelOutputTimeseries struct {
	*wm.OldModelOutputTimeseries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *oldModelOutputTimeseries) Render(w http.ResponseWriter, r *http.Request) error {
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

func (a *api) getModelOutputTimeseries(w http.ResponseWriter, r *http.Request) {
	timeseries, err := a.maas.GetOutputTimeseries(chi.URLParam(r, paramRunID), getFeature(r))
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &oldModelOutputTimeseries{timeseries})
}

func (a *api) getRegionalDataOutputStats(w http.ResponseWriter, r *http.Request) {
	params := getDatacubeParams(r)
	stats, err := a.dataOutput.GetRegionalOutputStats(params)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelRegionalOutputStatsResponse{stats})
}

func (a *api) getDataOutputStats(w http.ResponseWriter, r *http.Request) {
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	stats, err := a.dataOutput.GetOutputStats(params, timestamp)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	list := []render.Renderer{}
	for _, stat := range stats {
		list = append(list, &outputStatsResponse{stat})
	}
	render.RenderList(w, r, list)
}

func (a *api) getDataOutputTimeseries(w http.ResponseWriter, r *http.Request) {
	params := getDatacubeParams(r)
	regionID := getRegionID(r)
	var timeseries []*wm.TimeseriesValue
	var err error
	if regionID == "" {
		timeseries, err = a.dataOutput.GetOutputTimeseries(params)
	} else {
		timeseries, err = a.dataOutput.GetOutputTimeseriesByRegion(params, regionID)
	}
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	list := []render.Renderer{}
	for _, point := range timeseries {
		list = append(list, &modelOutputTimeseriesValue{point})
	}
	render.RenderList(w, r, list)
}

func (a *api) getDataOutputRegional(w http.ResponseWriter, r *http.Request) {
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	data, err := a.dataOutput.GetRegionAggregation(params, timestamp)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelOutputRegionalData{data})
}

func (a *api) getDataOutputRaw(w http.ResponseWriter, r *http.Request) {
	params := getDatacubeParams(r)
	data, err := a.dataOutput.GetRawData(params)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	list := []render.Renderer{}
	for _, point := range data {
		list = append(list, &modelOutputRawDataPoint{point})
	}
	render.RenderList(w, r, list)
}

func (a *api) getDataOutputHierarchy(w http.ResponseWriter, r *http.Request) {
	params := getHierarchyParams(r)
	data, err := a.dataOutput.GetRegionHierarchy(params)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, &data)
}
