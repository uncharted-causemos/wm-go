package api

import (
	"encoding/json"
	"net/http"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func getFacetNames(r *http.Request) ([]string, error) {
	raw := r.URL.Query().Get("facets")
	if raw == "" {
		return nil, &wm.Error{Code: wm.EINVALID, Message: "The facets list is missing from the query"}
	}

	var facetNames []string
	if err := json.Unmarshal([]byte(raw), &facetNames); err != nil {
		return nil, &wm.Error{Code: wm.EINVALID, Message: "Invalid facets list"}
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

func getTransform(r *http.Request) wm.Transform {
	transform := r.URL.Query().Get("transform")
	return wm.Transform(transform)
}

func getRegionID(r *http.Request) string {
	return r.URL.Query().Get("region_id")
}

func getQualifierName(r *http.Request) string {
	return r.URL.Query().Get("qualifier")
}

func getQualifierNames(r *http.Request) []string {
	// Using these silly short names since it will be repeated in the URL for every element
	// and we don't want to hit the URL length limit
	return r.URL.Query()["qlf[]"]
}

func getQualifierOptions(r *http.Request) []string {
	// Using these silly short names since it will be repeated in the URL for every element
	// and we don't want to hit the URL length limit
	return r.URL.Query()["q_opt[]"]
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

func getRegionListsParams(r *http.Request) wm.RegionListParams {
	var params wm.RegionListParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunIDs = r.URL.Query()["run_ids[]"]
	params.Feature = getFeature(r)
	return params
}

func getPipelineResultParams(r *http.Request) wm.PipelineResultsParams {
	var params wm.PipelineResultsParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunID = r.URL.Query().Get("run_id")
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
		return nil, &wm.Error{Code: wm.EINVALID, Message: "Invalid 'filters' parameter value"}
	}

	return filters, nil
}

func getGridTileOutputSpecs(r *http.Request) (wm.GridTileOutputSpecs, error) {
	raw := r.URL.Query().Get("specs")
	if raw == "" {
		return nil, &wm.Error{Code: wm.EINVALID, Message: "The 'specs' list is missing from the query"}
	}

	var specs []wm.GridTileOutputSpec
	if err := json.Unmarshal([]byte(raw), &specs); err != nil {
		return nil, &wm.Error{Code: wm.EINVALID, Message: "Invalid 'specs' parameter value"}
	}

	return specs, nil
}

func getTileDataExpression(r *http.Request) string {
	return r.URL.Query().Get("expression")
}
