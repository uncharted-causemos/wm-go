package storage

import (
	"fmt"
	"io/ioutil"
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	pb "gitlab.uncharted.software/WM/wm-proto/tiles"
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
	spec  wm.TileDataSpec
	data  []geoTile
}

// GetTile returns the tile containing model run output specified by the spec
func (s *Storage) GetTile(zoom, x, y uint32, specs wm.TileDataSpecs) ([]byte, error) {
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

	featureMap, err := createFeatures(results)
	if err != nil {
		return nil, err
	}
	for _, feature := range featureMap {
		tile.AddFeature(feature)
	}
	return tile.MVT()
}

// getRunOutput returns geotiled bucket aggregation result of the model run output specified by the spec, bound and zoom
func (s *Storage) getRunOutput(zoom, x, y uint32, spec wm.TileDataSpec) (chan geoTilesResult, chan error) {
	// TODO: Incorporate Feature
	out := make(chan geoTilesResult)
	er := make(chan error)
	go func() {
		defer close(er)
		defer close(out)
		// startTime, err := time.Parse(time.RFC3339, spec.Date)
		// if err != nil {
		// 	er <- err
		// 	return
		// }
		// key := fmt.Sprintf("%s/%s/%s/%d-%d-%d-%d.tile", spec.Model, spec.RunID, spec.Feature, startTime.Unix(), zoom, x, y)
		key := fmt.Sprintf("%s/%s/%s/%s-%d-%d-%d.tile", "consumption_model", "1aee48cd4d5286732367dc223f7b21e97bc23619815f7140763c2f9f7541dfac", "consumption_per_capita_per_day", "1262304000000", zoom, x, y)

		// TODO: Modify tile protobuf mapping and tile pipeline to save multiple zooms in one tile
		// TODO: Get max zoom across all requested models and split up tiles on the fly to account
		req, resp := s.client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
		err := req.Send()
		if err != nil {
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

		totalBins := tile.Bins.TotalBins
		totalBinsXY := uint32(math.Pow(2, (math.Log(float64(totalBins)) / math.Log(4))))
		var gts []geoTile
		for binPosition, binStats := range tile.Bins.Stats {
			gt := geoTile{
				Key: fmt.Sprintf("%d/%d/%d",
					tile.Coord.Z+uint32(math.Log2(float64(totalBinsXY))),
					tile.Coord.X*totalBinsXY+uint32(math.Mod(float64(binPosition), float64(totalBinsXY))),
					tile.Coord.Y*totalBinsXY+binPosition/totalBinsXY),
				SpatialAggregation: geoTileAggregation{Value: binStats.Sum},
			}
			gts = append(gts, gt)
		}
		wmTile := wm.NewTile(zoom, x, y, tileDataLayerName)
		result := geoTilesResult{
			bound: bound(wmTile.Bound()),
			zoom:  int(zoom),
			spec:  spec,
			data:  gts,
		}
		er <- nil
		out <- result
	}()
	return out, er
}

// createFeatures processes and merges the results and returns a map of geojson feature
func createFeatures(results []geoTilesResult) (map[string]geojson.Feature, error) {
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
