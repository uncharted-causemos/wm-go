package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// Bound represents a geo bound
type Bound struct {
	TopLeft     wm.Point `json:"top_left"`
	BottomRight wm.Point `json:"bottom_right"`
}

// GetTile returns the tile.
func (es *ES) GetTile(zoom, x, y uint32, specs wm.TileDataSpecs) (wm.Tile, error) {
	tile := wm.NewTile(zoom, x, y)
	var results []interface{}
	for _, spec := range specs {
		results = append(results, <-es.getRunOutput(Bound(tile.Bound()), zoom, spec))
	}
	for _, r := range results {
		// TODO: get model output data for each of the TileDataSpec,
		// combine them in to geojson features, and add the geo features(and their values) to the tile
		fmt.Printf("Result: \n%v\n", r)
	}
	return tile, nil
}

// get model run output data for given bounds
func (es *ES) getRunOutput(bound Bound, precision uint32, spec wm.TileDataSpec) <-chan interface{} {
	r := make(chan interface{})
	go func() {
		b, _ := json.Marshal(bound)
		data := map[string]interface{}{
			"RunID":     spec.RunID,
			"Feature":   spec.Feature,
			"Bound":     string(b),
			"Precision": 6 + int(precision), // 4096 cells. More details: https://wiki.openstreetmap.org/wiki/Zoom_levels
		}
		bodyTemplate := `{
			"query": {
				"bool": {
					"filter": [
						{
							"term": {
								"run_id": "{{.RunID}}"
							}
						},
						{
							"term": {
								"feature_name": "{{.Feature}}"
							}
						},
						{
							"geo_bounding_box": {
								"geo": {{.Bound}}
							}
						}
					]
				}
			},
			"aggregations": {
				"geotiled": {
					"geotile_grid": {
						"size": 10000,
						"field": "geo",
						"precision": {{.Precision}}
					},
					"aggregations": {
						"spatial_aggregation": {
							"avg": {
								"field": "feature_value"
							}
						}
					}
				}
			}
		}`
		// TODO: handle errors
		buf, _ := format(bodyTemplate, data)
		res, _ := es.client.Search(
			es.client.Search.WithContext(context.Background()),
			es.client.Search.WithIndex(spec.Model),
			es.client.Search.WithBody(buf),
			es.client.Search.WithPretty(),
		)
		r <- res
		close(r)
	}()
	return r
}

func format(text string, data interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := template.Must(template.New("").Parse(text)).Execute(&buf, data); err != nil {
		return nil, err
	}
	return &buf, nil
}
