package wm

// DataOutputTile defines the methods that the MaaS database implementation needs to
// satisfy.
type DataOutputTile interface {
	// GetTile returns mapbox vector tile
	GetTile(zoom, x, y uint32, specs TileDataSpecs, expression string) (*Tile, error)
}
