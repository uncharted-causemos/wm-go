package wm

import (
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
)

// TileDataSpecs is a list of TileDataSpecs
type TileDataSpecs []TileDataSpec

// TileDataSpec defines the tile data specifications to be used in the queries.
type TileDataSpec struct {
	Model     string `json:"model"`
	RunID     string `json:"runId"`
	Feature   string `json:"feature"`
	Date      string `json:"date"`
	ValueProp string `json:"valueProp"`
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
	zoom, x, y uint32
	layer      string
	features   geojson.FeatureCollection
}

// Bound returns tile bounds
func (t *Tile) Bound() Bound {
	bound := maptile.New(t.x, t.y, maptile.Zoom(t.zoom)).Bound()
	return Bound{
		Point{bound.LeftTop().Lat(), bound.LeftTop().Lon()},
		Point{bound.RightBottom().Lat(), bound.RightBottom().Lon()},
	}
}

// AddFeature loads geo feature to the tile
func (t *Tile) AddFeature(feature geojson.Feature) {
	t.features.Append(&feature)
}

// MVT returns the tile as mapbox vector tile format
func (t *Tile) MVT() ([]byte, error) {
	collections := map[string]*geojson.FeatureCollection{
		t.layer: &t.features,
	}
	layers := mvt.NewLayers(collections)
	layers.ProjectToTile(maptile.New(t.x, t.y, maptile.Zoom(t.zoom)))
	data, err := mvt.MarshalGzipped(layers)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewTile creates a new tile
func NewTile(zoom, x, y uint32, layerName string) Tile {
	features := *geojson.NewFeatureCollection()
	return Tile{
		zoom,
		x,
		y,
		layerName,
		features,
	}
}
