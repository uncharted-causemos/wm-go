package elastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// This index will be deprecated and use of it is only temporary
const parametersIndex = "parameters"

const modelTimeseriesIndex = "model-timeseries"

const modelRunIndex = "model-run-parameters"
const maxNumberOfRuns = 10000

// HACK: This function gets the run parameters data for given model including the ones that we don't have in our existing `model-run-parameters` index.
// This relies on old `parameters` index that has been created from ingesting parameters data from Jataware's postgres db.
func (es *ES) getModelRuns(model string) ([]*wm.ModelRun, error) {
	reqBody := fmt.Sprintf(`
		{
			"size": 0,
			"query": {
				"bool": {
					"filter": [
						{"term": {"model": "%s" }}
					]
				}
			},
			"aggs": {
				"by_run": {
					"terms": { "field": "run_id", "size": 1000 },
					"aggs": {
						"by_param": {
							"terms": {"field": "parameter_name", "size": 100 },
							"aggs": {
								"doc": {
										"top_hits": { "size": 1 }
								}
							}
						}
					}
				}
			}
		}
	`, model)

	res, err := es.client.Search(
		es.client.Search.WithIndex(parametersIndex),
		es.client.Search.WithBody(strings.NewReader(reqBody)),
	)
	if err != nil {
		return nil, err
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}

	var runs []*wm.ModelRun

	for _, run := range gjson.Get(body, "aggregations.by_run.buckets").Array() {
		parameters := make([]wm.ModelRunParameter, 0)
		for _, param := range run.Get("by_param.buckets").Array() {
			source := param.Get("doc.hits.hits.0._source")
			pName := source.Get("parameter_name").String()
			if pName != "" {
				parameters = append(parameters, wm.ModelRunParameter{
					Name:  pName,
					Value: source.Get("parameter_value").String(),
				})
			}
		}
		run := &wm.ModelRun{
			ID:         run.Get("key").String(),
			Model:      model,
			Parameters: parameters,
		}
		runs = append(runs, run)
	}

	return runs, nil
}

// Returns a map of runIds that has it's output processed and ingested in our system
func (es *ES) getAvailableRunIDMap(model string) (map[string]bool, error) {
	reqBody := fmt.Sprintf(`
		{
			"size": 0,
			"query": {
					"bool": {
							"filter": [
									{"term": {"model": "%s"} }
							]
					}
			},
			"aggs": {
					"by_run": {
							"terms": {"field": "run_id", "size": 1000 } 
					}
			}
	}
	`, strings.ToLower(model))

	res, err := es.client.Search(
		es.client.Search.WithIndex(modelTimeseriesIndex),
		es.client.Search.WithBody(strings.NewReader(reqBody)),
	)
	if err != nil {
		return nil, err
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}

	runIDs := make(map[string]bool)

	for _, run := range gjson.Get(body, "aggregations.by_run.buckets").Array() {
		runID := run.Get("key").String()
		runIDs[runID] = true
	}

	return runIDs, nil
}

func (es *ES) getScenarios(modelID string) ([]*wm.ModelRun, error) {
	index := "model_scenarios"
	rBody := fmt.Sprintf(`{
		"query": { 
			"bool": { 
				"filter": [ 
					{ "term":  { "model": "%s" }}
				]
			}
		}
	}`, modelID)
	res, err := es.client.Search(
		es.client.Search.WithIndex(index),
		es.client.Search.WithSize(maxNumberOfRuns),
		es.client.Search.WithBody(strings.NewReader(rBody)),
	)
	if err != nil {
		return nil, err
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}
	var runs []*wm.ModelRun
	for _, hit := range gjson.Get(body, "hits.hits").Array() {
		source := hit.Get("_source").String()
		var run wm.ModelRun
		if err := json.Unmarshal([]byte(source), &run); err != nil {
			return nil, err
		}
		runs = append(runs, &run)
	}
	return runs, nil
}

// GetModelRuns returns model runs
func (es *ES) GetModelRuns(model string) ([]*wm.ModelRun, error) {

	// If model is numeric we are dealing with supermaas model id (new data)
	if _, err := strconv.Atoi(model); err == nil {
		return es.getScenarios(model)
	}

	data := map[string]interface{}{
		"Model": model,
	}
	// TODO: Migrate this with new model run schema when ready
	bodyTemplate := `{
		"query": {
			"bool": {
				"filter": [
					{
						"term": {
							"model": "{{.Model}}"
						}
					}
				]
			}
		},
		"collapse": {
			"field": "run_id",
			"inner_hits": {
				"name": "parameters_by_run",
				"size": 100,
				"sort": [
					{ "parameter_name": "asc" }
				]
			}
		}
	}`
	buf, err := format(bodyTemplate, data)
	if err != nil {
		return nil, err
	}
	res, err := es.client.Search(
		es.client.Search.WithIndex(modelRunIndex),
		es.client.Search.WithSize(maxNumberOfRuns),
		es.client.Search.WithBody(buf),
	)
	if err != nil {
		return nil, err
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}

	var runs []*wm.ModelRun
	for _, hit := range gjson.Get(body, "hits.hits").Array() {
		parameters := make([]wm.ModelRunParameter, 0)
		for _, paramHit := range hit.Get("inner_hits.parameters_by_run.hits.hits").Array() {
			pName := paramHit.Get("_source.parameter_name").String()
			if pName != "" {
				parameters = append(parameters, wm.ModelRunParameter{
					Name:  pName,
					Value: paramHit.Get("_source.parameter_value").String(),
				})
			}
		}
		run := &wm.ModelRun{
			ID:         hit.Get("_source.run_id").String(),
			Model:      model,
			Parameters: parameters,
		}
		runs = append(runs, run)
	}

	//Hack: If runs for given model doesn't exist, try searching from the old 'parameters' index
	if len(runs) == 0 {
		r, err := es.getModelRuns(model)
		if err != nil {
			return nil, err
		}
		runs = r
	}

	// Filter out the runs that we could not process or do not have tiles for.
	rIDs, err := es.getAvailableRunIDMap(model)
	if err != nil {
		return nil, err
	}
	var results []*wm.ModelRun
	for _, run := range runs {
		if rIDs[run.ID] {
			results = append(results, run)
		}
	}

	return results, nil
}
