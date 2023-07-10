package api

import (
	"net/http"

	"github.com/go-chi/render"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type outputStatsResponse struct {
	*wm.OutputStatWithZoom
}

// Render allows to satisfy the render.Renderer interface.
func (msr *outputStatsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelRegionalOutputStatsResponse struct {
	*wm.ModelRegionalOutputStat
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelRegionalOutputStatsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputRegionalTimeseries struct {
	*wm.ModelOutputRegionalTimeSeries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputRegionalTimeseries) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputKeyedTimeseries struct {
	*wm.ModelOutputKeyedTimeSeries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputKeyedTimeseries) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type timeseriesResultChan struct {
	result chan []*wm.TimeseriesValue
	err    chan error
}

type modelOutputTimeseriesValue struct {
	*wm.TimeseriesValue
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputTimeseriesValue) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputSparklineValue float64

// Render allows to satisfy the render.Renderer interface.
func (m modelOutputSparklineValue) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputBulkRegionalData struct {
	*wm.ModelOutputBulkAggregateRegionalAdmins
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputBulkRegionalData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type regionalDataResultChan struct {
	result chan *wm.ModelOutputRegionalAdmins
	err    chan error
}

type modelOutputRegionalData struct {
	*wm.ModelOutputRegionalAdmins
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputRegionalData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputRawDataPoint map[string]interface{}

// Render allows to satisfy the render.Renderer interface.
func (m *modelOutputRawDataPoint) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newModelOutputRawDataPoint creates and return a modelOutputRaDataPoint
func newModelOutputRawDataPoint(d *wm.ModelOutputRawDataPoint) *modelOutputRawDataPoint {
	data := modelOutputRawDataPoint{}
	data["timestamp"] = d.Timestamp
	data["country"] = d.Country
	data["admin1"] = d.Admin1
	data["admin2"] = d.Admin2
	data["admin3"] = d.Admin3
	data["lat"] = d.Lat
	data["lng"] = d.Lng
	data["value"] = d.Value
	// Add qualifiers
	for k, v := range d.Qualifiers {
		data[k] = v
	}
	return &data
}

type modelOutputQualifierTimeseriesResponse struct {
	*wm.ModelOutputQualifierTimeseries
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputQualifierTimeseriesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type modelOutputQualifierValuesResponse struct {
	*wm.ModelOutputQualifierBreakdown
}

// Render allows to satisfy the render.Renderer interface.
func (msr *modelOutputQualifierValuesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *api) getRegionalDataOutputStats(w http.ResponseWriter, r *http.Request) error {
	op := "api.getRegionalDataOutputStats"
	params := getDatacubeParams(r)
	stats, err := a.dataOutput.GetRegionalOutputStats(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.Render(w, r, &modelRegionalOutputStatsResponse{stats})
	return nil
}

func (a *api) getDataOutputStats(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputStats"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	stats, err := a.dataOutput.GetOutputStats(params, timestamp)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, stat := range stats {
		list = append(list, &outputStatsResponse{stat})
	}
	render.RenderList(w, r, list)
	return nil
}

func getTimeseriesParamsForRegions(r *http.Request) ([]*wm.FullTimeseriesParams, error) {
	regionIDs, err := getRegionIDsFromBody(r)
	if err != nil {
		return nil, err
	}
	params := getDatacubeParams(r)
	transform := getTransform(r)

	if len(regionIDs) == 0 {
		regionIDs = append(regionIDs, "")
	}
	timeseriesParams := make([]*wm.FullTimeseriesParams, len(regionIDs))
	for i := 0; i < len(regionIDs); i++ {
		timeseriesParams[i] = &wm.FullTimeseriesParams{
			DatacubeParams: params,
			RegionID:       regionIDs[i],
			Transform:      transform,
			Key:            regionIDs[i],
		}
	}
	return timeseriesParams, nil
}

func (a *api) getAggregateDataOutputTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getAggregateDataOutputTimeseries"
	agg := getAgg(r)
	timeseriesParams, err := getTimeseriesParamsForRegions(r)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	keyedTimeSeries, err := a.getBulkTimeseries(timeseriesParams)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}

	var aggTimeSeries []*wm.TimeseriesValue
	if agg == "mean" {
		timeToValue := map[int64]*[2]float64{}
		sumAggregationForTimeseries(keyedTimeSeries, timeToValue)
		aggTimeSeries = applyMeanForTimeseries(timeToValue)
	}

	list := []render.Renderer{}
	for _, timeseries := range aggTimeSeries {
		list = append(list, &modelOutputTimeseriesValue{timeseries})
	}
	render.RenderList(w, r, list)
	return nil
}

func sumAggregationForTimeseries(keyedTimeSeries []*wm.ModelOutputKeyedTimeSeries, aggDict map[int64]*[2]float64) {
	for _, region := range keyedTimeSeries {
		for _, timeseries := range region.Timeseries {
			val, ok := aggDict[timeseries.Timestamp]
			if ok {
				val[0] += timeseries.Value
				val[1]++
			} else {
				aggDict[timeseries.Timestamp] = &[2]float64{timeseries.Value, 1}
			}
		}
	}
}

func applyMeanForTimeseries(countryAgg map[int64]*[2]float64) []*wm.TimeseriesValue {
	aggregations := make([]*wm.TimeseriesValue, len(countryAgg))
	i := 0
	for key, val := range countryAgg {
		aggregations[i] = &wm.TimeseriesValue{Timestamp: key, Value: val[0] / val[1]}
		i++
	}
	return aggregations
}

func (a *api) getBulkDataOutputRegionTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getBulkDataOutputRegionTimeseries"

	timeseriesParams, err := getTimeseriesParamsForRegions(r)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	keyedTimeSeries, err := a.getBulkTimeseries(timeseriesParams)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}

	list := []render.Renderer{}
	for _, timeseries := range keyedTimeSeries {
		list = append(list, &modelOutputRegionalTimeseries{&wm.ModelOutputRegionalTimeSeries{
			RegionID:   timeseries.Key,
			Timeseries: timeseries.Timeseries,
		}})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getBulkDataOutputGenericTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getBulkDataOutputGenericTimeseries"
	timeseriesParams, err := getTimeseriesParamsFromBody(r)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}

	keyedTimeSeries, err := a.getBulkTimeseries(timeseriesParams)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}

	list := []render.Renderer{}
	for _, timeseries := range keyedTimeSeries {
		list = append(list, &modelOutputKeyedTimeseries{timeseries})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getBulkTimeseries(timeseriesParams []*wm.FullTimeseriesParams) ([]*wm.ModelOutputKeyedTimeSeries, error) {
	keyedTimeSeries := make([]*wm.ModelOutputKeyedTimeSeries, len(timeseriesParams))
	resultChannels := make(map[string]timeseriesResultChan)
	for i := 0; i < len(keyedTimeSeries); i++ {
		params := timeseriesParams[i]
		rc := a.getTimeSeriesAsync(params.RegionID, params.DatacubeParams, params.Transform)
		resultChannels[params.Key] = rc
	}

	for i := 0; i < len(keyedTimeSeries); i++ {
		key := timeseriesParams[i].Key
		timeseries := <-resultChannels[key].result
		err := <-resultChannels[key].err
		if err != nil {
			if wm.ErrorCode(err) == wm.ENOTFOUND {
				keyedTimeSeries[i] = &wm.ModelOutputKeyedTimeSeries{
					Key:        key,
					Timeseries: []*wm.TimeseriesValue{},
				}
				continue
			}
			return nil, err
		}
		keyedTimeSeries[i] = &wm.ModelOutputKeyedTimeSeries{
			Key:        key,
			Timeseries: timeseries,
		}
	}

	return keyedTimeSeries, nil
}

func (a *api) getDataOutputTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputTimeseries"
	params := getDatacubeParams(r)
	regionID := getRegionID(r)
	transform := getTransform(r)
	var timeseries []*wm.TimeseriesValue
	var err error

	timeseries, err = a.getTimeSeries(regionID, params, transform)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, point := range timeseries {
		list = append(list, &modelOutputTimeseriesValue{point})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getTimeSeriesAsync(regionID string, params wm.DatacubeParams, transform wm.Transform) timeseriesResultChan {
	rc := make(chan []*wm.TimeseriesValue)
	ec := make(chan error)
	go func() {
		r, err := a.getTimeSeries(regionID, params, transform)
		rc <- r
		ec <- err
	}()
	return timeseriesResultChan{
		result: rc,
		err:    ec,
	}
}

func (a *api) getTimeSeries(regionID string, params wm.DatacubeParams, transform wm.Transform) ([]*wm.TimeseriesValue, error) {
	var timeseries []*wm.TimeseriesValue
	var err error

	if regionID == "" {
		timeseries, err = a.dataOutput.GetOutputTimeseries(params)
		if err != nil {
			return nil, err
		}
		return timeseries, nil
	}
	timeseries, err = a.dataOutput.GetOutputTimeseriesByRegion(params, regionID)
	if err != nil {
		return nil, err
	}
	if transform != "" {
		timeseries, err = a.dataOutput.TransformOutputTimeseriesByRegion(timeseries, wm.TransformConfig{Transform: transform, RegionID: regionID, DatacubeParams: &params})
		if err != nil {
			return nil, err
		}
	}

	return timeseries, nil
}

func (a *api) getDataOutputSparkline(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputSparkline"
	params := getDatacubeParams(r)
	rawRes := getRawDataResolution(r)
	rawLastTs, err := getRawDataLatestTimestamp(r)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}

	var sparkline []float64
	sparkline, err = a.dataOutput.GetOutputSparkline(params, wm.TemporalResolution(rawRes), rawLastTs)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, point := range sparkline {
		list = append(list, modelOutputSparklineValue(point))
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getBulkDataOutputRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getBulkDataOutputRegional"
	timestamps, err := getTimestampsFromBody(r)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	params := getDatacubeParams(r)
	transform := getTransform(r)
	aggForSelect := getAggForSelect(r)
	aggForAll := getAggForAll(r)
	bulkData := wm.ModelOutputBulkAggregateRegionalAdmins{}

	bulkRegionalData := make([]wm.ModelOutputBulkRegionalAdmins, len(timestamps.Timestamps))
	var totalBulkRegionalData []wm.ModelOutputBulkRegionalAdmins

	selectResultChannels := make(map[string]regionalDataResultChan)
	restResultChannels := make(map[string]regionalDataResultChan)

	for _, timestamp := range timestamps.Timestamps {
		rc := a.getRegionAggregationAsync(params, timestamp, transform)
		selectResultChannels[timestamp] = rc
	}

	// Only timestamps not in timestamps.Timestamps
	for _, timestamp := range timestamps.AllTimestamps {
		_, ok := selectResultChannels[timestamp]
		if !ok {
			rc := a.getRegionAggregationAsync(params, timestamp, transform)
			restResultChannels[timestamp] = rc
		}
	}

	for i, timestamp := range timestamps.Timestamps {
		data := <-selectResultChannels[timestamp].result
		err := <-selectResultChannels[timestamp].err
		if err != nil {
			if wm.ErrorCode(err) == wm.ENOTFOUND {
				bulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
					Timestamp: timestamp,
					ModelOutputRegionalAdmins: &wm.ModelOutputRegionalAdmins{
						Country: []wm.ModelOutputAdminData{},
						Admin1:  []wm.ModelOutputAdminData{},
						Admin2:  []wm.ModelOutputAdminData{},
						Admin3:  []wm.ModelOutputAdminData{},
					},
				}
				continue
			}
			return &wm.Error{Op: op, Err: err}
		}
		bulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
			Timestamp:                 timestamp,
			ModelOutputRegionalAdmins: data,
		}
	}
	bulkData.ModelOutputBulkRegionalAdmins = &bulkRegionalData
	if len(timestamps.AllTimestamps) != 0 {
		// avoid looking through subset of timestamps an unnecessary amount of times
		subsetTimeStamps := map[string]wm.ModelOutputBulkRegionalAdmins{}
		for _, regionalData := range bulkRegionalData {
			subsetTimeStamps[regionalData.Timestamp] = regionalData
		}
		// hold only data that wouldn't already be in bulkRegionalData
		totalBulkRegionalData = make([]wm.ModelOutputBulkRegionalAdmins, len(timestamps.AllTimestamps))
		for i, timestamp := range timestamps.AllTimestamps {
			channels, inRest := restResultChannels[timestamp]
			val, inSubset := subsetTimeStamps[timestamp]
			if inRest {
				data := <-channels.result
				err := <-channels.err
				if err != nil {
					if wm.ErrorCode(err) == wm.ENOTFOUND {
						totalBulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
							Timestamp: timestamp,
							ModelOutputRegionalAdmins: &wm.ModelOutputRegionalAdmins{
								Country: []wm.ModelOutputAdminData{},
								Admin1:  []wm.ModelOutputAdminData{},
								Admin2:  []wm.ModelOutputAdminData{},
								Admin3:  []wm.ModelOutputAdminData{},
							},
						}
						continue
					}
					return &wm.Error{Op: op, Err: err}
				}
				totalBulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
					Timestamp:                 timestamp,
					ModelOutputRegionalAdmins: data,
				}
			} else if inSubset {
				// copy it over to apply whatever aggregation we will need to
				totalBulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
					Timestamp:                 val.Timestamp,
					ModelOutputRegionalAdmins: val.ModelOutputRegionalAdmins,
				}
			} else {
				// should never happen
				totalBulkRegionalData[i] = wm.ModelOutputBulkRegionalAdmins{
					Timestamp: timestamp,
					ModelOutputRegionalAdmins: &wm.ModelOutputRegionalAdmins{
						Country: []wm.ModelOutputAdminData{},
						Admin1:  []wm.ModelOutputAdminData{},
						Admin2:  []wm.ModelOutputAdminData{},
						Admin3:  []wm.ModelOutputAdminData{},
					},
				}
			}
		}
	}

	if aggForSelect == "mean" {
		bulkData.SelectAgg = getMeanForBulkRegionalData(bulkRegionalData)
	}

	if aggForAll == "mean" && len(timestamps.AllTimestamps) != 0 {
		bulkData.AllAgg = getMeanForBulkRegionalData(totalBulkRegionalData)
	}

	render.Render(w, r, &modelOutputBulkRegionalData{&bulkData})
	return nil
}

