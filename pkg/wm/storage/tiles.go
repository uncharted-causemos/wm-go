package storage

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	pb "gitlab.uncharted.software/WM/wm-go/proto/tiles"
	"google.golang.org/protobuf/proto"
)

const tileDataLayerName = "maas"

// bound represents a geo bound
type bound struct {
	TopLeft     wm.Point `json:"top_left"`
	BottomRight wm.Point `json:"bottom_right"`
}

// geoTile is a single record of the ES geotile bucket aggregation result
type geoTile struct {
	Key                string             `json:"key"`
	SpatialAggregation geoTileAggregation `json:"spatial_aggregation"`
}

type geoTileAggregation struct {
	Value float64 `json:"value"`
}

// geoTilesResult is the ES geotile bucket aggregation result
type geoTilesResult struct {
	bound bound
	zoom  int
	spec  wm.GridTileOutputSpec
	data  []geoTile
}

// GetTile returns the tile containing model run output specified by the spec
func (s *Storage) GetTile(zoom, x, y uint32, specs wm.GridTileOutputSpecs, expression string) (*wm.Tile, error) {
	tile := wm.NewTile(zoom, x, y, tileDataLayerName)

	var errChs []chan error
	var resChs []chan geoTilesResult
	var results []geoTilesResult

	for _, spec := range specs {
		res, err := s.getRunOutput(zoom, x, y, spec)
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

	features, err := createFeatures(results)
	if err != nil {
		return nil, err
	}

	if expression != "" {
		if err := evaluateExpression(features, expression); err != nil {
			return nil, err
		}
	}

	for _, feature := range features {
		tile.AddFeature(feature)
	}
	return tile, nil
}

// evaluateExpression evaluate expression using feature properties as parameters and add the result back as new property to the given feature
func evaluateExpression(features []*geojson.Feature, expression string) error {
	op := "evaluateExpression"
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	for _, feature := range features {
		parameters := make(map[string]interface{})
		for key, value := range feature.Properties {
			if key != "id" {
				parameters[key] = value
			}
		}
		result, err := exp.Evaluate(parameters)
		if err != nil {
			// If the expression can not be evaluated, omit result property
		} else if v, ok := result.(float64); ok && math.IsInf(v, 0) {
			// Check if result is -Inf or Inf (eg. happens when a value is divided by zero)
			// Omit result property in this case a well
		} else {
			feature.Properties["result"] = result
		}
	}
	return nil
}

// getRunOutput returns geotiled bucket aggregation result of the model run output specified by the spec, bound and zoom
func (s *Storage) getRunOutput(zoom, x, y uint32, spec wm.GridTileOutputSpec) (chan geoTilesResult, chan error) {
	out := make(chan geoTilesResult)
	er := make(chan error)
	go func() {
		defer close(er)
		defer close(out)

		modelMaxPrecision := spec.MaxPrecision
		if modelMaxPrecision == 0 {
			// if zero value (not set)
			modelMaxPrecision = 99
		}

		bucketName := maasModelOutputBucket
		key := fmt.Sprintf("%s/%s/%s/%s/tiles/%d-%d-%d-%d.tile", spec.ModelID, spec.RunID, spec.Resolution, spec.Feature, spec.Timestamp, zoom, x, y)

		if spec.Model != "" {
			// For Backward compatibility to support old api and tile outputs
			// TODO: Remove this part if we no longer need to display old tile outputs
			startTime, err := time.Parse(time.RFC3339, spec.Date)
			if err != nil {
				er <- err
				return
			}
			timemillis := startTime.Unix() * 1000
			key = fmt.Sprintf("%s/%s/%s/%d-%d-%d-%d.tile", strings.ToLower(spec.Model), spec.RunID, spec.Feature, timemillis, zoom, x, y)
			bucketName = outputBucket
		}

		// Retrieve protobuf tile from S3
		req, resp := s.client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})

		//TODO: Need validation and better error handling
		err := req.Send()
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == s3.ErrCodeNoSuchKey {
					// // Tile not found errors are expected
					return
				}
				er <- err
				return
			}
			er <- err
			return
		}
		var tile pb.Tile
		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			er <- err
			return
		}
		if err := proto.Unmarshal(buf, &tile); err != nil {
			er <- err
			return
		}

		// Convert tile bin positions into z/x/y tile coordinates and save as geotiles
		totalBins := tile.Bins.TotalBins
		totalBinsXY := uint32(math.Pow(2, (math.Log(float64(totalBins)) / math.Log(4))))
		// bin(subtile) precision
		binPrecision := tile.Coord.Z + uint32(math.Log2(float64(totalBinsXY)))

		// difference between the supported max precision for the model output and the bin(subtile) precision
		var precisionDiff uint32

		if binPrecision > modelMaxPrecision {
			precisionDiff = binPrecision - modelMaxPrecision
		}
		// Note: If there is precision(or zoom level) difference beteween requested tile and the max precision of
		// the output (output resolution at which models look good), aggregate up each tile grid cell to bigger grid cell at max precision
		type binAgg struct {
			sum    float64
			weight float64 // or just count if agg is not weighted
		}
		tileMap := make(map[string]*binAgg)
		var gts []geoTile
		for binPosition, binStats := range tile.Bins.Stats {
			z := binPrecision - precisionDiff
			x := tile.Coord.X*totalBinsXY + uint32(math.Mod(float64(binPosition), float64(totalBinsXY)))
			y := tile.Coord.Y*totalBinsXY + binPosition/totalBinsXY
			// Use parent coord if there is precision difference
			for i := 0; i < int(precisionDiff); i++ {
				x = x / 2
				y = y / 2
			}
			coord := fmt.Sprintf("%d/%d/%d", z, x, y)
			if _, ok := tileMap[coord]; !ok {
				tileMap[coord] = &binAgg{}
			}
			sum, weight := getTileBinValue(binStats, spec.TemporalAggFunc)
			tileMap[coord].sum += sum
			tileMap[coord].weight += weight
		}

		// Create geotiles
		for coord, agg := range tileMap {
			value := agg.sum / float64(agg.weight) // default to mean
			if spec.SpatialAggFunc == "sum" {
				value = agg.sum
			}

			gts = append(gts, geoTile{
				Key:                coord,
				SpatialAggregation: geoTileAggregation{Value: value},
			})
		}

		wmTile := wm.NewTile(zoom, x, y, tileDataLayerName)
		result := geoTilesResult{
			bound: bound(wmTile.Bound()),
			zoom:  int(zoom),
			spec:  spec,
			data:  gts,
		}
		er <- nil
		if precisionDiff > 0 {
			result.data = subDivideTiles(result.data, binPrecision)
		}
		out <- result
	}()
	return out, er
}

