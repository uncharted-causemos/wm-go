package elastic

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

var datacubeTextSearchFields = []string{
	"model_description",
	"output_description",
	"parameter_descriptions",
}

const datacubesIndex = "datacubes"
const defaultSize = 100

// SearchDatacubes searches and returns datacubes
func (es *ES) SearchDatacubes(search string, filters []*wm.Filter) ([]*wm.Datacube, error) {
	var datacubes []*wm.Datacube
	options := queryOptions{
		filters: filters,
		search:  searchOptions{text: search, fields: datacubeTextSearchFields},
	}
	query, err := buildQuery(options)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"size":  defaultSize,
		"query": query,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	res, err := es.client.Search(
		es.client.Search.WithIndex(datacubesIndex),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithPretty(),
	)
	defer res.Body.Close()
	resBody := read(res.Body)
	if res.IsError() {
		return nil, errors.New(resBody)
	}
	hits := gjson.Get(resBody, "hits.hits").Array()
	for _, hit := range hits {
		doc := hit.Get("_source").String()
		var datacube *wm.Datacube
		if err := json.Unmarshal([]byte(doc), &datacube); err != nil {
			return nil, err
		}
		datacubes = append(datacubes, datacube)
	}
	return datacubes, nil
}
