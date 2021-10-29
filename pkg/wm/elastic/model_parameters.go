package elastic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const (
	modelParametersIndex = "data-model-parameters"
)

// GetModelParameters returns model parameters
func (es *ES) GetModelParameters(modelID string) ([]*wm.ModelParameter, error) {
	op := "ES.GetModelParameters"
	rBody := fmt.Sprintf(`{
		"query": { 
			"bool": { 
				"filter": [ 
					{ "term":  { "model_id": "%s" }}
				]
			}
		}
	}`, modelID)
	res, err := es.client.Search(
		es.client.Search.WithIndex(modelParametersIndex),
		es.client.Search.WithSize(100),
		es.client.Search.WithBody(strings.NewReader(rBody)),
	)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, &wm.Error{Op: op, Message: body}
	}
	params := make([]*wm.ModelParameter, 0)
	for _, hit := range gjson.Get(body, "hits.hits").Array() {
		source := hit.Get("_source").String()
		var param wm.ModelParameter
		if err := json.Unmarshal([]byte(source), &param); err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		params = append(params, &param)
	}
	return params, nil
}
