package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type modelOutputStatsResponse struct {
	*wm.ModelOutputStat
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputStatsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputTimeseries struct {
	*wm.ModelOutputTimeseries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputTimeseries) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputRegionalData struct {
	*wm.ModelOutputRegionalAdmins
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputRegionalData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getModelOutputStats(w http.ResponseWriter, r *http.Request) {
	stats, err := a.maas.GetOutputStats(chi.URLParam(r, paramRunID), getFeature(r))
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelOutputStatsResponse{stats})
}

func (a *api) getModelOutputTimeseries(w http.ResponseWriter, r *http.Request) {
	timeseries, err := a.maas.GetOutputTimeseries(chi.URLParam(r, paramRunID), getFeature(r))
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelOutputTimeseries{timeseries})
}




func (a *api) getDataOutputStats(w http.ResponseWriter, r *http.Request) {
	params := getModelOutputParams(r)
	stats, err := a.dataOutput.GetOutputStats(params)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelOutputStatsResponse{stats})
}

func (a *api) getDataOutputTimeseries(w http.ResponseWriter, r *http.Request) {
	params := getModelOutputParams(r)
	timeseries, err := a.dataOutput.GetOutputTimeseries(params)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelOutputTimeseries{timeseries})
}

func (a *api) getDataOutputRegional(w http.ResponseWriter, r *http.Request) {
	params := getModelOutputParams(r)
	timestamp := getTimestamp(r)
	data, err := a.dataOutput.GetRegionAggregation(params, timestamp)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &modelOutputRegionalData{data})
}

func (a *api) getModelSummary(w http.ResponseWriter, r *http.Request) {
	modelID := getModelID(r)
	feature := getFeature(r)
	summary, err := a.dataOutput.GetModelSummary(modelID, feature)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, summary)
}
