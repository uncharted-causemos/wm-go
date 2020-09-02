package api

import (
	"net/http"

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
