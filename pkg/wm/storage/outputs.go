package storage

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/mapstructure"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

type s3ResultChan struct {
	result chan []byte
	err    chan error
}

func getRegionLevels() []string {
	return []string{"country", "admin1", "admin2", "admin3"}
}

// Get the s3 bucket based on the output runID
func getBucket(s *Storage, outputRunID string) string {
	bucket := s.bucketInfo.ModelsBucket
	if outputRunID == "indicator" {
		bucket = s.bucketInfo.IndicatorsBucket
	}
	return bucket
}

// Get the index of provided string in given string slice
func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// GetOutputExtrema - gets min and max statistics
func (s *Storage) GetOutputExtrema(params wm.DatacubeParams) (*wm.RegionalExtremaSelected, error) {
	op := "Storage.GetOutputExtrema"
	key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/stats/default/extrema.json",
		params.DataID, params.RunID, params.Resolution, params.Feature, params.AdminLevel)
	bucket := getBucket(s, params.RunID)
	buf, err := getFileFromS3(s, bucket, aws.String(key))

	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	var output wm.RegionalExtrema
	err = json.Unmarshal(buf, &output)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	// remove unneeded data based on agg parameters from user
	var filter = fmt.Sprintf(`s_%s_t_%s`, params.SpatialAggFunc, params.TemporalAggFunc)
	filteredOutput := make(wm.RegionalExtremaSelected)

	for key := range output.Min {
		if key == filter {
			filteredOutput["min"] = output.Min[key]
		}
	}

	for key := range output.Max {
		if key == filter {
			filteredOutput["max"] = output.Max[key]
		}
	}

	return &filteredOutput, nil
}

// GetRegionalOutputStats returns regional output statistics
func (s *Storage) GetRegionalOutputStats(params wm.DatacubeParams) (*wm.ModelRegionalOutputStat, error) {
	op := "Storage.GetRegionalOutputStats"
	regionMap := make(map[string]*wm.ModelOutputStat)
	for _, level := range getRegionLevels() {
		var regionKey = fmt.Sprintf("regional/%s", level)
		stats, err := s.getOutputStats(params, regionKey)
		if err == nil {
			regionMap[level] = stats
		}
	}
	var statsByRegion wm.ModelRegionalOutputStat
	err := mapstructure.Decode(regionMap, &statsByRegion)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &statsByRegion, nil
}

