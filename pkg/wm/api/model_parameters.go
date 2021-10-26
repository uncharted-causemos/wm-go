package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type modelParameterResponse struct {
	*wm.ModelParameter
}

// Render allows to satisfy the render.Renderer interface.
func (mr *modelParameterResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getModelParameters(w http.ResponseWriter, r *http.Request) error {
	params, err := a.maas.GetModelParameters(chi.URLParam(r, paramModelID))
	if err != nil {
		return err
	}
	list := []render.Renderer{}
	for _, p := range params {
		list = append(list, &modelParameterResponse{p})
	}
	render.RenderList(w, r, list)
	return nil
}