func getMeanForBulkRegionalData(bulkRegionalData []wm.ModelOutputBulkRegionalAdmins) *wm.ModelOutputRegionalAdmins {
	aggData := wm.ModelOutputRegionalAdmins{}
	countryAgg := map[string]*[2]float64{}
	admin1Agg := map[string]*[2]float64{}
	admin2Agg := map[string]*[2]float64{}
	admin3Agg := map[string]*[2]float64{}
	for _, regionalData := range bulkRegionalData {
		sumAggregation(regionalData.Country, countryAgg)
		sumAggregation(regionalData.Admin1, admin1Agg)
		sumAggregation(regionalData.Admin2, admin2Agg)
		sumAggregation(regionalData.Admin3, admin3Agg)
	}
	aggData.Country = applyMeanAggregation(countryAgg)
	aggData.Admin1 = applyMeanAggregation(admin1Agg)
	aggData.Admin2 = applyMeanAggregation(admin2Agg)
	aggData.Admin3 = applyMeanAggregation(admin3Agg)
	return &aggData
}

func sumAggregation(regionProperty []wm.ModelOutputAdminData, aggDict map[string]*[2]float64) {
	for _, property := range regionProperty {
		val, ok := aggDict[property.ID]
		if ok {
			val[0] += property.Value
			val[1]++
		} else {
			aggDict[property.ID] = &[2]float64{property.Value, 1}
		}
	}
}

