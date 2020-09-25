package api

import (
	"net/http"
	"strconv"

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

	var zxy [3]uint32
	for i, key := range []string{paramZoom, paramX, paramY} {
		v, err := strconv.ParseUint(chi.URLParam(r, key), 10, 32)
		if err != nil {
			a.errorResponse(w, err, http.StatusBadRequest)
			return
		}
		zxy[i] = uint32(v)
	}

	tile, err := a.maasStorage.GetTile(zxy[0], zxy[1], zxy[2], specs)
	if err != nil {
		a.errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeMVT)
	w.Header().Set("Content-Encoding", contentEncodingGzip)
	w.Write(tile)
}
