package wm

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

// Tile interface
type Tile interface {
	GetBounds()
	AddFeatures()
	ToMVT() (string, error)
}
