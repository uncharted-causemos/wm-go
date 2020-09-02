package api

import (
	"net/http"

	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type datacubesResponse struct {
	*wm.Datacube
}

type countDatacubesResponse uint64

// Render allows to satisfy the render.Renderer interface.
func (d *datacubesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render satisfies the render.Renderer interface.
func (cd countDatacubesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getDatacubes(w http.ResponseWriter, r *http.Request) {
	filters, err := getFilters(r, wm.ContextDatacube)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}
	datacubes, err := a.maas.SearchDatacubes(getSearch(r), filters)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	list := []render.Renderer{}
	for _, datacube := range datacubes {
		list = append(list, &datacubesResponse{datacube})
	}
	render.RenderList(w, r, list)
}

func (a *api) countDatacubes(w http.ResponseWriter, r *http.Request) {
	filters, err := getFilters(r, wm.ContextDatacube)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}
	count, err := a.maas.CountDatacubes(getSearch(r), filters)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, countDatacubesResponse(count))
}
