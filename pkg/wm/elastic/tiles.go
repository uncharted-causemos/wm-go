package elastic

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// bound represents a geo bound
type bound struct {
	TopLeft     wm.Point `json:"top_left"`
	BottomRight wm.Point `json:"bottom_right"`
}

// geoTile is a single record of the ES geotile bucket aggregation result
type geoTile struct {
	Key                string `json:"key"`
	DocCount           int    `json:"doc_count"`
	SpatialAggregation struct {
		Value float64 `json:"value"`
	} `json:"spatial_aggregation"`
}

// geoTiles is the ES geotile bucket aggregation result
type geoTiles []geoTile

// GetTile returns the tile.
func (es *ES) GetTile(zoom, x, y uint32, specs wm.TileDataSpecs) (wm.Tile, error) {
	tile := wm.NewTile(zoom, x, y)
	var results []geoTiles
	for _, spec := range specs {
		out, _ := es.getRunOutput(bound(tile.Bound()), zoom, spec)
		results = append(results, <-out)
	}
	for _, r := range results {
		// TODO: get model output data for each of the TileDataSpec,
		// combine them in to geojson features, and add the geo features(and their values) to the tile
		fmt.Printf("Result: \n%v\n", r)
	}
	return tile, nil
}

// getRunOutput returns geotiled bucket aggregation result of the model run output specified by the spec, bound and zoom
func (es *ES) getRunOutput(bound bound, precision uint32, spec wm.TileDataSpec) (<-chan geoTiles, <-chan error) {
	out := make(chan geoTiles)
	er := make(chan error)
	go func() {
		defer close(out)
		b, _ := json.Marshal(bound)
		data := map[string]interface{}{
			"RunID":     spec.RunID,
			"Feature":   spec.Feature,
			"Bound":     string(b),
			"Precision": 1, // 4096 cells. More details: https://wiki.openstreetmap.org/wiki/Zoom_levels
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
		buf, err := format(bodyTemplate, data)
		if err != nil {
			er <- err
			return
		}
		res, err := es.client.Search(
			es.client.Search.WithIndex(spec.Model),
			es.client.Search.WithBody(buf),
		)
		if err != nil {
			er <- err
			return
		}
		buckets := gjson.Get(read(res.Body), "aggregations.geotiled.buckets").String()
		var result geoTiles
		if err := json.Unmarshal([]byte(buckets), &result); err != nil {
			er <- err
			return
		}
		out <- result
	}()
	return out, er
}
