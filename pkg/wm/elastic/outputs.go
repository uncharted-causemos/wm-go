package elastic

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const (
	modelTimeseriesIndex  = "data-model-timeseries"
	modelOutputStatsIndex = "data-model-output-stats"
)

// GetOutputStats returns model output stats
func (es *ES) GetOutputStats(runID string, feature string) (*wm.ModelOutputStat, error) {
	zoom := 8
	rBody := fmt.Sprintf(`{
		"query": {
			"bool": {
				"filter": [
					{ "term":  { "run_id": "%s" }},
					{ "term":  { "feature_name": "%s" }},
					{ "term":  { "zoom": %d }}
				]
			}
		}
	}`, runID, feature, zoom)
	res, err := es.client.Search(
		es.client.Search.WithIndex(modelOutputStatsIndex),
		es.client.Search.WithSize(10000),
		es.client.Search.WithBody(strings.NewReader(rBody)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}

	hits := gjson.Get(body, "hits.hits").Array()

	if len(hits) == 0 {
		return nil, errors.New("GetOutputStats: No stats found")
	}

	source := hits[0].Get("_source")

	result := &wm.ModelOutputStat{
		Min: source.Get("min_avg").Float(),
		Max: source.Get("max_avg").Float(),
	}

	return result, nil
}

// GetOutputTimeseries returns model output timeseries
func (es *ES) GetOutputTimeseries(runID string, feature string) (*wm.OldModelOutputTimeseries, error) {
	zoom := 8
	rBody := fmt.Sprintf(`{
		"query": {
			"bool": {
				"filter": [
					{ "term":  { "run_id": "%s" }},
					{ "term":  { "feature_name": "%s" }},
					{ "term":  { "zoom": %d }}
				]
			}
		},
		"sort": [
			{
				"timestamp": { "order": "asc" }
			}
		]
	}`, runID, feature, zoom)
	res, err := es.client.Search(
		es.client.Search.WithIndex(modelTimeseriesIndex),
		es.client.Search.WithSize(10000),
		es.client.Search.WithBody(strings.NewReader(rBody)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}

	series := []wm.TimeseriesValue{}
	for _, hit := range gjson.Get(body, "hits.hits").Array() {
		source := hit.Get("_source")
		val := wm.TimeseriesValue{
			Timestamp: source.Get("timestamp").Int(),
			Value:     source.Get("avg_bin_avg").Float(),
		}
		series = append(series, val)
	}
	return &wm.OldModelOutputTimeseries{Timeseries: series}, nil
}
