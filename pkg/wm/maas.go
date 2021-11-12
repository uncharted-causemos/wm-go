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

// HierarchyParams represent parameters needed to fetch a region hierarchy
type HierarchyParams struct {
	DataID  string `json:"data_id"`
	RunID   string `json:"run_id"`
	Feature string `json:"feature"`
}

// RegionListParams represent parameters needed to fetch region lists representing the hierarchy
type RegionListParams struct {
	DataID  string   `json:"data_id"`
	RunIDs  []string `json:"run_ids"`
	Feature string   `json:"feature"`
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

// ModelOutputQualifierTimeseries represent a timeseries for one qualifier value
type ModelOutputQualifierTimeseries struct {
	Name       string             `json:"name"`
	Timeseries []*TimeseriesValue `json:"timeseries"`
}

// ModelOutputRegionQualifierBreakdown represent a list of qualifier breakdown values for a specific region
type ModelOutputRegionQualifierBreakdown struct {
	ID     string             `json:"id"`
	Values map[string]float64 `json:"values"`
}

// ModelOutputQualifierBreakdown represent a list of qualifier breakdown values
type ModelOutputQualifierBreakdown struct {
	Name    string                       `json:"name"`
	Options []*ModelOutputQualifierValue `json:"options"`
}

// ModelOutputQualifierValue represent a breakdown value for one qualifier value
type ModelOutputQualifierValue struct {
	Name  string   `json:"name"`
	Value *float64 `json:"value"` //nil represents missing value
}

// ModelOutputHierarchy is a hierarchy where each region maps to a map of more specific regions.
type ModelOutputHierarchy map[string]interface{}

// ModelOutputStat represent min and max stat of the model output data
type ModelOutputStat struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// OutputStatWithZoom represent min and max stat of the output data for a specific zoom level
type OutputStatWithZoom struct {
	Zoom uint8   `json:"zoom"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// ModelRegionalOutputStat represent regional data for all admin levels
type ModelRegionalOutputStat struct {
	Country *ModelOutputStat `json:"country"`
	Admin1  *ModelOutputStat `json:"admin1"`
	Admin2  *ModelOutputStat `json:"admin2"`
	Admin3  *ModelOutputStat `json:"admin3"`
}

// RegionListOutput represents region list hierarchies for all admin levels
type RegionListOutput struct {
	Country []string `json:"country"`
	Admin1  []string `json:"admin1"`
	Admin2  []string `json:"admin2"`
	Admin3  []string `json:"admin3"`
}

// ModelOutputRegionalAdmins represent regional data for all admin levels
type ModelOutputRegionalAdmins struct {
	Country []ModelOutputAdminData `json:"country"`
	Admin1  []ModelOutputAdminData `json:"admin1"`
	Admin2  []ModelOutputAdminData `json:"admin2"`
	Admin3  []ModelOutputAdminData `json:"admin3"`
}

// ModelOutputRegionalQualifiers represent regional data for all admin levels broken down by qualifiers
type ModelOutputRegionalQualifiers struct {
	Country []ModelOutputRegionQualifierBreakdown `json:"country"`
	Admin1  []ModelOutputRegionQualifierBreakdown `json:"admin1"`
	Admin2  []ModelOutputRegionQualifierBreakdown `json:"admin2"`
	Admin3  []ModelOutputRegionQualifierBreakdown `json:"admin3"`
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

type TransformConfig struct {
	Transform  string `json:"transform"`
	RegionID   string `json:"region_id"`
	Resolution string `json:"resolution"`
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
}

// DataOutput defines the methods that output database implementation needs to satisfy
type DataOutput interface {
	// GetTile returns mapbox vector tile
	GetTile(zoom, x, y uint32, specs GridTileOutputSpecs, expression string) (*Tile, error)

	// GetOutputStats returns datacube output stats
	GetOutputStats(params DatacubeParams, timestamp string) ([]*OutputStatWithZoom, error)

	// GetRegionalOutputStats returns regional output statistics
	GetRegionalOutputStats(params DatacubeParams) (*ModelRegionalOutputStat, error)

	// GetOutputTimeseries returns datacube output timeseries
	GetOutputTimeseries(params DatacubeParams) ([]*TimeseriesValue, error)

	// GetOutputTimeseriesByRegion returns timeseries data for a specific region
	GetOutputTimeseriesByRegion(params DatacubeParams, regionID string) ([]*TimeseriesValue, error)

	// GetRegionAggregation returns regional data for ALL admin regions at ONE timestamp
	GetRegionAggregation(params DatacubeParams, timestamp string) (*ModelOutputRegionalAdmins, error)

	// GetRawData returns datacube raw data
	GetRawData(params DatacubeParams) ([]*ModelOutputRawDataPoint, error)

	// GetRegionHierarchy returns region hierarchy output
	GetRegionHierarchy(params HierarchyParams) (*ModelOutputHierarchy, error)

	// GetHierarchyLists returns region hierarchies in list form
	GetHierarchyLists(params RegionListParams) (*RegionListOutput, error)

	// GetQualifierTimeseries returns datacube output timeseries broken down by qualifiers
	GetQualifierTimeseries(params DatacubeParams, qualifier string, qualifierOptions []string) ([]*ModelOutputQualifierTimeseries, error)

	// GetQualifierData returns datacube output data broken down by qualifiers for ONE timestamp
	GetQualifierData(params DatacubeParams, timestamp string, qualifiers []string) ([]*ModelOutputQualifierBreakdown, error)

	// GetQualifierRegional returns datacube output data broken down by qualifiers for ONE timestamp
	GetQualifierRegional(params DatacubeParams, timestamp string, qualifier string) (*ModelOutputRegionalQualifiers, error)

	// TransformOutputTimeseriesByRegion returns transformed timeseries data
	TransformOutputTimeseriesByRegion(timeseries []*TimeseriesValue, config TransformConfig) ([]*TimeseriesValue, error)

	// TransformRegionAggregation returns transformed regional data for ALL admin regions at ONE timestamp
	TransformRegionAggregation(data *ModelOutputRegionalAdmins, timestamp string, config TransformConfig) (*ModelOutputRegionalAdmins, error)
}

// VectorTile defines methods that tile storage/database needs to satisfy
type VectorTile interface {
	GetVectorTile(zoom, x, y uint32, tilesetName string) ([]byte, error)
}
