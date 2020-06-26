package wm

// ModelRun represent a model run
type ModelRun struct {
	ID         string              `json:"id"`
	Model      string              `json:"model"`
	Parameters []ModelRunParameter `json:"parameters"`
}

// ModelRunParameter represent a model run parameter value
type ModelRunParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ModelParameter represent a model parameter metadata
type ModelParameter struct {
	Choices     []string    `json:"choices,omitempty"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Maximum     interface{} `json:"maximum,omitempty"`
	Minimum     interface{} `json:"minimum,omitempty"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
}

// MaaS defines the methods that the MaaS database implementation needs to
// satisfy.
type MaaS interface {
	// GetTile returns mapbox vector tile
	GetTile(zoom, x, y uint32, specs TileDataSpecs) ([]byte, error)

	// GetModelRuns returns all model runs for the given model
	GetModelRuns(model string) ([]*ModelRun, error)

	// GetModelParameters returns available parameters for the model
	GetModelParameters(model string) ([]*ModelParameter, error)
}

// ModelService defines the interface for external REST API
type ModelService interface {
	GetModelParameters(model string) ([]*ModelParameter, error)
}
