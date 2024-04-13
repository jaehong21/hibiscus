package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/jaehong21/hibiscus/internal/aws"
)

var client *s3.Client

func DescribeBuckets() ([]types.Bucket, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}
	buckets, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return buckets.Buckets, nil
}

func setupClient() error {
	if client != nil {
		return nil
	}
	cfg, err := aws.GetAWSConfig(context.Background())
	if err != nil {
		return err
	}
	client = s3.NewFromConfig(cfg)
	return nil
}
