package wm

// MaaS defines the methods that the MaaS database implementation needs to
// satisfy.
type MaaS interface {
	// GetTile returns mapbox vector tile
	GetTile(zoom, x, y uint32, specs TileDataSpecs) ([]byte, error)
}
