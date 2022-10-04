package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

// BucketInfo contains bucket name mapping information
type BucketInfo struct {
	TileOutputBucket string `json:"tileOutputBucket"`
	VectorTileBucket string `json:"vectorTileBucket"`
	ModelsBucket     string `json:"modelsBucket"`
	IndicatorsBucket string `json:"indicatorsBucket"`
}

// Storage wraps the client and serves as the basis of the wm.MaaSData interface.
type Storage struct {
	client     *s3.S3
	bucketInfo *BucketInfo
	logger     *zap.SugaredLogger
}

// New instantiates and returns a new Storage instance using the provided Config.
func New(cfg *aws.Config, bucketInfo *BucketInfo, logger *zap.SugaredLogger) (*Storage, error) {
	sess := session.Must(session.NewSession(cfg))
	client := s3.New(sess)

	return &Storage{
		client,
		bucketInfo,
		logger,
	}, nil
}