func applyMeanAggregation(countryAgg map[string]*[2]float64) []wm.ModelOutputAdminData {
	aggregations := make([]wm.ModelOutputAdminData, len(countryAgg))
	i := 0
	for key, val := range countryAgg {
		aggregations[i] = wm.ModelOutputAdminData{ID: key, Value: val[0] / val[1]}
		i++
	}
	return aggregations
}

func (a *api) getDataOutputRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRegional"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	transform := getTransform(r)
	data, err := a.getRegionAggregation(params, timestamp, transform)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.Render(w, r, &modelOutputRegionalData{data})
	return nil
}

func (a *api) getRegionAggregationByAdminLevel(w http.ResponseWriter, r *http.Request) error {
	op := "api.getRegionAggregationByAdminLevel"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	transform := getTransform(r)
	adminLevel := getAdminLevel(r)

	data, err := a.dataOutput.GetRegionAggregationByAdminLevel(params, timestamp, adminLevel)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	if transform != "" {
		data, err = a.dataOutput.TransformRegionAggregationByAdminLevel(data, wm.TransformConfig{Transform: transform, DatacubeParams: &params})
		if err != nil {
			return &wm.Error{Op: op, Err: err}
		}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getRegionAggregationAsync(params wm.DatacubeParams, timestamp string, transform wm.Transform) regionalDataResultChan {
	rc := make(chan *wm.ModelOutputRegionalAdmins)
	ec := make(chan error)
	go func() {
		r, err := a.getRegionAggregation(params, timestamp, transform)
		rc <- r
		ec <- err
	}()
	return regionalDataResultChan{
		result: rc,
		err:    ec,
	}
}

func (a *api) getRegionAggregation(params wm.DatacubeParams, timestamp string, transform wm.Transform) (*wm.ModelOutputRegionalAdmins, error) {
	data, err := a.dataOutput.GetRegionAggregation(params, timestamp)
	if err != nil {
		return nil, err
	}
	if transform != "" {
		data, err = a.dataOutput.TransformRegionAggregation(data, timestamp, wm.TransformConfig{Transform: transform})
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (a *api) getDataOutputRaw(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRaw"
	params := getDatacubeParams(r)
	data, err := a.dataOutput.GetRawData(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, point := range data {
		list = append(list, newModelOutputRawDataPoint(point))
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputRegionLists(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputRegionLists"
	params := getRegionListsParams(r)
	data, err := a.dataOutput.GetRegionLists(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getDataOutputQualifierCounts(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierCounts"
	params := getQualifierInfoParams(r)
	data, err := a.dataOutput.GetQualifierCounts(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getDataOutputQualifierLists(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierLists"
	params := getQualifierInfoParams(r)
	qualifiers := getQualifierNames(r)
	data, err := a.dataOutput.GetQualifierLists(params, qualifiers)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, &data)
	return nil
}

func (a *api) getDataOutputQualifierTimeseries(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierTimeseries"
	params := getDatacubeParams(r)
	regionID := getRegionID(r)
	qualifier := getQualifierName(r)
	qualifierOptions := getQualifierOptions(r)
	transform := getTransform(r)

	var data []*wm.ModelOutputQualifierTimeseries
	var err error
	if regionID == "" {
		data, err = a.dataOutput.GetQualifierTimeseries(params, qualifier, qualifierOptions)
	} else {
		data, err = a.dataOutput.GetQualifierTimeseriesByRegion(params, qualifier, qualifierOptions, regionID)
		if transform != "" {
			data, err = a.dataOutput.TransformOutputQualifierTimeseriesByRegion(data, wm.TransformConfig{Transform: transform, RegionID: regionID})
			if err != nil {
				return &wm.Error{Op: op, Err: err}
			}
		}
	}

	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, timeseries := range data {
		list = append(list, &modelOutputQualifierTimeseriesResponse{timeseries})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputQualifierData(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierData"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	qualifiers := getQualifierNames(r)
	data, err := a.dataOutput.GetQualifierData(params, timestamp, qualifiers)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	list := []render.Renderer{}
	for _, value := range data {
		list = append(list, &modelOutputQualifierValuesResponse{value})
	}
	render.RenderList(w, r, list)
	return nil
}

func (a *api) getDataOutputQualifierRegional(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputQualifierRegional"
	params := getDatacubeParams(r)
	timestamp := getTimestamp(r)
	qualifier := getQualifierName(r)
	transform := getTransform(r)
	data, err := a.dataOutput.GetQualifierRegional(params, timestamp, qualifier)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	if transform != "" {
		data, err = a.dataOutput.TransformQualifierRegional(data, timestamp, wm.TransformConfig{Transform: transform})
		if err != nil {
			return &wm.Error{Op: op, Err: err}
		}
	}
	render.JSON(w, r, data)
	return nil
}

func (a *api) getDataOutputPipelineResults(w http.ResponseWriter, r *http.Request) error {
	op := "api.getDataOutputPipelineResults"
	params := getPipelineResultParams(r)
	data, err := a.dataOutput.GetPipelineResults(params)
	if err != nil {
		return &wm.Error{Op: op, Err: err}
	}
	render.JSON(w, r, data)
	return nil
}
