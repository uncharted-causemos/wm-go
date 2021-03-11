package wm

import (
	"encoding/json"

	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
)

// TileDataSpecs is a list of TileDataSpecs
type TileDataSpecs []TileDataSpec

// TileDataSpec defines the tile data specifications to be used in the queries.
type TileDataSpec struct {
	Model        string `json:"model"`
	RunID        string `json:"runId"`
	Feature      string `json:"feature"`
	Date         string `json:"date"`
	ValueProp    string `json:"valueProp"`
	MaxPrecision uint32 `json:"maxPrecision"`
}

// Point is a lon/lat point
type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Bound represent rectangular bound
type Bound struct {
	TopLeft     Point `json:"topLeft"`
	BottomRight Point `json:"bottomRight"`
}

// Tile is an individual tile from MaaS.
type Tile struct {
	Zoom     uint32                    `json:"zoom"`
	X        uint32                    `json:"x"`
	Y        uint32                    `json:"y"`
	Layer    string                    `json:"layer"`
	Features geojson.FeatureCollection `json:"features"`
}

// Bound returns tile bounds
func (t *Tile) Bound() Bound {
	bound := maptile.New(t.X, t.Y, maptile.Zoom(t.Zoom)).Bound()
	return Bound{
		Point{bound.LeftTop().Lat(), bound.LeftTop().Lon()},
		Point{bound.RightBottom().Lat(), bound.RightBottom().Lon()},
	}
}

// AddFeature loads geo feature to the tile
func (t *Tile) AddFeature(feature *geojson.Feature) {
	t.Features.Append(feature)
}

// MVT returns the tile as mapbox vector tile format
func (t *Tile) MVT() ([]byte, error) {
	collections := map[string]*geojson.FeatureCollection{
		t.Layer: &t.Features,
	}
	layers := mvt.NewLayers(collections)
	layers.ProjectToTile(maptile.New(t.X, t.Y, maptile.Zoom(t.Zoom)))
	data, err := mvt.MarshalGzipped(layers)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (t *Tile) String() string {
	s, _ := json.MarshalIndent(t, "", "  ")
	return string(s)
}

// NewTile creates a new tile
func NewTile(zoom, x, y uint32, layerName string) *Tile {
	features := *geojson.NewFeatureCollection()
	return &Tile{
		zoom,
		x,
		y,
		layerName,
		features,
	}
}

// MvtToJSON parses mapbox vector tile into json. Json representation of the vector tile would be useful for debugging
func MvtToJSON(tile []byte) ([]byte, error) {
	layers, err := mvt.UnmarshalGzipped(tile)
	if err != nil {
		return nil, err
	}
	json, err := json.MarshalIndent(layers, "", "  ")
	if err != nil {
		return nil, err
	}
	return json, nil
}
