package wm

import (
	"fmt"

	"github.com/paulmach/orb/geojson"
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

// Tile is an individual tile from MaaS.
type Tile struct {
	zoom, x, y int
	features   []*geojson.Feature
}

// Bound returns tile bounds
func (t *Tile) Bound() {
	// return orb.Bound{
	// 	orb.Point{1.1, 2.2},
	// 	orb.Point{1.2, 2.3},
	// }
	fmt.Println("NYI")
}

// AddFeatures loads geo features to the tile
func (t *Tile) AddFeatures() {
	fmt.Println("NYI")
}

// ToMVT returns the tile as mapbox vector tile format
func (t *Tile) ToMVT() (string, error) {
	return fmt.Sprintf("%d/%d/%d", t.zoom, t.x, t.y), nil
}

// NewTile creates a new tile
func NewTile(zoom int, x int, y int) Tile {
	var features []*geojson.Feature
	return Tile{
		zoom,
		x,
		y,
		features,
	}
}

// // Tile interface
// type Tile interface {
// 	Bound() Bound
// 	AddFeatures()
// 	ToMVT() (string, error)
// }
