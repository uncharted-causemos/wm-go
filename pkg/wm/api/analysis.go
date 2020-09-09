package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type analysisResponse struct {
	*wm.Analysis
}

func (a *analysisResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) createAnalysis(w http.ResponseWriter, r *http.Request) {
	var analysisObj *wm.Analysis
	err := decodeJSONBody(r, &analysisObj)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	newAnalysis, err := a.data.CreateAnalysis(analysisObj)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &analysisResponse{newAnalysis})
}

func (a *api) updateAnalysis(w http.ResponseWriter, r *http.Request) {
	analysisID := chi.URLParam(r, paramAnalysisID)
	var analysisObj *wm.Analysis
	err := decodeJSONBody(r, &analysisObj)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	updated, err := a.data.UpdateAnalysis(analysisID, analysisObj)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &analysisResponse{updated})
}

func (a *api) updateAnalysisState(w http.ResponseWriter, r *http.Request) {
	analysisID := chi.URLParam(r, paramAnalysisID)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	state := string(body)
	if ok := wm.IsJSON(state); !ok {
		a.errorResponse(w, NewHTTPError(fmt.Errorf("Bad json"), http.StatusBadRequest, "Request body must be a valid json string"), http.StatusInternalServerError)
		return
	}
	updated, err := a.data.UpdateAnalysisState(analysisID, state)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Write([]byte(updated))
}

func (a *api) deleteAnalysis(w http.ResponseWriter, r *http.Request) {
	analysisID := chi.URLParam(r, paramAnalysisID)
	err := a.data.DeleteAnalysis(analysisID)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) getAnalysisByID(w http.ResponseWriter, r *http.Request) {
	analysisID := chi.URLParam(r, paramAnalysisID)
	analysis, err := a.data.GetAnalysisByID(analysisID)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if analysis == nil {
		a.errorResponse(w, NewHTTPError(err, http.StatusNotFound, "Resource not found"), http.StatusNotFound)
		return
	}
	render.Render(w, r, &analysisResponse{analysis})
}

func (a *api) getAnalyses(w http.ResponseWriter, r *http.Request) {
	filters, err := getFilters(r, wm.ContextAnalysis)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}
	analyses, err := a.data.GetAnalyses(filters)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	list := []render.Renderer{}
	for _, analysis := range analyses {
		list = append(list, &analysisResponse{analysis})
	}
	render.RenderList(w, r, list)
}
