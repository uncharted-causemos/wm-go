package wm

// MaaS defines the methods that the MaaS database implementation needs to
// satisfy.
type MaaS interface {
	GetTile(zoom int, x int, y int, specs TileDataSpecs) (Tile, error)
}
