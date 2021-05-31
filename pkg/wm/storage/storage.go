package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

const outputBucket = "tiles-v3"
const vectorTileBucket = "vector-tiles"
const maasModelOutputBucket = "models"
const maasIndicatorOutputBucket = "indicators"

// Storage wraps the client and serves as the basis of the wm.MaaSData interface.
type Storage struct {
	client *s3.S3
	logger *zap.SugaredLogger
}

// New instantiates and returns a new Storage instance using the provided Config.
func New(cfg *aws.Config, logger *zap.SugaredLogger) (*Storage, error) {
	sess := session.Must(session.NewSession(cfg))
	client := s3.New(sess)

	return &Storage{
		client,
		logger,
	}, nil
}