// GetOutputStats returns datacube output stats
func (s *Storage) getOutputStats(params wm.DatacubeParams, filename string) (*wm.ModelOutputStat, error) {
	op := "Storage.getOutputStats"
	key := fmt.Sprintf("%s/%s/%s/%s/stats/%s.json",
		params.DataID, params.RunID, params.Resolution, params.Feature, filename)

	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	//The format is {"0":{ <stats> }}
	statsAt0 := make(map[string]map[string]float64)
	err = json.Unmarshal(buf, &statsAt0)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	if len(statsAt0) == 0 {
		return nil, &wm.Error{Code: wm.ENOTFOUND, Op: op, Message: "No stats found"}
	}

	//Take the first item from `statsAt0`
	var stats wm.ModelOutputStat
	for _, val := range statsAt0 {
		minKey := fmt.Sprintf("min_s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
		maxKey := fmt.Sprintf("max_s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
		stats.Min = val[minKey]
		stats.Max = val[maxKey]
		break
	}

	return &stats, nil
}

// GetOutputStats returns stats for grid output data
func (s *Storage) GetOutputStats(params wm.DatacubeParams, timestamp string) ([]*wm.OutputStatWithZoom, error) {
	op := "Storage.GetOutputStats"
	key := fmt.Sprintf("%s/%s/%s/%s/stats/grid/%s.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature, timestamp)

	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	// Read and parse csv
	r := csv.NewReader(bytes.NewReader(buf))
	records, err := r.ReadAll()
	if err != nil || len(records) == 0 {
		return nil, &wm.Error{Op: op, Err: fmt.Errorf("Invalid output stats file. s3_key: %s", key)}
	}
	minCol := fmt.Sprintf("min_s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
	maxCol := fmt.Sprintf("max_s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
	minColIndex := index(records[0], minCol)
	maxColIndex := index(records[0], maxCol)
	if minColIndex == -1 || maxColIndex == -1 {
		return nil, &wm.Error{Code: wm.EINVALID, Op: op, Message: fmt.Sprintf("Invalid agg functions. Spatial: %s, Temporal: %s", params.SpatialAggFunc, params.TemporalAggFunc)}
	}

	stats := make([]*wm.OutputStatWithZoom, 0)
	for i := 1; i < len(records); i++ {
		record := records[i]
		zoom, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		min, err := strconv.ParseFloat(record[minColIndex], 64)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		max, err := strconv.ParseFloat(record[maxColIndex], 64)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		stats = append(stats, &wm.OutputStatWithZoom{
			Zoom: uint8(zoom),
			Min:  min,
			Max:  max,
		})
	}
	return stats, nil
}

// GetOutputTimeseries returns datacube output timeseries
func (s *Storage) GetOutputTimeseries(params wm.DatacubeParams) ([]*wm.TimeseriesValue, error) {
	// op := "Storage.GetOutputTimeseries"
	key := fmt.Sprintf("%s/%s/%s/%s/timeseries/global/global.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature)

	return getTimeseriesFromCsv(s, key, params)
}

// GetOutputSparkline returns a datacube output sparkline
// If rawRes and rawLatestTimestamp is provided, try correcting incomplete last value
func (s *Storage) GetOutputSparkline(params wm.DatacubeParams, rawRes wm.TemporalResolution, rawLatestTimestamp int64) ([]float64, error) {
	op := "Storage.GetOutputSparkline"
	series, err := s.GetOutputTimeseries(params)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	if rawRes != "" && rawLatestTimestamp != 0 {
		series = correctIncompleteTimeseries(series, params.TemporalAggFunc, params.Resolution, rawRes, rawLatestTimestamp)
	}
	return toSparkline(series)
}

// GetOutputTimeseriesByRegion returns timeseries data for a specific region
func (s *Storage) GetOutputTimeseriesByRegion(params wm.DatacubeParams, regionID string) ([]*wm.TimeseriesValue, error) {
	// op := "Storage.GetOutputTimeseriesByRegion"
	// Deconstruct Region ID to get admin region levels
	regions := strings.Split(regionID, "__")
	regionLevel := getRegionLevels()[len(regions)-1]
	key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/timeseries/default/%s.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature, regionLevel, regionID)

	return getTimeseriesFromCsv(s, key, params)
}

// GetRegionAggregationByAdminLevel returns regional data for given admin level at ONE timestamp
func (s *Storage) GetRegionAggregationByAdminLevel(params wm.DatacubeParams, timestamp string, adminLevel wm.AdminLevel) (*wm.ModelOutputRegional, error) {
	op := "Storage.GetRegionAggregationByAdminLevel"
	key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/default/default.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature, adminLevel, timestamp)

	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	result := make(wm.ModelOutputRegional)
	points := make([]wm.ModelOutputAdminData, 0)

	// Read and parse csv
	r := csv.NewReader(bytes.NewReader(buf))
	isHeader := true
	valueCol := fmt.Sprintf("s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
	valueColIndex := -1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		// Parse header and find the index of the target value column
		if isHeader {
			valueColIndex = index(record, valueCol)
			if valueColIndex == -1 {
				return nil, &wm.Error{Code: wm.EINVALID, Op: op, Message: fmt.Sprintf("Invalid agg functions. Spatial: %s, Temporal: %s", params.SpatialAggFunc, params.TemporalAggFunc)}
			}
			isHeader = false
		} else {
			regionID := record[0]
			value, err := strconv.ParseFloat(record[valueColIndex], 64)
			if err != nil {
				continue
			}
			points = append(points, wm.ModelOutputAdminData{
				ID:    regionID,
				Value: value,
			})
		}
	}
	result[adminLevel] = points

	return &result, nil
}

// GetRegionAggregation returns regional data for ALL admin regions at ONE timestamp
func (s *Storage) GetRegionAggregation(params wm.DatacubeParams, timestamp string) (*wm.ModelOutputRegionalAdmins, error) {
	op := "Storage.GetRegionAggregation"

	data := make(map[string][]wm.ModelOutputAdminData)
	resultChannels := make(map[string]s3ResultChan)

	for _, level := range getRegionLevels() {
		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/default/default.csv",
			params.DataID, params.RunID, params.Resolution, params.Feature, level, timestamp)
		rc := getFileFromS3Async(s, getBucket(s, params.RunID), aws.String(key))
		resultChannels[level] = rc
	}
	for _, level := range getRegionLevels() {
		buf := <-resultChannels[level].result
		err := <-resultChannels[level].err
		if err != nil {
			if wm.ErrorCode(err) == wm.ENOTFOUND {
				data[level] = make([]wm.ModelOutputAdminData, 0)
				continue
			} else {
				return nil, &wm.Error{Op: op, Err: err}
			}
		}

		points := make([]wm.ModelOutputAdminData, 0)
		// Read and parse csv
		r := csv.NewReader(bytes.NewReader(buf))
		isHeader := true
		valueCol := fmt.Sprintf("s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
		valueColIndex := -1
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, &wm.Error{Op: op, Err: err}
			}
			// Parse header and find the index of the target value column
			if isHeader {
				valueColIndex = index(record, valueCol)
				if valueColIndex == -1 {
					return nil, &wm.Error{Code: wm.EINVALID, Op: op, Message: fmt.Sprintf("Invalid agg functions. Spatial: %s, Temporal: %s", params.SpatialAggFunc, params.TemporalAggFunc)}
				}
				isHeader = false
			} else {
				regionID := record[0]
				value, err := strconv.ParseFloat(record[valueColIndex], 64)
				if err != nil {
					continue
				}
				points = append(points, wm.ModelOutputAdminData{
					ID:    regionID,
					Value: value,
				})
			}
		}

		data[level] = points
	}

	var regionalData wm.ModelOutputRegionalAdmins
	err := mapstructure.Decode(data, &regionalData)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &regionalData, nil
}

// GetRegionLists returns region hierarchies in list form
func (s *Storage) GetRegionLists(params wm.RegionListParams) (*wm.RegionListOutput, error) {
	op := "Storage.GetRegionLists"
	var regionalData wm.RegionListOutput
	// allOutputMap is meant to be a map from strings ie. 'country' to a set (map[string]bool is used as a set)
	allOutputMap := make(map[string]map[string]bool)
	for _, region := range getRegionLevels() {
		allOutputMap[region] = make(map[string]bool)
	}
	resultChannels := make(map[string]s3ResultChan)
	for _, runID := range params.RunIDs {
		key := fmt.Sprintf("%s/%s/raw/%s/info/region_lists.json", params.DataID, runID, params.Feature)
		rc := getFileFromS3Async(s, getBucket(s, runID), aws.String(key))
		resultChannels[runID] = rc
	}
	// Populate sets in allOutputMap map values with regions
	for _, runID := range params.RunIDs {
		buf := <-resultChannels[runID].result
		err := <-resultChannels[runID].err
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		outputMap := make(map[string][]string)
		err = json.Unmarshal(buf, &outputMap)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		for adminLevel, regions := range outputMap {
			for _, region := range regions {
				allOutputMap[adminLevel][region] = true
			}
		}
	}
	// Get a map of strings to lists from allOutputMap (instead of strings to sets)
	regionListMap := make(map[string][]string)
	for adminLevel, regions := range allOutputMap {
		keys := make([]string, len(regions))
		i := 0
		for k := range regions {
			keys[i] = k
			i++
		}
		regionListMap[adminLevel] = keys
	}
	err := mapstructure.Decode(regionListMap, &regionalData)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &regionalData, nil
}

// GetQualifierCounts returns the number of qualifier values per qualifier
func (s *Storage) GetQualifierCounts(params wm.QualifierInfoParams) (*wm.QualifierCountsOutput, error) {
	op := "Storage.GetQualifierCounts"
	key := fmt.Sprintf("%s/%s/raw/%s/info/qualifier_counts.json",
		params.DataID, params.RunID, params.Feature)
	bucket := getBucket(s, params.RunID)
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	var output wm.QualifierCountsOutput
	err = json.Unmarshal(buf, &output)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &output, nil
}

// GetQualifierLists returns the number of qualifier values per qualifier
func (s *Storage) GetQualifierLists(params wm.QualifierInfoParams, qualifiers []string) (*wm.QualifierListsOutput, error) {
	op := "Storage.GetQualifierLists"
	bucket := getBucket(s, params.RunID)

	resultChannels := make(map[string]s3ResultChan)
	outputLists := make(map[string][]string)
	for _, qualifier := range qualifiers {
		key := fmt.Sprintf("%s/%s/raw/%s/info/qualifiers/%s.json",
			params.DataID, params.RunID, params.Feature, qualifier)
		rc := getFileFromS3Async(s, bucket, aws.String(key))
		resultChannels[qualifier] = rc
	}
	for _, qualifier := range qualifiers {
		buf := <-resultChannels[qualifier].result
		err := <-resultChannels[qualifier].err
		if err != nil {
			if wm.ErrorCode(err) != wm.ENOTFOUND {
				return nil, &wm.Error{Op: op, Err: err}
			}
			// If there was no data, return []
			outputLists[qualifier] = []string{}
			continue
		}
		var output []string
		err = json.Unmarshal(buf, &output)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		outputLists[qualifier] = output
	}

	var qualifierLists wm.QualifierListsOutput
	err := mapstructure.Decode(outputLists, &qualifierLists)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &qualifierLists, nil
}

// GetPipelineResults returns the pipeline results file
func (s *Storage) GetPipelineResults(params wm.PipelineResultsParams) (*wm.PipelineResultsOutput, error) {
	op := "Storage.GetPipelineResults"
	key := fmt.Sprintf("%s/%s/results/results.json", params.DataID, params.RunID)
	bucket := getBucket(s, params.RunID)
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	var output wm.PipelineResultsOutput
	err = json.Unmarshal(buf, &output)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &output, nil
}

// GetRawData returns datacube output or indicator raw data
func (s *Storage) GetRawData(params wm.DatacubeParams) ([]*wm.ModelOutputRawDataPoint, error) {
	op := "Storage.GetRawData"
	key := fmt.Sprintf("%s/%s/raw/%s/raw/raw.csv",
		params.DataID, params.RunID, params.Feature)

	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	data := make([]*wm.ModelOutputRawDataPoint, 0)
	requiredColSet := map[string]bool{"timestamp": true, "country": true, "admin1": true, "admin2": true, "admin3": true, "lat": true, "lng": true, "value": true}
	requiredColsIndexMap := make(map[string]int)
	qualifierColsIndexMap := make(map[string]int)

	// Parse raw csv data
	r := csv.NewReader(bytes.NewReader(buf))
	isHeader := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		// Parse header and find the index of each column
		if isHeader {
			for i, v := range record {
				if requiredColSet[v] {
					requiredColsIndexMap[v] = i
				} else {
					qualifierColsIndexMap[v] = i
				}
			}
			isHeader = false
		} else {
			dataPoint := &wm.ModelOutputRawDataPoint{}
			if index, ok := requiredColsIndexMap["timestamp"]; ok {
				timstamp, err := strconv.ParseInt(record[index], 10, 64)
				if err != nil {
					return nil, &wm.Error{Op: op, Err: err}
				}
				dataPoint.Timestamp = timstamp
			}
			if index, ok := requiredColsIndexMap["country"]; ok {
				dataPoint.Country = record[index]
			}
			if index, ok := requiredColsIndexMap["admin1"]; ok {
				dataPoint.Admin1 = record[index]
			}
			if index, ok := requiredColsIndexMap["admin2"]; ok {
				dataPoint.Admin2 = record[index]
			}
			if index, ok := requiredColsIndexMap["admin3"]; ok {
				dataPoint.Admin3 = record[index]
			}
			if index, ok := requiredColsIndexMap["lat"]; ok {
				v := record[index]
				if v != "" {
					lat, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return nil, &wm.Error{Op: op, Err: err}
					}
					dataPoint.Lat = &lat
				}
			}
			if index, ok := requiredColsIndexMap["lng"]; ok {
				v := record[index]
				if v != "" {
					lng, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return nil, &wm.Error{Op: op, Err: err}
					}
					dataPoint.Lng = &lng
				}
			}
			if index, ok := requiredColsIndexMap["value"]; ok {
				v := record[index]
				if v != "" {
					value, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return nil, &wm.Error{Op: op, Err: err}
					}
					dataPoint.Value = &value
				}
			}
			// Add extra columns
			qualifiers := make(map[string]string)
			for col, index := range qualifierColsIndexMap {
				qualifiers[col] = record[index]
			}
			dataPoint.Qualifiers = qualifiers

			data = append(data, dataPoint)
		}
	}
	return data, nil
}

