package elastic

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const datacubesIndex = "data-datacubes"
const indicatorDataIndex = "data-indicators"
const defaultSize = 100

// SearchDatacubes searches and returns datacubes
func (es *ES) SearchDatacubes(filters []*wm.Filter) ([]*wm.Datacube, error) {
	var datacubes []*wm.Datacube
	options := queryOptions{
		filters: filters,
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
		datacube.SearchScore = hit.Get("_score").Float()
		datacubes = append(datacubes, datacube)
	}
	return datacubes, nil
}

// CountDatacubes returns data cube count
func (es *ES) CountDatacubes(filters []*wm.Filter) (uint64, error) {
	options := queryOptions{
		filters: filters,
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

// GetIndicatorData returns the indicator time series
func (es *ES) GetIndicatorData(indicatorName string, modelName string, units []string) ([]*wm.IndicatorDataPoint, error) {
	options := queryOptions{
		filters: []*wm.Filter{
			{
				Field:        wm.FieldIndicatorVariable,
				StringValues: []string{indicatorName},
			},
			{
				Field:        wm.FieldIndicatorDataset,
				StringValues: []string{modelName},
			},
		},
	}
	if len(units) > 0 {
		options.filters = append(options.filters, &wm.Filter{
			Field:        wm.FieldIndicatorUnit,
			StringValues: units,
		})
	}
	query, err := buildQuery(options)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{}
	if len(query) > 0 {
		body["query"] = query
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	res, err := es.client.Search(
		es.client.Search.WithIndex(indicatorDataIndex),
		es.client.Search.WithSize(10000),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithPretty(),
	)
	defer res.Body.Close()
	resBody := read(res.Body)
	if res.IsError() {
		return nil, errors.New(resBody)
	}

	hits := gjson.Get(resBody, "hits.hits").Array()

	var dataPoints []*wm.IndicatorDataPoint
	for _, hit := range hits {
		doc := hit.Get("_source").String()
		var dataPoint *wm.IndicatorDataPoint
		if err := json.Unmarshal([]byte(doc), &dataPoint); err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, dataPoint)
	}

	return dataPoints, nil
}
