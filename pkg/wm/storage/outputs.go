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

// GetRegionalOutputStats returns regional output statistics
func (s *Storage) GetRegionalOutputStats(params wm.DatacubeParams) (*wm.ModelRegionalOutputStat, error) {
	regionMap := make(map[string]*wm.ModelOutputStat)
	for _, level := range []string{"country", "admin1", "admin2", "admin3"} {
		var regionKey = fmt.Sprintf("regional/%s", level)
		stats, err := s.GetOutputStats(params, regionKey)
		if err == nil {
			regionMap[level] = stats
		}
	}
	var statsByRegion wm.ModelRegionalOutputStat
	err := mapstructure.Decode(regionMap, &statsByRegion)
	if err != nil {
		return nil, err
	}
	return &statsByRegion, nil
}

// GetOutputStats returns datacube output stats
func (s *Storage) GetOutputStats(params wm.DatacubeParams, filename string) (*wm.ModelOutputStat, error) {
	key := fmt.Sprintf("%s/%s/%s/%s/stats/%s.json",
		params.DataID, params.RunID, params.Resolution, params.Feature, filename)

	bucket := maasModelOutputBucket
	if params.RunID == "indicator" {
		bucket = maasIndicatorOutputBucket
	}
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, err
	}

	//The format is {"0":{ <stats> }}
	statsAt0 := make(map[string]map[string]float64)
	err = json.Unmarshal(buf, &statsAt0)
	if err != nil {
		s.logger.Errorw("Error while unmarshalling", "err", err)
		return nil, err
	}

	if len(statsAt0) == 0 {
		s.logger.Errorf("No stats found")
		return nil, nil
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

// GetOutputTimeseries returns datacube output timeseries
func (s *Storage) GetOutputTimeseries(params wm.DatacubeParams) ([]*wm.TimeseriesValue, error) {
	key := fmt.Sprintf("%s/%s/%s/%s/timeseries/s_%s_t_%s.json",
		params.DataID, params.RunID, params.Resolution, params.Feature, params.SpatialAggFunc, params.TemporalAggFunc)

	bucket := maasModelOutputBucket
	if params.RunID == "indicator" {
		bucket = maasIndicatorOutputBucket
	}
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, err
	}

	var series []*wm.TimeseriesValue
	err = json.Unmarshal(buf, &series)
	if err != nil {
		s.logger.Errorw("Error while unmarshalling", "err", err)
		return nil, err
	}
	return series, nil
}

// GetOutputTimeseriesByRegion returns timeseries data for a specific region
func (s *Storage) GetOutputTimeseriesByRegion(params wm.DatacubeParams, regionID string) ([]*wm.TimeseriesValue, error) {
	// Deconstruct Region ID to get admin region levels
	regions := strings.Split(regionID, "__")
	regionLevel := getRegionLevels()[len(regions)-1]
	key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/timeseries/%s.csv",
		params.DataID, params.RunID, params.Resolution, params.Feature, regionLevel, regionID)

	bucket := maasModelOutputBucket
	if params.RunID == "indicator" {
		bucket = maasIndicatorOutputBucket
	}
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, err
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
			return nil, err
		}
		// Parse header and find the index of the target value column
		if isHeader {
			for i, v := range record {
				if v == valueCol {
					valueColIndex = i
				}
			}
			if valueColIndex == -1 {
				return nil, fmt.Errorf("csv: column, %s does not exist", valueCol)
			}
			isHeader = false
		} else {
			timestamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				return nil, err
			}
			value, err := strconv.ParseFloat(record[valueColIndex], 64)
			if err != nil {
				return nil, err
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

	data := make(map[string][]interface{})
	for _, level := range []string{"country", "admin1", "admin2", "admin3"} {
		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/s_%s_t_%s.json",
			params.DataID, params.RunID, params.Resolution, params.Feature, level,
			timestamp, params.SpatialAggFunc, params.TemporalAggFunc)

		bucket := maasModelOutputBucket
		if params.RunID == "indicator" {
			bucket = maasIndicatorOutputBucket
		}

		buf, err := getFileFromS3(s, bucket, aws.String(key))

		if err != nil {
			reqerr, ok := err.(awserr.RequestFailure)
			if reqerr.Code() == "NoSuchKey" && ok {
				data[level] = make([]interface{}, 0)
			} else {
				return nil, err
			}
		} else {
			var points []interface{}
			err = json.Unmarshal(buf, &points)
			if err != nil {
				s.logger.Errorw("Error while unmarshalling", "err", err)
				return nil, err
			}
			data[level] = points
		}
	}

	var regionalData wm.ModelOutputRegionalAdmins
	err := mapstructure.Decode(data, &regionalData)
	if err != nil {
		s.logger.Errorw("Error while unmarshalling admin regions", "err", err)
		return nil, err
	}
	return &regionalData, nil
}

// GetRawData returns datacube output or indicator raw data
func (s *Storage) GetRawData(params wm.DatacubeParams) ([]*wm.ModelOutputRawDataPoint, error) {
	key := fmt.Sprintf("%s/%s/raw/%s/raw/raw.json",
		params.DataID, params.RunID, params.Feature)

	bucket := maasModelOutputBucket
	if params.RunID == "indicator" {
		bucket = maasIndicatorOutputBucket
	}
	buf, err := getFileFromS3(s, bucket, aws.String(key))
	if err != nil {
		return nil, err
	}

	var series []*wm.ModelOutputRawDataPoint
	err = json.Unmarshal(buf, &series)
	if err != nil {
		s.logger.Errorw("Error while unmarshalling", "err", err)
		return nil, err
	}
	return series, nil
}

func getFileFromS3(s *Storage, bucket string, key *string) ([]byte, error) {
	req, resp := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    key,
	})

	err := req.Send()
	if err != nil {
		s.logger.Errorw("Fetching agg file from S3 returned error", "err", err)
		return nil, err
	}

	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		s.logger.Errorw("Error reading response from S3 request", "err", err)
		return nil, err
	}
	return buf, nil
}
