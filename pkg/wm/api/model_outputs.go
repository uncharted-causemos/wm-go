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
