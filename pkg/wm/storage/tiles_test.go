package storage

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/stretchr/testify/require"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

func toPolygon(z, x, y uint32) orb.Polygon {
	return maptile.New(x, y, maptile.Zoom(z)).Bound().ToPolygon()
}

func TestTiles(t *testing.T) {
	s, err := New(nil, "")
	if err != nil {
		log.Fatal(err)
	}
	specs := wm.TileDataSpecs{
		wm.TileDataSpec{Model: "consumption_model", RunID: "1aee48cd4d5286732367dc223f7b21e97bc23619815f7140763c2f9f7541dfac", Feature: "FEATURE_NAME", Date: "2020-01"},
	}
	s.GetTile(9, 322, 244, specs)
	require.NotNil(t, s)
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
