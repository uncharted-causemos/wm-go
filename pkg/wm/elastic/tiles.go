package elastic

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
	bound bound
	zoom  int
	spec  wm.TileDataSpec
	data  []geoTile
}

var modelMaxPrecision = map[string]uint32{
	"G-Range":            10,
	"consumption_model":  14,
	"asset_wealth_model": 14,
	"malnutrition_model": 15,
	"population_model":   15,
	"PIHM":               16,
	//
	// Note: could not figure out max precision for following modles since we don't have them in our es currently
	//
	// "cropland_model":          99,
	// "world_population_africa": 99,
	// "CHIRPS":                  99,
	// "flood_index_model":       99,
	// "DSSAT":                   99,
	// "APSIM":                   99,
	// "CLEM":                    99,
	// "lpjml":                   99,
	// "lpjml_historic":          99,
	// "yield_anomalies_lpjml":   99,
	// "FSC":                     99,
	// "market_price_model":      99,
	// "multi_twist":             99,
}

// GetTile returns the tile containing model run output specified by the spec
func (es *ES) GetTile(zoom, x, y uint32, specs wm.TileDataSpecs) ([]byte, error) {
	tile := wm.NewTile(zoom, x, y, tileDataLayerName)

	var errChs []chan error
	var resChs []chan geoTilesResult
	var results []geoTilesResult

	for _, spec := range specs {
		res, err := es.getRunOutput(bound(tile.Bound()), zoom, spec)
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
func (es *ES) getRunOutput(bound bound, zoom uint32, spec wm.TileDataSpec) (chan geoTilesResult, chan error) {
	// TODO: Cache the result of this function since this is resource/compute intensive
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
		// Target precision for the results, + 6 precision results 4096 cells in the bound. More details: https://wiki.openstreetmap.org/wiki/Zoom_levels
		precision := zoom + 6
		maxPrecision, ok := modelMaxPrecision[spec.Model]
		if !ok {
			maxPrecision = precision
		}
		// let output precision be >= zoom and <= maxPrecision
		outputPrecision := math.Max(float64(zoom), math.Min(float64(maxPrecision), float64(precision)))
		data := map[string]interface{}{
			"RunID":       spec.RunID,
			"Feature":     spec.Feature,
			"Bound":       string(b),
			"Precision":   outputPrecision,
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
			bound: bound,
			zoom:  int(zoom),
			spec:  spec,
			data:  []geoTile{},
		}
		if err := json.Unmarshal([]byte(buckets), &result.data); err != nil {
			er <- err
			return
		}
		er <- nil
		result.data = subDivideTiles(result.data, precision)
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

// subDivideTiles divides each tile of geoTiles into subdivided tiles at given precision and returns the result.
// If tile precision(zoom level) of given tiles >= precision, just returns the original geoTiles
func subDivideTiles(geoTiles []geoTile, precision uint32) []geoTile {
	var tiles []geoTile
	for _, geoTile := range geoTiles {
		tiles = append(tiles, divideTile(geoTile, precision)...)
	}
	return tiles
}

// divideTile divides the tile into 4^level smaller subtiles
// For given tile with zoom level 1 and 2 as level param, it will produce 16 (4^2) subtiles of zoom level 3
func divideTile(tile geoTile, level uint32) []geoTile {
	var z, x, y uint32
	fmt.Sscanf(tile.Key, "%d/%d/%d", &z, &x, &y)

	if level <= z {
		return []geoTile{tile}
	}
	var tiles []geoTile
	topLeft := geoTile{
		Key:                fmt.Sprintf("%d/%d/%d", z+1, 2*x, 2*y),
		SpatialAggregation: tile.SpatialAggregation,
	}
	topRight := geoTile{
		Key:                fmt.Sprintf("%d/%d/%d", z+1, 2*x+1, 2*y),
		SpatialAggregation: tile.SpatialAggregation,
	}
	bottomLeft := geoTile{
		Key:                fmt.Sprintf("%d/%d/%d", z+1, 2*x, 2*y+1),
		SpatialAggregation: tile.SpatialAggregation,
	}
	bottomRight := geoTile{
		Key:                fmt.Sprintf("%d/%d/%d", z+1, 2*x+1, 2*y+1),
		SpatialAggregation: tile.SpatialAggregation,
	}
	tiles = append(tiles, divideTile(topLeft, level-1)...)
	tiles = append(tiles, divideTile(topRight, level-1)...)
	tiles = append(tiles, divideTile(bottomLeft, level-1)...)
	tiles = append(tiles, divideTile(bottomRight, level-1)...)
	return tiles
}