// GetQualifierTimeseries returns datacube output timeseries broken down by qualifiers
func (s *Storage) GetQualifierTimeseries(params wm.DatacubeParams, qualifier string, qualifierOptions []string) ([]*wm.ModelOutputQualifierTimeseries, error) {
	op := "Storage.GetQualifierTimeseries"
	key := fmt.Sprintf("%s/%s/%s/%s/timeseries/qualifiers/%s/s_%s_t_%s.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature, qualifier,
		params.SpatialAggFunc, params.TemporalAggFunc)
	// Want to return one timeseries per column
	// CSV format:
	// timestamp,Battles,Protests,Riots,Strategic developments,Violence against civilians
	// 852076800000,1583.0,0.0,0.0,0.0,313.0
	// 854755200000,187.0,3.0,0.0,0.0,40.0

	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	// Read and parse csv
	var values []*wm.ModelOutputQualifierTimeseries
	r := csv.NewReader(bytes.NewReader(buf))
	isHeader := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		// Parse header and find the index of the target value column
		if isHeader {
			// The first column is timestamp, rest are values
			numValues := len(record) - 1
			if numValues < 1 {
				break
			}
			values = make([]*wm.ModelOutputQualifierTimeseries, numValues)
			for i := 0; i < numValues; i++ {
				values[i] = &wm.ModelOutputQualifierTimeseries{
					Name: record[i+1], Timeseries: make([]*wm.TimeseriesValue, 0)}
			}
			isHeader = false
		} else {
			timestamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				return nil, &wm.Error{Op: op, Err: err}
			}
			for i := 0; i+1 < len(record) && i < len(values); i++ {
				value, err := strconv.ParseFloat(record[i+1], 64)
				if err != nil {
					continue
				}
				values[i].Timeseries = append(values[i].Timeseries, &wm.TimeseriesValue{
					Timestamp: timestamp, Value: value})
			}
		}
	}

	// Filter the qualifier options, only include what's specified in qualifierOptions
	filteredValues := make([]*wm.ModelOutputQualifierTimeseries, 0)
	for _, timeseries := range values {
		if timeseries != nil {
			for _, name := range qualifierOptions {
				if timeseries.Name == name {
					sortTimeseries(timeseries.Timeseries)
					filteredValues = append(filteredValues, timeseries)
					break
				}
			}
		}
	}
	return filteredValues, nil
}

