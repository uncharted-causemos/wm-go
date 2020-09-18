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

// Datacube represent a datacube object
type Datacube struct {
	ID                     string                   `json:"id"`
	Type                   string                   `json:"type"`
	Model                  string                   `json:"model"`
	Category               []string                 `json:"category"`
	ModelDescription       string                   `json:"model_description"`
	Label                  string                   `json:"label"`
	Maintainer             string                   `json:"maintainer"`
	Source                 string                   `json:"source"`
	OutputName             string                   `json:"output_name"`
	OutputDescription      string                   `json:"output_description"`
	OutputUnits            string                   `json:"output_units"`
	OutputUnitsDescription string                   `json:"output_units_description"`
	Parameters             []string                 `json:"parameters"`
	ParameterDescriptions  []string                 `json:"parameter_descriptions"`
	Concepts               []DatacubeConceptMapping `json:"concepts"`
	Country                []string                 `json:"country"`
	Admin1                 []string                 `json:"admin1"`
	Admin2                 []string                 `json:"admin2"`
	Period                 []DateRange              `json:"period"`
}

// DateRange represent a date range
type DateRange struct {
	Gte string `json:"gte"`
	Lte string `json:"lte"`
}

// DatacubeConceptMapping represent a related concept mapped to corresponding datacube
type DatacubeConceptMapping struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
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

	// SearchDatacubes search and returns datacubes
	SearchDatacubes(filters []*Filter) ([]*Datacube, error)

	// CountDatacubes returns datacubes count
	CountDatacubes(filters []*Filter) (uint64, error)

	// GetConcepts returns list of concepts
	GetConcepts() ([]string, error)
}

// ModelService defines the interface for external REST API
type ModelService interface {
	GetModelParameters(model string) ([]*ModelParameter, error)
	GetConcepts() ([]string, error)
}
