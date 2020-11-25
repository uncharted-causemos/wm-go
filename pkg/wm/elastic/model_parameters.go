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

const (
	modelParametersIndex = "model_parameters"
)

// GetModelParameters returns model runs
func (es *ES) GetModelParameters(model string) ([]*wm.ModelParameter, error) {
	// If model is numeric we are dealing with supermaas model id (new data)
	if _, err := strconv.Atoi(model); err == nil {
		return es.getModelParameters(model)
	}
	// For legacy data  (old maas)
	return es.modelService.GetModelParameters(model)
}

func (es *ES) getModelParameters(modelID string) ([]*wm.ModelParameter, error) {
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
		return nil, err
	}
	body := read(res.Body)
	if res.IsError() {
		return nil, errors.New(body)
	}
	params := make([]*wm.ModelParameter, 0)
	for _, hit := range gjson.Get(body, "hits.hits").Array() {
		source := hit.Get("_source").String()
		var param wm.ModelParameter
		if err := json.Unmarshal([]byte(source), &param); err != nil {
			return nil, err
		}
		params = append(params, &param)
	}
	return params, nil
}