// GetQualifierTimeseriesByRegion returns datacube output timeseries broken down by qualifiers for a specific region
func (s *Storage) GetQualifierTimeseriesByRegion(params wm.DatacubeParams, qualifier string, qualifierOptions []string, regionID string) ([]*wm.ModelOutputQualifierTimeseries, error) {
	op := "Storage.GetQualifierTimeseriesByRegion"
	// Deconstruct Region ID to get admin region levels
	regions := strings.Split(regionID, "__")
	regionLevel := getRegionLevels()[len(regions)-1]

	type resultChan struct {
		result chan []*wm.TimeseriesValue
		err    chan error
	}
	outputTimeseries := make([]*wm.ModelOutputQualifierTimeseries, 0)
	chanMap := make(map[string]resultChan)
	for _, qOpt := range qualifierOptions {
		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/timeseries/qualifiers/%s/%s/%s.csv",
			params.DataID, params.RunID, params.Resolution, params.Feature, regionLevel,
			qualifier, qOpt, regionID)

		chanMap[qOpt] = resultChan{result: make(chan []*wm.TimeseriesValue), err: make(chan error)}
		go func(qo string) {
			series, err := getTimeseriesFromCsv(s, key, params)
			chanMap[qo].result <- series
			chanMap[qo].err <- err
		}(qOpt)
	}
	for _, qOpt := range qualifierOptions {
		series := <-chanMap[qOpt].result
		err := <-chanMap[qOpt].err
		if err != nil {
			if wm.ErrorCode(err) != wm.ENOTFOUND {
				return nil, &wm.Error{Op: op, Err: err}
			}
			// If there was no data for this combination, return []
			series = []*wm.TimeseriesValue{}
		}
		outputTimeseries = append(outputTimeseries, &wm.ModelOutputQualifierTimeseries{
			Name: qOpt, Timeseries: series})
	}

	return outputTimeseries, nil
}

