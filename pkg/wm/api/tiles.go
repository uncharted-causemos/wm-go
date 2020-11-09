package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

const contentTypeMVT = "application/vnd.mapbox-vector-tile"
const contentEncodingGzip = "gzip"

func (a *api) getTile(w http.ResponseWriter, r *http.Request) {
	specs, err := getTileDataSpecs(r)
	if err != nil {
		a.errorResponse(w, err, http.StatusBadRequest)
		return
	}
	expression := getTileDataExpression(r)
	debug := r.URL.Query().Get("debug")

	var zxy [3]uint32
	for i, key := range []string{paramZoom, paramX, paramY} {
		v, err := strconv.ParseUint(chi.URLParam(r, key), 10, 32)
		if err != nil {
			a.errorResponse(w, err, http.StatusBadRequest)
			return
		}
		zxy[i] = uint32(v)
	}

	tile, err := a.dataOutputTile.GetTile(zxy[0], zxy[1], zxy[2], specs, expression)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	// if debug flag is provided, write the string representation of the tile
	if strings.ToLower(debug) == "true" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(tile.String()))
		return
	}
	result, err := tile.MVT()
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeMVT)
	w.Header().Set("Content-Encoding", contentEncodingGzip)
	w.Write(result)
}
