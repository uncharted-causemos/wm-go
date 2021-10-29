package api

import (
	"net/http"

	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type conceptResponse string

// Render allows to satisfy the render.Renderer interface.
func (cr conceptResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getConcepts(w http.ResponseWriter, r *http.Request) error {
	op := "api.getConcepts"
	concepts, err := a.maas.GetConcepts()
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, c := range concepts {
		list = append(list, conceptResponse(c))
	}
	render.RenderList(w, r, list)
	return nil
}