// GetQualifierData returns datacube output data broken down by qualifiers for ONE timestamp
func (s *Storage) GetQualifierData(params wm.DatacubeParams, timestamp string, qualifiers []string) ([]*wm.ModelOutputQualifierBreakdown, error) {
	op := "Storage.GetQualifierData"
	allQualifiers := make([]*wm.ModelOutputQualifierBreakdown, len(qualifiers))

	resultChannels := make(map[string]s3ResultChan)
	for _, qualifier := range qualifiers {
		key := fmt.Sprintf("%s/%s/%s/%s/timeseries/qualifiers/%s/s_%s_t_%s.csv",
			params.DataID, params.RunID, params.Resolution, params.Feature, qualifier,
			params.SpatialAggFunc, params.TemporalAggFunc)
		// Want to return one row
		// CSV format:
		// timestamp,Battles,Protests,Riots,Strategic developments,Violence against civilians
		// 852076800000,1583.0,0.0,0.0,0.0,313.0
		// 854755200000,187.0,3.0,0.0,0.0,40.0
		rc := getFileFromS3Async(s, getBucket(s, params.RunID), aws.String(key))
		resultChannels[qualifier] = rc
	}
	for qualifierIndex, qualifier := range qualifiers {
		buf := <-resultChannels[qualifier].result
		err := <-resultChannels[qualifier].err
		if err != nil {
			allQualifiers[qualifierIndex] = nil
			continue
		}

		// Read and parse csv
		var values []*wm.ModelOutputQualifierValue
		r := csv.NewReader(bytes.NewReader(buf))
		isHeader := true
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, &wm.Error{Op: op, Err: err}
			}
			// Parse header and find the index of the target value column
			if isHeader {
				// The first column is timestamp, rest are values
				numValues := len(record) - 1
				if numValues < 1 {
					break
				}
				values = make([]*wm.ModelOutputQualifierValue, numValues)
				for i := 0; i < numValues; i++ {
					values[i] = &wm.ModelOutputQualifierValue{Name: record[i+1]}
				}
				isHeader = false
			} else {
				if timestamp == record[0] {
					for i := 0; i+1 < len(record) && i < len(values); i++ {
						value, err := strconv.ParseFloat(record[i+1], 64)
						if err != nil {
							values[i].Value = nil //set missing values to nil
						} else {
							values[i].Value = &value
						}
					}
					break
				}
			}
		}
		allQualifiers[qualifierIndex] = &wm.ModelOutputQualifierBreakdown{Name: qualifier, Options: values}
	}
	return allQualifiers, nil
}

