package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type tilesResponse struct {
	Tile string
}

// Render allows Project to satisfy the render.Renderer interface.
func (tr *tilesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getTile(w http.ResponseWriter, r *http.Request) {
	specs, err := getTileDataSpecs(r)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	var zxy [3]uint32
	for i, key := range []string{"zoom", "x", "y"} {
		v, err := strconv.ParseUint(chi.URLParam(r, key), 10, 32)
		if err != nil {
			a.errorResponse(w, err, http.StatusBadRequest)
			return
		}
		zxy[i] = uint32(v)
	}

	tile, err := a.maas.GetTile(zxy[0], zxy[1], zxy[2], specs)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	res, err := tile.MVT()
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	render.Render(w, r, &tilesResponse{Tile: res})
}
