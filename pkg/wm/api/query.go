package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func getFacetNames(r *http.Request) ([]string, error) {
	raw := r.URL.Query().Get("facets")
	if raw == "" {
		return nil, errors.New("The facets list is missing from the query")
	}

	var facetNames []string
	if err := json.Unmarshal([]byte(raw), &facetNames); err != nil {
		return nil, err
	}

	return facetNames, nil
}

func getSearch(r *http.Request) string {
	return r.URL.Query().Get("search")
}

func getIndicator(r *http.Request) string {
	return r.URL.Query().Get("indicator")
}

func getModel(r *http.Request) string {
	return r.URL.Query().Get("model")
}

func getFeature(r *http.Request) string {
	return r.URL.Query().Get("feature")
}

func getTimestamp(r *http.Request) string {
	return r.URL.Query().Get("timestamp")
}

func getModelId(r *http.Request) string {
	return r.URL.Query().Get("model_id")
}

func getModelOutputParams(r *http.Request) wm.ModelOutputParams {
	// This could be neater with github.com/gorilla/schema but no need for this dependency
	var params wm.ModelOutputParams
	params.ModelId = getModelId(r)
	params.RunId = r.URL.Query().Get("run_id")
	params.Feature = getFeature(r)
	params.Resolution = r.URL.Query().Get("resolution")
	params.TemporalAggFunc = r.URL.Query().Get("temporal_agg")
	params.SpatialAggFunc = r.URL.Query().Get("spatial_agg")
	return params
}

func getUnits(r *http.Request) ([]string, bool) {
	units, ok := r.URL.Query()["unit"]
	return units, ok
}

func getFilters(r *http.Request, context wm.FilterContext) ([]*wm.Filter, error) {
	raw := r.URL.Query().Get("filters")
	if raw == "" {
		return nil, nil
	}

	filters, err := parseFilters([]byte(raw), context)
	if err != nil {
		return nil, err
	}

	return filters, nil
}

func getTileDataSpecs(r *http.Request) (wm.TileDataSpecs, error) {
	raw := r.URL.Query().Get("specs")
	if raw == "" {
		return nil, errors.New("The specs list is missing from the query")
	}

	var specs []wm.TileDataSpec
	if err := json.Unmarshal([]byte(raw), &specs); err != nil {
		return nil, err
	}

	return specs, nil
}

func getTileDataExpression(r *http.Request) string {
	return r.URL.Query().Get("expression")
}
