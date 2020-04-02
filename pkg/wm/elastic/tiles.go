package elastic

import "gitlab.uncharted.software/WM/wm-go/pkg/wm"

// GetTiles returns the tiles.
func (es *ES) GetTiles(filters []*wm.Filter) (wm.Tiles, error) {
	return []*wm.Tile{}, nil
}
