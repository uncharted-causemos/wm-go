package elastic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func toPolygon(z, x, y uint32) orb.Polygon {
	return maptile.New(x, y, maptile.Zoom(z)).Bound().ToPolygon()
}
func TestCreateFeatures(t *testing.T) {
	tests := []struct {
		input  []geoTilesResult
		expect map[string]geojson.Feature
	}{
		{
			input: []geoTilesResult{
				{
					spec: wm.TileDataSpec{ValueProp: "crop"},
					data: []geoTile{
						{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 3}},
						{Key: "5/20/15", SpatialAggregation: geoTileAggregation{Value: 5}},
					},
				},
				{
					spec: wm.TileDataSpec{ValueProp: "rainfall"},
					data: []geoTile{
						{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 30}},
						{Key: "5/20/15", SpatialAggregation: geoTileAggregation{Value: 49}},
						{Key: "5/20/16", SpatialAggregation: geoTileAggregation{Value: 70}},
					},
				},
			},
			expect: map[string]geojson.Feature{
				"5/19/15": {Type: "Feature", Geometry: toPolygon(5, 19, 15), Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30}},
				"5/20/15": {Type: "Feature", Geometry: toPolygon(5, 20, 15), Properties: geojson.Properties{"id": "5/20/15", "crop": 5, "rainfall": 49}},
				"5/20/16": {Type: "Feature", Geometry: toPolygon(5, 20, 16), Properties: geojson.Properties{"id": "5/20/16", "rainfall": 70}},
			},
		},
	}
	for i, test := range tests {
		results, _ := createFeatures(test.input)
		// Note: reflect.DeepEqual failed with comparing maps when keys are in different order. So compare using json instead
		actual, _ := json.MarshalIndent(results, "", "\t")
		expect, _ := json.MarshalIndent(test.expect, "", "\t")
		if string(actual) != string(expect) {
			t.Errorf("Test case %d\ncreateFeatures returned: \n%s\ninstead of:\n%s\n for input %v", i, string(actual), string(expect), spew.Sdump(test.input))
		}
	}
}

func TestSubDivideTiles(t *testing.T) {
	type params struct {
		geoTiles  []geoTile
		precision uint32
	}
	tests := []struct {
		input  params
		expect []geoTile
	}{
		{
			input: params{
				geoTiles: []geoTile{
					{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 3}},
					{Key: "5/20/15", SpatialAggregation: geoTileAggregation{Value: 5}},
				},
				precision: 6,
			},
			expect: []geoTile{
				{Key: "6/38/30", SpatialAggregation: geoTileAggregation{Value: 3}},
				{Key: "6/39/30", SpatialAggregation: geoTileAggregation{Value: 3}},
				{Key: "6/38/31", SpatialAggregation: geoTileAggregation{Value: 3}},
				{Key: "6/39/31", SpatialAggregation: geoTileAggregation{Value: 3}},

				{Key: "6/40/30", SpatialAggregation: geoTileAggregation{Value: 5}},
				{Key: "6/41/30", SpatialAggregation: geoTileAggregation{Value: 5}},
				{Key: "6/40/31", SpatialAggregation: geoTileAggregation{Value: 5}},
				{Key: "6/41/31", SpatialAggregation: geoTileAggregation{Value: 5}},
			},
		},
		{
			input: params{
				geoTiles: []geoTile{
					{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 2}},
				},
				precision: 7,
			},
			expect: []geoTile{
				{Key: "7/76/60", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/77/60", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/76/61", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/77/61", SpatialAggregation: geoTileAggregation{Value: 2}},

				{Key: "7/78/60", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/79/60", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/78/61", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/79/61", SpatialAggregation: geoTileAggregation{Value: 2}},

				{Key: "7/76/62", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/77/62", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/76/63", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/77/63", SpatialAggregation: geoTileAggregation{Value: 2}},

				{Key: "7/78/62", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/79/62", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/78/63", SpatialAggregation: geoTileAggregation{Value: 2}},
				{Key: "7/79/63", SpatialAggregation: geoTileAggregation{Value: 2}},
			},
		},
		{
			input: params{
				geoTiles: []geoTile{
					{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 3}},
				},
				precision: 5,
			},
			expect: []geoTile{
				{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 3}},
			},
		},
		{
			input: params{
				geoTiles: []geoTile{
					{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 3}},
				},
				precision: 4,
			},
			expect: []geoTile{
				{Key: "5/19/15", SpatialAggregation: geoTileAggregation{Value: 3}},
			},
		},
		{
			input: params{
				geoTiles:  []geoTile{},
				precision: 4,
			},
			expect: []geoTile{},
		},
	}
	for i, test := range tests {
		results := subDivideTiles(test.input.geoTiles, test.input.precision)
		equal := reflect.DeepEqual(results, test.expect)
		if !equal {
			t.Errorf("Test case %d\nsubdivideTiles returned: \n%v\ninstead of:\n%v\n for input %v", i, spew.Sdump(results), spew.Sdump(test.expect), spew.Sdump(test.input))
		}
	}
}
