package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type modelRunsResponse struct {
	*wm.ModelRun
}

// Render allows to satisfy the render.Renderer interface.
func (mr *modelRunsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getModelRuns(w http.ResponseWriter, r *http.Request) error {
	op := "api.getModelRuns"
	runs, err := a.maas.GetModelRuns(chi.URLParam(r, paramModelID))
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, run := range runs {
		list = append(list, &modelRunsResponse{run})
	}
	render.RenderList(w, r, list)
	return nil
}