// GetQualifierRegional returns datacube output data broken down by qualifiers for ONE timestamp
func (s *Storage) GetQualifierRegional(params wm.DatacubeParams, timestamp string, qualifier string) (*wm.ModelOutputRegionalQualifiers, error) {
	op := "Storage.GetQualifierRegional"

	data := make(map[string][]*wm.ModelOutputRegionQualifierBreakdown)
	resultChannels := make(map[string]s3ResultChan)
	for _, level := range getRegionLevels() {
		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/qualifiers/%s.csv",
			params.DataID, params.RunID, params.Resolution, params.Feature, level, timestamp, qualifier)
		// Want to return all values from one column grouped by region
		// CSV format:
		// id,qualifier,s_sum_t_mean,s_mean_t_mean,s_sum_t_sum,s_mean_t_sum
		// Central African Republic__Bangui,Protests,0.0,0.0,0.0,0.0
		// Central African Republic__Bangui,Violence against civilians,0.0,0.0,0.0,0.0
		// Central African Republic__Ouham,Violence against civilians,10.0,10.0,10.0,10.0
		rc := getFileFromS3Async(s, getBucket(s, params.RunID), aws.String(key))
		resultChannels[level] = rc
	}
	for _, level := range getRegionLevels() {
		buf := <-resultChannels[level].result
		err := <-resultChannels[level].err
		if err != nil {
			if wm.ErrorCode(err) == wm.ENOTFOUND {
				data[level] = []*wm.ModelOutputRegionQualifierBreakdown{}
				continue
			} else {
				return nil, &wm.Error{Op: op, Err: err}
			}
		}

		// Read and parse csv
		regionMap := make(map[string]map[string]float64)
		r := csv.NewReader(bytes.NewReader(buf))
		colName := fmt.Sprintf("s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
		valueColIndex := -1
		isHeader := true
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, &wm.Error{Op: op, Err: err}
			}
			// Parse header and find the index of the target value column
			if isHeader {
				valueColIndex = index(record, colName)
				if valueColIndex == -1 {
					return nil, fmt.Errorf("csv: column, %s does not exist", colName)
				}
				isHeader = false
			} else {
				region := record[0]
				value, err := strconv.ParseFloat(record[valueColIndex], 64)
				if err != nil {
					continue
				}

				if regionValues, ok := regionMap[region]; ok {
					regionValues[record[1]] = value
				} else {
					regionMap[region] = map[string]float64{
						record[1]: value,
					}
				}
			}
		}

		var regions []*wm.ModelOutputRegionQualifierBreakdown
		for k, v := range regionMap {
			regions = append(regions, &wm.ModelOutputRegionQualifierBreakdown{
				ID:     k,
				Values: v,
			})
		}

		data[level] = regions
	}

	var regionalData wm.ModelOutputRegionalQualifiers
	err := mapstructure.Decode(data, &regionalData)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &regionalData, nil
}

