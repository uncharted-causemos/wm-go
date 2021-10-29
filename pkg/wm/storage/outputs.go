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

func getRegionLevels() []string {
	return []string{"country", "admin1", "admin2", "admin3"}
}

// Get the s3 bucket based on the output runID
func getBucket(outputRunID string) string {
	bucket := maasModelOutputBucket
	if outputRunID == "indicator" {
		bucket = maasIndicatorOutputBucket
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

	buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
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

	buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	// Read and parse csv
	r := csv.NewReader(bytes.NewReader(buf))
	records, err := r.ReadAll()
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
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
	op := "Storage.GetOutputTimeseries"
	key := fmt.Sprintf("%s/%s/%s/%s/timeseries/s_%s_t_%s.json",
		params.DataID, params.RunID, params.Resolution, params.Feature, params.SpatialAggFunc, params.TemporalAggFunc)

	buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	var series []*wm.TimeseriesValue
	err = json.Unmarshal(buf, &series)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return series, nil
}

// GetOutputTimeseriesByRegion returns timeseries data for a specific region
func (s *Storage) GetOutputTimeseriesByRegion(params wm.DatacubeParams, regionID string) ([]*wm.TimeseriesValue, error) {
	op := "Storage.GetOutputTimeseriesByRegion"
	// Deconstruct Region ID to get admin region levels
	regions := strings.Split(regionID, "__")
	regionLevel := getRegionLevels()[len(regions)-1]
	key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/timeseries/%s.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature, regionLevel, regionID)

	buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	// Read and parse csv
	series := make([]*wm.TimeseriesValue, 0)
	r := csv.NewReader(bytes.NewReader(buf))
	isHeader := true
	valueCol := fmt.Sprintf("s_%s_t_%s", params.SpatialAggFunc, params.TemporalAggFunc)
	if params.SpatialAggFunc == "count" {
		// when counting data points spatially, temporal aggregation function doesn't matter
		valueCol = "s_count"
	}
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
				return nil, &wm.Error{Op: op, Err: err}
			}
			series = append(series, &wm.TimeseriesValue{
				Timestamp: timestamp,
				Value:     value,
			})
		}
	}
	return series, nil
}

// GetRegionAggregation returns regional data for ALL admin regions at ONE timestamp
func (s *Storage) GetRegionAggregation(params wm.DatacubeParams, timestamp string) (*wm.ModelOutputRegionalAdmins, error) {
	op := "Storage.GetRegionAggregation"

	data := make(map[string][]interface{})
	for _, level := range []string{"country", "admin1", "admin2", "admin3"} {
		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/s_%s_t_%s.json",
			params.DataID, params.RunID, params.Resolution, params.Feature, level,
			timestamp, params.SpatialAggFunc, params.TemporalAggFunc)

		buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))

		if err != nil {
			if wm.ErrorCode(err) == wm.ENOTFOUND {
				data[level] = make([]interface{}, 0)
				continue
			} else {
				return nil, &wm.Error{Op: op, Err: err}
			}
		}

		var points []interface{}
		err = json.Unmarshal(buf, &points)
		if err != nil {
			return nil, &wm.Error{Op: op, Err: err}
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

// GetRegionHierarchy returns region hierarchy output
func (s *Storage) GetRegionHierarchy(params wm.HierarchyParams) (*wm.ModelOutputHierarchy, error) {
	op := "Storage.GetRegionHierarchy"
	key := fmt.Sprintf("%s/%s/raw/%s/hierarchy/hierarchy.json",
		params.DataID, params.RunID, params.Feature)
	bucket := maasModelOutputBucket
	if params.RunID == "indicator" {
		bucket = maasIndicatorOutputBucket
	}
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	var output wm.ModelOutputHierarchy
	err = json.Unmarshal(buf, &output)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &output, nil
}

// GetHierarchyLists returns region hierarchies in list form
func (s *Storage) GetHierarchyLists(params wm.RegionListParams) (*wm.RegionListOutput, error) {
	op := "Storage.GetHierarchyLists"
	var regionalData wm.RegionListOutput
	// allOutputMap is meant to be a map from strings ie. 'country' to a set (map[string]bool is used as a set)
	allOutputMap := make(map[string]map[string]bool)
	for _, region := range getRegionLevels() {
		allOutputMap[region] = make(map[string]bool)
	}
	// Populate sets in allOutputMap map values with regions
	for _, runID := range params.RunIDs {
		key := fmt.Sprintf("%s/%s/raw/%s/hierarchy/region_lists.json", params.DataID, runID, params.Feature)
		bucket := getBucket(runID)
		buf, err := getFileFromS3(s, bucket, aws.String(key))
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

// GetRawData returns datacube output or indicator raw data
func (s *Storage) GetRawData(params wm.DatacubeParams) ([]*wm.ModelOutputRawDataPoint, error) {
	op := "Storage.GetRawData"
	key := fmt.Sprintf("%s/%s/raw/%s/raw/raw.json",
		params.DataID, params.RunID, params.Feature)

	buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	var series []*wm.ModelOutputRawDataPoint
	err = json.Unmarshal(buf, &series)
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return series, nil
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

	buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
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
					filteredValues = append(filteredValues, timeseries)
					break
				}
			}
		}
	}
	return filteredValues, nil
}

// GetQualifierData returns datacube output data broken down by qualifiers for ONE timestamp
func (s *Storage) GetQualifierData(params wm.DatacubeParams, timestamp string, qualifiers []string) ([]*wm.ModelOutputQualifierBreakdown, error) {
	op := "Storage.GetQualifierData"
	allQualifiers := make([]*wm.ModelOutputQualifierBreakdown, len(qualifiers))

	for qualifierIndex, qualifier := range qualifiers {
		key := fmt.Sprintf("%s/%s/%s/%s/timeseries/qualifiers/%s/s_%s_t_%s.csv",
			params.DataID, params.RunID, params.Resolution, params.Feature, qualifier,
			params.SpatialAggFunc, params.TemporalAggFunc)
		// Want to return one row
		// CSV format:
		// timestamp,Battles,Protests,Riots,Strategic developments,Violence against civilians
		// 852076800000,1583.0,0.0,0.0,0.0,313.0
		// 854755200000,187.0,3.0,0.0,0.0,40.0

		buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))
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
	for _, level := range []string{"country", "admin1", "admin2", "admin3"} {

		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/qualifiers/%s.csv",
			params.DataID, params.RunID, params.Resolution, params.Feature, level, timestamp, qualifier)
		// Want to return all values from one column grouped by region
		// CSV format:
		// id,qualifier,s_sum_t_mean,s_mean_t_mean,s_sum_t_sum,s_mean_t_sum
		// Central African Republic__Bangui,Protests,0.0,0.0,0.0,0.0
		// Central African Republic__Bangui,Violence against civilians,0.0,0.0,0.0,0.0
		// Central African Republic__Ouham,Violence against civilians,10.0,10.0,10.0,10.0

		buf, err := getFileFromS3(s, getBucket(params.RunID), aws.String(key))

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
