package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

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

func getFeature(r *http.Request) string {
	return r.URL.Query().Get("feature")
}

func getTimestamp(r *http.Request) string {
	return r.URL.Query().Get("timestamp")
}

func getTimestampsFromBody(r *http.Request) (wm.Timestamps, error) {
	var tss wm.Timestamps

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return wm.Timestamps{}, err
	}

	err = json.Unmarshal(body, &tss)
	if err != nil {
		return wm.Timestamps{}, err
	}

	return tss, nil
}

func getAggForSelect(r *http.Request) string {
	return r.URL.Query().Get("aggForSelect")
}

func getAggForAll(r *http.Request) string {
	return r.URL.Query().Get("aggForAll")
}

func getTransform(r *http.Request) wm.Transform {
	transform := r.URL.Query().Get("transform")
	return wm.Transform(transform)
}

func getAdminLevel(r *http.Request) wm.AdminLevel {
	adminLevel := r.URL.Query().Get("admin_level")
	return wm.AdminLevel(adminLevel)
}

func getAgg(r *http.Request) string {
	return r.URL.Query().Get("agg")
}

func getRegionID(r *http.Request) string {
	return r.URL.Query().Get("region_id")
}

func getRawDataResolution(r *http.Request) wm.TemporalResolution {
	return wm.TemporalResolution(r.URL.Query().Get("raw_res"))
}

func getRawDataLatestTimestamp(r *http.Request) (int64, error) {
	val := r.URL.Query().Get("raw_latest_ts")
	if val == "" {
		return 0, nil
	}
	ts, err := strconv.Atoi(val)
	if err != nil {
		return 0, &wm.Error{Code: wm.EINVALID, Message: "Invalid 'raw_latest_ts' parameter value"}
	}
	return int64(ts), nil
}

type regionIDsBody struct {
	RegionIDs []string `json:"region_ids"`
}

func getRegionIDsFromBody(r *http.Request) ([]string, error) {
	var regionIDs regionIDsBody

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &regionIDs)
	if err != nil {
		return nil, err
	}

	return regionIDs.RegionIDs, nil
}

type timeseriesParamsBody struct {
	TimeseriesParams []*wm.FullTimeseriesParams `json:"timeseries_params"`
}

func getTimeseriesParamsFromBody(r *http.Request) ([]*wm.FullTimeseriesParams, error) {
	var params timeseriesParamsBody

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		return nil, err
	}

	return params.TimeseriesParams, nil
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
	params.Resolution = wm.TemporalResolutionOption(r.URL.Query().Get("resolution"))
	params.TemporalAggFunc = wm.AggregationOption(r.URL.Query().Get("temporal_agg"))
	params.SpatialAggFunc = wm.AggregationOption(r.URL.Query().Get("spatial_agg"))
	params.AdminLevel = getAdminLevel(r)
	return params
}

func getRegionListsParams(r *http.Request) wm.RegionListParams {
	var params wm.RegionListParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunIDs = r.URL.Query()["run_ids[]"]
	params.Feature = getFeature(r)
	return params
}

func getQualifierInfoParams(r *http.Request) wm.QualifierInfoParams {
	var params wm.QualifierInfoParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunID = r.URL.Query().Get("run_id")
	params.Feature = getFeature(r)
	return params
}

func getPipelineResultParams(r *http.Request) wm.PipelineResultsParams {
	var params wm.PipelineResultsParams
	params.DataID = r.URL.Query().Get("data_id")
	params.RunID = r.URL.Query().Get("run_id")
	return params
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