func getTimeseriesFromCsv(s *Storage, key string, params wm.DatacubeParams) ([]*wm.TimeseriesValue, error) {
	op := "getTimeseriesFromCsv"
	buf, err := getFileFromS3(s, getBucket(s, params.RunID), aws.String(key))
	if err != nil {
		return nil, err
	}

	// Read and parse csv
	series := make([]*wm.TimeseriesValue, 0)
	r := csv.NewReader(bytes.NewReader(buf))
	isHeader := true
	valueCol := fmt.Sprintf("s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
	valueColIndex := -1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
		}
		// Parse header and find the index of the target value column
		if isHeader {
			valueColIndex = index(record, valueCol)
			if valueColIndex == -1 {
				return nil, &wm.Error{Code: wm.EINVALID, Op: op, Message: fmt.Sprintf("Invalid agg functions. Spatial: %s, Temporal: %s", params.SpatialAggFunc, params.TemporalAggFunc)}
			}
			isHeader = false
		} else {
			timestamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				return nil, &wm.Error{Op: op, Err: err}
			}
			value, err := strconv.ParseFloat(record[valueColIndex], 64)
			if err != nil {
				continue
			}
			series = append(series, &wm.TimeseriesValue{
				Timestamp: timestamp,
				Value:     value,
			})
		}
	}
	sortTimeseries(series)
	return series, nil
}

// getFileFromS3Async returns a struct include result buf and error channels
func getFileFromS3Async(s *Storage, bucket string, key *string) s3ResultChan {
	rc := make(chan []byte)
	ec := make(chan error)
	go func() {
		r, err := getFileFromS3(s, bucket, key)
		rc <- r
		ec <- err
	}()
	return s3ResultChan{
		result: rc,
		err:    ec,
	}
}

func getFileFromS3(s *Storage, bucket string, key *string) ([]byte, error) {
	op := "getFileFromS3"
	req, resp := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    key,
	})

	err := req.Send()
	if err != nil {
		reqerr, ok := err.(awserr.RequestFailure)
		if reqerr.Code() == "NoSuchKey" && ok {
			return nil, &wm.Error{Code: wm.ENOTFOUND, Message: "Resource not found", Op: op}
		}
		return nil, &wm.Error{Op: op, Err: fmt.Errorf("feetching agg file from S3 returned error for key: %s: %w", *key, err)}
	}

	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, &wm.Error{Op: op, Err: fmt.Errorf("error reading response from s3 request: %w", err)}
	}
	return buf, nil
}