func getTileBinValue(tileBinStats *pb.TileStats, temporalAggFunc string) (float64, float64) {
	//For old api backward compatibility
	if temporalAggFunc == "" {
		return tileBinStats.Avg, 1
	} else if temporalAggFunc == "sum" {
		return tileBinStats.SSumTSum, tileBinStats.Weight
	} else {
		return tileBinStats.SSumTMean, tileBinStats.Weight
	}
}

// createFeatures processes and merges the geotile results and returns a list of geojson features
func createFeatures(results []geoTilesResult) ([]*geojson.Feature, error) {
	featureMap := map[string]*geojson.Feature{}
	for _, result := range results {
		for _, gt := range result.data {
			if _, ok := featureMap[gt.Key]; !ok {
				var z, x, y uint32
				if _, err := fmt.Sscanf(gt.Key, "%d/%d/%d", &z, &x, &y); err != nil {
					return nil, err
				}
				polygon := maptile.New(x, y, maptile.Zoom(z)).Bound().ToPolygon()
				f := geojson.NewFeature(polygon)
				f.Properties["id"] = gt.Key
				featureMap[gt.Key] = f
			}
			featureMap[gt.Key].Properties[result.spec.ValueProp] = gt.SpatialAggregation.Value
		}
	}
	var features []*geojson.Feature
	for _, feature := range featureMap {
		features = append(features, feature)
	}
	return features, nil
}

// subDivideTiles divides each tile of geoTiles into subdivided tiles at given precision and returns the result.
// If tile precision(zoom level) of given tiles >= precision, just returns the original geoTiles
func subDivideTiles(geoTiles []geoTile, precision uint32) []geoTile {
	tiles := []geoTile{}
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
	// Details on tile calculation: https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Subtiles
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
	tiles = append(tiles, divideTile(topLeft, level)...)
	tiles = append(tiles, divideTile(topRight, level)...)
	tiles = append(tiles, divideTile(bottomLeft, level)...)
	tiles = append(tiles, divideTile(bottomRight, level)...)
	return tiles
}
