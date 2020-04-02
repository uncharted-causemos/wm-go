package api

import (
	"net/http"

	"github.com/go-chi/render"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type tilesResponse struct {
	wm.Tiles
}

// Render allows Project to satisfy the render.Renderer interface.
func (fr *tilesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getTiles(w http.ResponseWriter, r *http.Request) {
	filters, err := getFilters(r)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	ts, err := a.maas.GetTiles(filters)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	render.Render(w, r, &tilesResponse{ts})
}
