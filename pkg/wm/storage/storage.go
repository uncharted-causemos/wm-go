package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Storage wraps the client and serves as the basis of the wm.MaaSData interface.
type Storage struct {
	client *s3.S3
	bucket string
}

// New instantiates and returns a new Storage instance using the provided Config.
func New(cfg *aws.Config, bucket string) (*Storage, error) {
	sess := session.Must(session.NewSession(cfg))
	client := s3.New(sess)

	return &Storage{
		client,
		bucket,
	}, nil
}
