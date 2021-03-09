package storage

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetVectorTile returns mapbox vectortile
func (s *Storage) GetVectorTile(zoom, x, y uint32, tilesetName string) ([]byte, error) {
	key := fmt.Sprintf("%d/%d/%d.pbf", zoom, x, y)
	fmt.Println(tilesetName, key)

	// Retrieve protobuf tile from S3
	req, resp := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(vectorTileBucket),
		Key:    aws.String(key),
	})
	err := req.Send()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == s3.ErrCodeNoSuchKey {
				// Tile not found errors are expected
				return nil, err
			}
			return nil, err
		}
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
