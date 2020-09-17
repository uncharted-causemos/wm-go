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

const datacubesIndex = "datacubes-supermaas-2020-09-15"
const defaultSize = 100

// SearchDatacubes searches and returns datacubes
func (es *ES) SearchDatacubes(search string, filters []*wm.Filter) ([]*wm.Datacube, error) {
	var datacubes []*wm.Datacube
	options := queryOptions{
		filters: filters,
		search:  searchOptions{text: search, fields: datacubeTextSearchFields},
	}
	query, err := buildBoolQuery(options)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"size": defaultSize,
	}
	if len(query) > 0 {
		body["query"] = query
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

// CountDatacubes returns data cube count
func (es *ES) CountDatacubes(search string, filters []*wm.Filter) (uint64, error) {
	options := queryOptions{
		filters: filters,
		search:  searchOptions{text: search, fields: datacubeTextSearchFields},
	}
	query, err := buildQuery(options)
	if err != nil {
		return 0, err
	}
	body := map[string]interface{}{}
	if len(query) > 0 {
		body["query"] = query
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return 0, err
	}
	res, err := es.client.Count(
		es.client.Count.WithIndex(datacubesIndex),
		es.client.Count.WithBody(&buf),
		es.client.Count.WithPretty(),
	)
	defer res.Body.Close()
	resBody := read(res.Body)
	if res.IsError() {
		return 0, errors.New(resBody)
	}
	count := gjson.Get(resBody, "count").Uint()
	return count, nil
}
