package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
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
	s.GetTile(9, 322, 244, specs, "")
	require.NotNil(t, s)
}

func TestCreateFeatures(t *testing.T) {
	tests := []struct {
		input  []geoTilesResult
		expect []geojson.Feature
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
			expect: []geojson.Feature{
				{Type: "Feature", Geometry: toPolygon(5, 19, 15), Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30}},
				{Type: "Feature", Geometry: toPolygon(5, 20, 15), Properties: geojson.Properties{"id": "5/20/15", "crop": 5, "rainfall": 49}},
				{Type: "Feature", Geometry: toPolygon(5, 20, 16), Properties: geojson.Properties{"id": "5/20/16", "rainfall": 70}},
			},
		},
	}
	for i, test := range tests {
		results, _ := createFeatures(test.input)
		sort.Slice(results, func(i, j int) bool {
			return fmt.Sprintf("%v", results[i].Properties["id"]) < fmt.Sprintf("%v", results[j].Properties["id"])
		})
		actual, _ := json.MarshalIndent(results, "", "\t")
		expect, _ := json.MarshalIndent(test.expect, "", "\t")
		if string(actual) != string(expect) {
			t.Errorf("Test case %d\ncreateFeatures returned: \n%s\ninstead of:\n%s\n for input %v", i, string(actual), string(expect), spew.Sdump(test.input))
		}
	}
}

func TestEvaluateExpression(t *testing.T) {
	type inputParams struct {
		features   []*geojson.Feature
		expression string
	}
	tests := []struct {
		description string
		input       inputParams
		expect      []*geojson.Feature
	}{
		{
			description: "Test addition",
			input: inputParams{
				features: []*geojson.Feature{
					{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30}},
				},
				expression: "[rainfall] + [crop]",
			},
			expect: []*geojson.Feature{
				{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30, "result": 33}},
			},
		},
		{
			description: "Test subtraction",
			input: inputParams{
				features: []*geojson.Feature{
					{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30}},
					{Properties: geojson.Properties{"id": "5/19/16", "crop": 21, "rainfall": 20}},
				},
				expression: "[rainfall] - [crop]",
			},
			expect: []*geojson.Feature{
				{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30, "result": 27}},
				{Properties: geojson.Properties{"id": "5/19/16", "crop": 21, "rainfall": 20, "result": -1}},
			},
		},
		{
			description: "Test multiplication",
			input: inputParams{
				features: []*geojson.Feature{
					{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30}},
				},
				expression: "[rainfall] * [crop]",
			},
			expect: []*geojson.Feature{
				{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30, "result": 90}},
			},
		},
		{
			description: "Test division",
			input: inputParams{
				features: []*geojson.Feature{
					{Properties: geojson.Properties{"id": "5/19/15", "crop": 4, "rainfall": 30}},
				},
				expression: "[rainfall] / [crop]",
			},
			expect: []*geojson.Feature{
				{Properties: geojson.Properties{"id": "5/19/15", "crop": 4, "rainfall": 30, "result": 7.5}},
			},
		},
		{
			description: "Test division by zero (result should be null)",
			input: inputParams{
				features: []*geojson.Feature{
					{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30}},
					{Properties: geojson.Properties{"id": "5/19/16", "crop": 0, "rainfall": 20}},
					{Properties: geojson.Properties{"id": "5/19/17", "crop": 0, "rainfall": -10}},
				},
				expression: "[rainfall] / [crop]",
			},
			expect: []*geojson.Feature{
				{Properties: geojson.Properties{"id": "5/19/15", "crop": 3, "rainfall": 30, "result": 10}},
				{Properties: geojson.Properties{"id": "5/19/16", "crop": 0, "rainfall": 20, "result": nil}},
				{Properties: geojson.Properties{"id": "5/19/17", "crop": 0, "rainfall": -10, "result": nil}},
			},
		},
	}
	for _, test := range tests {
		evaluateExpression(test.input.features, test.input.expression)
		expect, _ := json.MarshalIndent(test.expect, "", "\t")
		result, _ := json.MarshalIndent(test.input.features, "", "\t")
		if string(result) != string(expect) {
			t.Errorf("%s\ntransformFeature returned: \n%s\ninstead of:\n%s\n for input %v", test.description, result, expect, spew.Sdump(test.input))
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
