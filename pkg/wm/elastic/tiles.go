package elastic

import (
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
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

// geoTilesResult is the ES geotile bucket aggregation result
type geoTilesResult struct {
	Bound     bound
	Precision int
	Spec      wm.TileDataSpec
	Data      []geoTile
}

// GetTile returns the tile.
func (es *ES) GetTile(zoom, x, y uint32, specs wm.TileDataSpecs) (wm.Tile, error) {
	tile := wm.NewTile(zoom, x, y)
	var results []geoTilesResult
	for _, spec := range specs {
		// + 6 precision results 4096 cells in the bound. More details: https://wiki.openstreetmap.org/wiki/Zoom_levels
		// TODO: fix precision
		out, _ := es.getRunOutput(bound(tile.Bound()), zoom+1, spec)
		results = append(results, <-out)
	}
	featureMap := map[string]geojson.Feature{}
	for _, result := range results {
		for _, gt := range result.Data {
			if _, ok := featureMap[gt.Key]; !ok {
				var z, x, y uint32
				fmt.Sscanf(gt.Key, "%d/%d/%d", &z, &x, &y)
				polygon := maptile.New(x, y, maptile.Zoom(z)).Bound().ToPolygon()
				featureMap[gt.Key] = geojson.Feature{
					Type:     "Feature",
					Geometry: polygon,
					Properties: geojson.Properties{
						"id": gt.Key,
					},
				}
			}
			featureMap[gt.Key].Properties[result.Spec.ValueProp] = gt.SpatialAggregation.Value
		}
	}
	for _, feature := range featureMap {
		tile.AddFeature(feature)
	}
	return tile, nil
}

// getRunOutput returns geotiled bucket aggregation result of the model run output specified by the spec, bound and zoom
func (es *ES) getRunOutput(bound bound, precision uint32, spec wm.TileDataSpec) (<-chan geoTilesResult, <-chan error) {
	out := make(chan geoTilesResult)
	er := make(chan error)
	go func() {
		defer close(out)
		b, _ := json.Marshal(bound)
		data := map[string]interface{}{
			"RunID":     spec.RunID,
			"Feature":   spec.Feature,
			"Bound":     string(b),
			"Precision": precision,
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
		result := geoTilesResult{
			Bound:     bound,
			Precision: int(precision),
			Spec:      spec,
			Data:      []geoTile{},
		}
		if err := json.Unmarshal([]byte(buckets), &result.Data); err != nil {
			er <- err
			return
		}
		out <- result
	}()
	return out, er
}
