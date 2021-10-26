package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const contentTypeMVT = "application/vnd.mapbox-vector-tile"
const contentEncodingGzip = "gzip"

func (a *api) getVectorTile(w http.ResponseWriter, r *http.Request) error {
	var zxy [3]uint32
	for i, key := range []string{paramZoom, paramX, paramY} {
		v, err := strconv.ParseUint(chi.URLParam(r, key), 10, 32)
		if err != nil {
			return &wm.Error{Code: wm.EINVALID, Message: "Invalid tile coordinates"}
		}
		zxy[i] = uint32(v)
	}
	debug := r.URL.Query().Get("debug")
	tileSet := chi.URLParam(r, paramTileSetName)
	tile, err := a.vectorTile.GetVectorTile(zxy[0], zxy[1], zxy[2], tileSet)
	if err != nil {
		return nil
	}
	// if debug flag is provided, write the string representation of the tile
	if strings.ToLower(debug) == "true" {
		tileJSON, err := wm.MvtToJSON(tile)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(tileJSON)
		return nil
	}

	w.Header().Set("Content-Type", contentTypeMVT)
	w.Header().Set("Content-Encoding", contentEncodingGzip)
	w.Write(tile)
	return nil
}

func (a *api) getTile(w http.ResponseWriter, r *http.Request) error {
	specs, err := getGridTileOutputSpecs(r)
	if err != nil {
		return &wm.Error{Code: wm.EINVALID, Message: "Invalid tile specs"}
	}
	expression := getTileDataExpression(r)
	debug := r.URL.Query().Get("debug")

	var zxy [3]uint32
	for i, key := range []string{paramZoom, paramX, paramY} {
		v, err := strconv.ParseUint(chi.URLParam(r, key), 10, 32)
		if err != nil {
			return &wm.Error{Code: wm.EINVALID, Message: "Invalid tile coordinates"}
		}
		zxy[i] = uint32(v)
	}

	tile, err := a.dataOutput.GetTile(zxy[0], zxy[1], zxy[2], specs, expression)
	if err != nil {
		return nil
	}
	result, err := tile.MVT()
	if err != nil {
		return nil
	}
	// if debug flag is provided, write the string representation of the tile
	if strings.ToLower(debug) == "true" {
		tileJSON, err := wm.MvtToJSON(result)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(tileJSON)
		return nil
	}
	w.Header().Set("Content-Type", contentTypeMVT)
	w.Header().Set("Content-Encoding", contentEncodingGzip)
	w.Write(result)
	return nil
}
