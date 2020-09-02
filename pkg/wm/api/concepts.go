package api

import (
	"net/http"

	"github.com/go-chi/render"
)

type conceptResponse string

// Render allows to satisfy the render.Renderer interface.
func (cr conceptResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getConcepts(w http.ResponseWriter, r *http.Request) {
	concepts, err := a.maas.GetConcepts()
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	list := []render.Renderer{}
	for _, c := range concepts {
		list = append(list, conceptResponse(c))
	}
	render.RenderList(w, r, list)
}
