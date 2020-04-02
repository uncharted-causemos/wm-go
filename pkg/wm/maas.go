package wm

// MaaS defines the methods that the MaaS database implementation needs to
// satisfy.
type MaaS interface {
	GetTiles(filters []*Filter) (Tiles, error)
}
