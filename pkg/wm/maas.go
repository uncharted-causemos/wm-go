package wm

// TemporalResolution defines the temporal resolution type
type TemporalResolution string

// Temporal resolution type
const (
	TemporalResolutionAnnual  TemporalResolution = "annual"
	TemporalResolutionMonthly TemporalResolution = "monthly"
	TemporalResolutionDekad   TemporalResolution = "dekad"
	TemporalResolutionWeekly  TemporalResolution = "weekly"
	TemporalResolutionDaily   TemporalResolution = "daily"
	TemporalResolutionOther   TemporalResolution = "other"
)

// AggregationOption defines the available aggregation options
type AggregationOption string

// Available aggregation options
const (
	AggregationOptionMean AggregationOption = "mean"
	AggregationOptionSum  AggregationOption = "sum"
)

// TemporalResolutionOption defines the available temporal resolution options
type TemporalResolutionOption string

// Available temporal resolution options
const (
	TemporalResolutionOptionYear  TemporalResolutionOption = "year"
	TemporalResolutionOptionMonth TemporalResolutionOption = "month"
)

// DatacubeParams represent common parameters for requesting model run data
type DatacubeParams struct {
	DataID          string                   `json:"data_id"`
	RunID           string                   `json:"run_id"`
	Feature         string                   `json:"feature"`
	Resolution      TemporalResolutionOption `json:"resolution"`
	TemporalAggFunc AggregationOption        `json:"temporal_agg"`
	SpatialAggFunc  AggregationOption        `json:"spatial_agg"`
}

// FullTimeseriesParams represent all parameters for fetching a timeseries
type FullTimeseriesParams struct {
	DatacubeParams
	RegionID  string    `json:"region_id"`
	Transform Transform `json:"transform"`
	Key       string    `json:"key"`
}

// RegionListParams represent parameters needed to fetch region lists representing the hierarchy
type RegionListParams struct {
	DataID  string   `json:"data_id"`
	RunIDs  []string `json:"run_ids"`
	Feature string   `json:"feature"`
}

// QualifierInfoParams represent common parameters needed to fetch summary info about qualifiers
type QualifierInfoParams struct {
	DataID  string `json:"data_id"`
	RunID   string `json:"run_id"`
	Feature string `json:"feature"`
}

// PipelineResultsParams represent parameters needed to fetch pipeline results
type PipelineResultsParams struct {
	DataID string `json:"data_id"`
	RunID  string `json:"run_id"`
}

