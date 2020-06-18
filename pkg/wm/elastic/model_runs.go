package elastic

import (
	"errors"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const modelRunIndex = "model-run-parameters"

// GetModelRuns returns model runs
func (es *ES) GetModelRuns(model string) ([]*wm.ModelRun, error) {
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
		var parameters []wm.ModelRunParameter
		for _, paramHit := range hit.Get("inner_hits.parameters_by_run.hits.hits").Array() {
			parameters = append(parameters, wm.ModelRunParameter{
				Name:  paramHit.Get("_source.parameter_name").String(),
				Value: paramHit.Get("_source.parameter_value").String(),
			})
		}
		run := &wm.ModelRun{
			ID:        hit.Get("_source.run_id").String(),
			Model:     model,
			Parameter: parameters,
		}
		runs = append(runs, run)
	}
	return runs, nil
}
