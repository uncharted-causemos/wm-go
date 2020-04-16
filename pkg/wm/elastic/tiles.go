package elastic

import (
	"fmt"

	"github.com/paulmach/orb/geojson"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// Tile is an individual tile from MaaS.
// (TODO: I might need to move this to different place)
type Tile struct {
	zoom, x, y int
	features   []*geojson.Feature
}

// GetBounds returns tile bounds
func (t *Tile) GetBounds() {
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
func NewTile(zoom int, x int, y int) *Tile {
	var features []*geojson.Feature
	return &Tile{
		zoom,
		x,
		y,
		features,
	}
}

// GetTile returns the tile.
func (es *ES) GetTile(zoom int, x int, y int, specs wm.TileDataSpecs) (wm.Tile, error) {
	tile := NewTile(zoom, x, y)

	// TODO: get model output data for each of the TileDataSpec,
	// combine them in to geojson features, and add the geo features(and their values) to the tile
	var results []interface{}
	for _, spec := range specs {
		results = append(results, <-es.getRunOutput("bound", spec))
	}
	for _, r := range results {
		// TODO: process the results and add the features to the tile
		fmt.Println(r)
	}
	return tile, nil
}

// get model run output data for given bounds
func (es *ES) getRunOutput(bound string, spec wm.TileDataSpec) <-chan interface{} {
	r := make(chan interface{})
	go func() {
		// TODO: query es and get result
		r <- "result: " + spec.Feature
		close(r)
	}()
	return r
}
