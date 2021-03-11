package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const outputBucket = "tiles-v3"
const vectorTileBucket = "vector-tiles"

// Storage wraps the client and serves as the basis of the wm.MaaSData interface.
type Storage struct {
	client *s3.S3
}

// New instantiates and returns a new Storage instance using the provided Config.
func New(cfg *aws.Config) (*Storage, error) {
	sess := session.Must(session.NewSession(cfg))
	client := s3.New(sess)

	return &Storage{
		client,
	}, nil
}
