package elastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/tidwall/gjson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

const tileDataLayerName = "maas"

// bound represents a geo bound
type bound struct {
	TopLeft     wm.Point `json:"top_left"`
	BottomRight wm.Point `json:"bottom_right"`
}

// geoTile is a single record of the ES geotile bucket aggregation result
type geoTile struct {
	Key                string `json:"key"`
	SpatialAggregation struct {
		Value float64 `json:"value"`
	} `json:"spatial_aggregation"`
}

// geoTilesResult is the ES geotile bucket aggregation result
type geoTilesResult struct {
	bound     bound
	precision int
	spec      wm.TileDataSpec
	data      []geoTile
}

// GetTile returns the tile containing model run output specified by the spec
func (es *ES) GetTile(zoom, x, y uint32, specs wm.TileDataSpecs) ([]byte, error) {
	tile := wm.NewTile(zoom, x, y, tileDataLayerName)
	precision := zoom + 6 // + 6 precision results 4096 cells in the bound. More details: https://wiki.openstreetmap.org/wiki/Zoom_levels

	var errChs []chan error
	var resChs []chan geoTilesResult
	var results []geoTilesResult

	for _, spec := range specs {
		res, err := es.getRunOutput(bound(tile.Bound()), precision, spec)
		errChs = append(errChs, err)
		resChs = append(resChs, res)
	}
	for _, err := range errChs {
		if e := <-err; e != nil {
			return nil, e
		}
	}
	for _, r := range resChs {
		results = append(results, <-r)
	}

	featureMap, err := es.createFeatures(results)
	if err != nil {
		return nil, err
	}
	for _, feature := range featureMap {
		tile.AddFeature(feature)
	}
	return tile.MVT()
}

// getRunOutput returns geotiled bucket aggregation result of the model run output specified by the spec, bound and zoom
func (es *ES) getRunOutput(bound bound, precision uint32, spec wm.TileDataSpec) (chan geoTilesResult, chan error) {
	out := make(chan geoTilesResult)
	er := make(chan error)
	go func() {
		defer close(er)
		defer close(out)
		startTime, err := time.Parse(time.RFC3339, spec.Date)
		if err != nil {
			er <- err
			return
		}
		b, _ := json.Marshal(bound)
		data := map[string]interface{}{
			"RunID":       spec.RunID,
			"Feature":     spec.Feature,
			"Bound":       string(b),
			"Precision":   precision,
			"StartTime":   startTime.Format(time.RFC3339),
			"EndTime":     startTime.AddDate(0, 1, 0).Format(time.RFC3339), // Add a month since we want monthly data
			"Aggregation": "avg",
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
						},
						{
							"range": {
								"timestamp": {
									"gte": "{{.StartTime}}",
									"lte": "{{.EndTime}}"
								}
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
							"{{.Aggregation}}": {
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
			es.client.Search.WithIndex(strings.ToLower(spec.Model)),
			es.client.Search.WithBody(buf),
		)
		if err != nil {
			er <- err
			return
		}
		body := read(res.Body)
		if res.IsError() {
			er <- errors.New(body)
			return
		}

		buckets := gjson.Get(body, "aggregations.geotiled.buckets").String()
		result := geoTilesResult{
			bound:     bound,
			precision: int(precision),
			spec:      spec,
			data:      []geoTile{},
		}
		if err := json.Unmarshal([]byte(buckets), &result.data); err != nil {
			er <- err
			return
		}
		er <- nil
		out <- result
	}()
	return out, er
}

// createFeatures processes and merges the results and returns a map of geojson feature
func (es *ES) createFeatures(results []geoTilesResult) (map[string]geojson.Feature, error) {
	featureMap := map[string]geojson.Feature{}
	for _, result := range results {
		for _, gt := range result.data {
			if _, ok := featureMap[gt.Key]; !ok {
				var z, x, y uint32
				if _, err := fmt.Sscanf(gt.Key, "%d/%d/%d", &z, &x, &y); err != nil {
					return nil, err
				}
				polygon := maptile.New(x, y, maptile.Zoom(z)).Bound().ToPolygon()
				f := *geojson.NewFeature(polygon)
				f.Properties["id"] = gt.Key
				featureMap[gt.Key] = f
			}
			featureMap[gt.Key].Properties[result.spec.ValueProp] = gt.SpatialAggregation.Value
		}
	}
	return featureMap, nil
}

func normalize(geoTiles []geoTile, precision int) []geoTile {
	var tiles []geoTile
	for _, geoTile := range geoTiles {
		tiles = append(tiles, divideTile(geoTile, precision)...)
	}
	return geoTiles
}

// divideTile divides the tile into 4^level smaller ones with same value
func divideTile(tile geoTile, level int) []geoTile {
	if level == 0 {
		return []geoTile{tile}
	}
	var tiles []geoTile
	tiles = append(tiles, divideTile(tile, level-1)...)
	tiles = append(tiles, divideTile(tile, level-1)...)
	tiles = append(tiles, divideTile(tile, level-1)...)
	tiles = append(tiles, divideTile(tile, level-1)...)
	return tiles
}
