package elastic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const (
	modelScenariosIndex = "data-model-scenarios"
	maxNumberOfRuns     = 10000
)

// GetModelRuns returns model runs
func (es *ES) GetModelRuns(modelID string) ([]*wm.ModelRun, error) {
	op := "ES.GetModelRuns"
	rBody := fmt.Sprintf(`{
		"query": {
			"bool": {
				"filter": [
					{ "term":  { "model_id": "%s" }},
					{ "term":  { "output_tile": "READY" }}
				]
			}
		},
		"sort": [
			{
				"created": { "order": "desc" }
			}
		]
	}`, modelID)
	res, err := es.client.Search(
		es.client.Search.WithIndex(modelScenariosIndex),
		es.client.Search.WithSize(maxNumberOfRuns),
		es.client.Search.WithBody(strings.NewReader(rBody)),
	)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, &wm.Error{Op: op, Message: body}
	}
	var runs []*wm.ModelRun
	for _, hit := range gjson.Get(body, "hits.hits").Array() {
		source := hit.Get("_source")
		sourceStr := source.String()
		var run wm.ModelRun
		if err := json.Unmarshal([]byte(sourceStr), &run); err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		run.Model = source.Get("model_id").String()
		runs = append(runs, &run)
	}
	return runs, nil
}
