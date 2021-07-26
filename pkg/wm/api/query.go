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

func getRegionID(r *http.Request) string {
	return r.URL.Query().Get("region_id")
}

func getDatacubeParams(r *http.Request) wm.DatacubeParams {
	// This could be neater with github.com/gorilla/schema but no need for this dependency
	var params wm.DatacubeParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunID = r.URL.Query().Get("run_id")
	params.Feature = getFeature(r)
	params.Resolution = r.URL.Query().Get("resolution")
	params.TemporalAggFunc = r.URL.Query().Get("temporal_agg")
	params.SpatialAggFunc = r.URL.Query().Get("spatial_agg")
	return params
}

func getHierarchyParams(r *http.Request) wm.HierarchyParams {
	var params wm.HierarchyParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunID = r.URL.Query().Get("run_id")
	params.Feature = getFeature(r)
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

func getGridTileOutputSpecs(r *http.Request) (wm.GridTileOutputSpecs, error) {
	raw := r.URL.Query().Get("specs")
	if raw == "" {
		return nil, errors.New("The specs list is missing from the query")
	}

	var specs []wm.GridTileOutputSpec
	if err := json.Unmarshal([]byte(raw), &specs); err != nil {
		return nil, err
	}

	return specs, nil
}

func getTileDataExpression(r *http.Request) string {
	return r.URL.Query().Get("expression")
}