// TimeseriesValue represent a timeseries data point
type TimeseriesValue struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type ExtremaValue struct {
	RegionId  string  `json:"region_id"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
}

type ExtremaValues struct {
	SsumTsum   []ExtremaValue `json:"s_sum_t_sum"`
	SmeanTsum  []ExtremaValue `json:"s_mean_t_sum"`
	SsumTmean  []ExtremaValue `json:"s_sum_t_mean"`
	SmeanTmean []ExtremaValue `json:"s_mean_t_mean"`
}

type Extrema struct {
	Min ExtremaValues `json:"min"`
	Max ExtremaValues `json:"max"`
}

// ModelOutputRawDataPoint represent a raw data point
type ModelOutputRawDataPoint struct {
	Timestamp  int64             `json:"timestamp"`
	Country    string            `json:"country"`
	Admin1     string            `json:"admin1"`
	Admin2     string            `json:"admin2"`
	Admin3     string            `json:"admin3"`
	Lat        *float64          `json:"lat"`
	Lng        *float64          `json:"lng"`
	Value      *float64          `json:"value"`
	Qualifiers map[string]string `json:"qualifiers"`
}

// ModelOutputQualifierTimeseries represent a timeseries for one qualifier value
type ModelOutputQualifierTimeseries struct {
	Name       string             `json:"name"`
	Timeseries []*TimeseriesValue `json:"timeseries"`
}

// ModelOutputKeyedTimeSeries holds time series values for a unique key
type ModelOutputKeyedTimeSeries struct {
	Key        string             `json:"key"`
	Timeseries []*TimeseriesValue `json:"timeseries"`
}

// ModelOutputRegionalTimeSeries holds regional time series values
type ModelOutputRegionalTimeSeries struct {
	RegionID   string             `json:"region_id"`
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

// QualifierCountsOutput provides the number of qualifier values per qualifier
// as well as the thresholds used when computing
type QualifierCountsOutput struct {
	Thresholds map[string]int32 `json:"thresholds"`
	Counts     map[string]int32 `json:"counts"`
}

// QualifierListsOutput provides a mapping of qualifiers to a list of all its values
type QualifierListsOutput map[string][]string

// PipelineResultsOutput represents the pipeline results file
type PipelineResultsOutput struct {
	OutputAggValues []interface{} `json:"output_agg_values,omitempty"`
	DataInfo        interface{}   `json:"data_info"`
}

// Timestamps holds input for bulk-regional-data
type Timestamps struct {
	Timestamps    []string `json:"timestamps"`
	AllTimestamps []string `json:"all_timestamps"`
}

// ModelOutputBulkAggregateRegionalAdmins holds all bulk and aggregate regional data
type ModelOutputBulkAggregateRegionalAdmins struct {
	ModelOutputBulkRegionalAdmins *[]ModelOutputBulkRegionalAdmins `json:"regional_data"`
	SelectAgg                     *ModelOutputRegionalAdmins       `json:"select_agg"`
	AllAgg                        *ModelOutputRegionalAdmins       `json:"all_agg"`
}

// ModelOutputBulkRegionalAdmins associates a timestamp for regional data
type ModelOutputBulkRegionalAdmins struct {
	Timestamp                  string `json:"timestamp"`
	*ModelOutputRegionalAdmins `json:"data"`
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

// Transform is type for available transforms
type Transform string

// Available transforms
const (
	TransformPerCapita     Transform = "percapita"
	TransformPerCapita1K   Transform = "percapita1k"
	TransformPerCapita1M   Transform = "percapita1m"
	TransformNormalization Transform = "normalization"
)

// TransformConfig defines transform configuration
type TransformConfig struct {
	Transform   Transform `json:"transform"`
	RegionID    string    `json:"region_id"`
	ScaleFactor float64   `json:"scale_factor"`
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

	// GetOutputExtrema returns extrema json
	GetOutputExtrema(params DatacubeParams) (*Extrema, error)

	// GetOutputSparkline returns datacube output sparkline
	GetOutputSparkline(params DatacubeParams, rawRes TemporalResolution, rawLatestTimestamp int64) ([]float64, error)

	// GetOutputTimeseriesByRegion returns timeseries data for a specific region
	GetOutputTimeseriesByRegion(params DatacubeParams, regionID string) ([]*TimeseriesValue, error)

	// GetRegionAggregation returns regional data for ALL admin regions at ONE timestamp
	GetRegionAggregation(params DatacubeParams, timestamp string) (*ModelOutputRegionalAdmins, error)

	// GetRawData returns datacube raw data
	GetRawData(params DatacubeParams) ([]*ModelOutputRawDataPoint, error)

	// GetRegionLists returns region hierarchies in list form
	GetRegionLists(params RegionListParams) (*RegionListOutput, error)

	// GetQualifierCounts returns region hierarchy output
	GetQualifierCounts(params QualifierInfoParams) (*QualifierCountsOutput, error)

	// GetQualifierLists returns region hierarchy output
	GetQualifierLists(params QualifierInfoParams, qualifiers []string) (*QualifierListsOutput, error)

	// GetPipelineResults returns the pipeline results file
	GetPipelineResults(params PipelineResultsParams) (*PipelineResultsOutput, error)

	// GetQualifierTimeseries returns datacube output timeseries broken down by qualifiers
	GetQualifierTimeseries(params DatacubeParams, qualifier string, qualifierOptions []string) ([]*ModelOutputQualifierTimeseries, error)

	// GetQualifierTimeseriesByRegion returns datacube output timeseries broken down by qualifiers for a specific region
	GetQualifierTimeseriesByRegion(params DatacubeParams, qualifier string, qualifierOptions []string, regionID string) ([]*ModelOutputQualifierTimeseries, error)

	// GetQualifierData returns datacube output data broken down by qualifiers for ONE timestamp
	GetQualifierData(params DatacubeParams, timestamp string, qualifiers []string) ([]*ModelOutputQualifierBreakdown, error)

	// GetQualifierRegional returns datacube output data broken down by qualifiers for ONE timestamp
	GetQualifierRegional(params DatacubeParams, timestamp string, qualifier string) (*ModelOutputRegionalQualifiers, error)

	// TransformOutputTimeseriesByRegion returns transformed timeseries data
	TransformOutputTimeseriesByRegion(timeseries []*TimeseriesValue, config TransformConfig) ([]*TimeseriesValue, error)

	// TransformRegionAggregation returns transformed regional data for ALL admin regions at ONE timestamp
	TransformRegionAggregation(data *ModelOutputRegionalAdmins, timestamp string, config TransformConfig) (*ModelOutputRegionalAdmins, error)

	// TransformOutputQualifierTimeseriesByRegion returns transformed qualifier timeseries data
	TransformOutputQualifierTimeseriesByRegion(data []*ModelOutputQualifierTimeseries, config TransformConfig) ([]*ModelOutputQualifierTimeseries, error)

	// TransformQualifierRegional returns transformed qualifier regional data for ALL admin regions at ONE timestamp
	TransformQualifierRegional(data *ModelOutputRegionalQualifiers, timestamp string, config TransformConfig) (*ModelOutputRegionalQualifiers, error)
}

// VectorTile defines methods that tile storage/database needs to satisfy
type VectorTile interface {
	GetVectorTile(zoom, x, y uint32, tilesetName string) ([]byte, error)
}
