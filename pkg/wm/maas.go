package wm

// ModelRun represent a model run
type ModelRun struct {
	ID         string              `json:"id"`
	Model      string              `json:"model"`
	Parameters []ModelRunParameter `json:"parameters"`
}

// DatacubeParams represent common parameters for requesting model run data
type DatacubeParams struct {
	DataID          string `json:"data_id"`
	RunID           string `json:"run_id"`
	Feature         string `json:"feature"`
	Resolution      string `json:"resolution"`
	TemporalAggFunc string `json:"temporal_agg"`
	SpatialAggFunc  string `json:"spatial_agg"`
}

// OldModelOutputTimeseries represent the old time series model output data
type OldModelOutputTimeseries struct {
	Timeseries []TimeseriesValue `json:"timeseries"`
}

// TimeseriesValue represent a timeseries data point
type TimeseriesValue struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// ModelOutputRawDataPoint represent a raw data point
type ModelOutputRawDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
	Admin2    string  `json:"admin2"`
	Admin3    string  `json:"admin3"`
	Value     float64 `json:"value"`
}

// ModelOutputStat represent min and max stat of the model output data
type ModelOutputStat struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ModelRegionalOutputStat represent regional data for all admin levels
type ModelRegionalOutputStat struct {
	Country *ModelOutputStat `json:"country"`
	Admin1  *ModelOutputStat `json:"admin1"`
	Admin2  *ModelOutputStat `json:"admin2"`
	Admin3  *ModelOutputStat `json:"admin3"`
}

// ModelOutputRegionalAdmins represent regional data for all admin levels
type ModelOutputRegionalAdmins struct {
	Country []ModelOutputAdminData `json:"country"`
	Admin1  []ModelOutputAdminData `json:"admin1"`
	Admin2  []ModelOutputAdminData `json:"admin2"`
	Admin3  []ModelOutputAdminData `json:"admin3"`
}

// ModelOutputAdminData represent a data point of regional data
type ModelOutputAdminData struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
}

// ModelRunParameter represent a model run parameter value
type ModelRunParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ModelParameter represent a model parameter metadata
type ModelParameter struct {
	Choices     []interface{} `json:"choices,omitempty"`
	Default     interface{}   `json:"default"`
	Description string        `json:"description"`
	Maximum     interface{}   `json:"maximum,omitempty"`
	Minimum     interface{}   `json:"minimum,omitempty"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
}

// IndicatorDataPoint represent a data point for an indicator datacube
type IndicatorDataPoint struct {
	Admin1    string  `json:"admin1"`
	Admin2    string  `json:"admin2"`
	Country   string  `json:"country"`
	Dataset   string  `json:"dataset"`
	Unit      string  `json:"value_unit"`
	Mean      float64 `json:"mean"`
	Sum       float64 `json:"sum"`
	Timestamp float64 `json:"timestamp"`
	Variable  string  `json:"variable"`
	Value     float64 `json:"value"`
}

// Datacube represent a datacube object
type Datacube struct {
	ID                     string                   `json:"id"`
	Type                   string                   `json:"type"`
	Model                  string                   `json:"model"`
	ModelID                string                   `json:"model_id"`
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
	SearchScore            float64                  `json:"_search_score,omitempty"`
	Variable               string                   `json:"variable,omitempty"`
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

	// GetModelRuns returns all model runs for the given model
	GetModelRuns(model string) ([]*ModelRun, error)

	// GetModelParameters returns available parameters for the model
	GetModelParameters(model string) ([]*ModelParameter, error)

	// GetIndicatorData returns the indicator time series
	GetIndicatorData(indicatorName string, modelName string, unit []string) ([]*IndicatorDataPoint, error)

	// SearchDatacubes search and returns datacubes
	SearchDatacubes(filters []*Filter) ([]*Datacube, error)

	// CountDatacubes returns datacubes count
	CountDatacubes(filters []*Filter) (uint64, error)

	// GetConcepts returns a list of concepts
	GetConcepts() ([]string, error)

	// GetOutputStats returns model output stats
	GetOutputStats(runID string, feature string) (*ModelOutputStat, error)

	// GetOutputTimeseries returns model output timeseries
	GetOutputTimeseries(runID string, feature string) (*OldModelOutputTimeseries, error)
}

// DataOutput defines the methods that output database implementation needs to satisfy
type DataOutput interface {
	// GetTile returns mapbox vector tile
	GetTile(zoom, x, y uint32, specs GridTileOutputSpecs, expression string) (*Tile, error)

	// GetOutputStats returns datacube output stats
	GetOutputStats(params DatacubeParams, filename string) (*ModelOutputStat, error)

	// GetRegionalOutputStats returns regional output statistics
	GetRegionalOutputStats(params DatacubeParams) (*ModelRegionalOutputStat, error)

	// GetOutputTimeseries returns datacube output timeseries
	GetOutputTimeseries(params DatacubeParams) ([]*TimeseriesValue, error)

	// GetOutputTimeseriesByRegion returns timeseries data for a specific region
	GetOutputTimeseriesByRegion(params DatacubeParams, regionID string) ([]*TimeseriesValue, error)

	// GetRegionAggregation returns regional data for ALL admin regions at ONE timestamp
	GetRegionAggregation(params DatacubeParams, timestamp string) (*ModelOutputRegionalAdmins, error)

	// GetRawData returns datacube output or indicator raw data
	GetRawData(params DatacubeParams) ([]*ModelOutputRawDataPoint, error)
}

// VectorTile defines methods that tile storage/database needs to satisfy
type VectorTile interface {
	GetVectorTile(zoom, x, y uint32, tilesetName string) ([]byte, error)
}
