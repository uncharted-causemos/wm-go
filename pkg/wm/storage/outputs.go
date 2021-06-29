package storage

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/mapstructure"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
	"io/ioutil"
)

// GetOutputStats returns model output stats
func (s *Storage) GetOutputStats(params wm.ModelOutputParams) (*wm.ModelOutputStat, error) {
	key := fmt.Sprintf("%s/%s/%s/%s/stats/stats.json",
		params.ModelID, params.RunID, params.Resolution, params.Feature)

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

// GetOutputTimeseries returns model output timeseries
func (s *Storage) GetOutputTimeseries(params wm.ModelOutputParams) ([]*wm.TimeseriesValue, error) {
	key := fmt.Sprintf("%s/%s/%s/%s/timeseries/s_%s_t_%s.json",
		params.ModelID, params.RunID, params.Resolution, params.Feature, params.SpatialAggFunc, params.TemporalAggFunc)

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

// GetRegionAggregation returns regional data for ALL admin regions at ONE timestamp
func (s *Storage) GetRegionAggregation(params wm.ModelOutputParams, timestamp string) (*wm.ModelOutputRegionalAdmins, error) {

	data := make(map[string][]interface{})
	for _, level := range []string{"country", "admin1", "admin2", "admin3"} {
		key := fmt.Sprintf("%s/%s/%s/%s/regional/%s/aggs/%s/s_%s_t_%s.json",
			params.ModelID, params.RunID, params.Resolution, params.Feature, level,
			timestamp, params.SpatialAggFunc, params.TemporalAggFunc)

		bucket := maasModelOutputBucket
		if params.RunID == "indicator" {
			bucket = maasIndicatorOutputBucket
		}

		buf, err := getFileFromS3(s, bucket, aws.String(key))

		if err != nil {
			regerr, ok := err.(awserr.RequestFailure);
			if regerr.Code() == "NoSuchKey" && ok {
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

// GetRawData returns model output or indicator raw data
func (s *Storage) GetRawData(params wm.ModelOutputParams) ([]*wm.ModelOutputRawDataPoint, error) {
	key := fmt.Sprintf("%s/%s/raw/%s/raw/raw.json",
		params.ModelID, params.RunID, params.Feature)

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
