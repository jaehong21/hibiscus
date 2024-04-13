package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

var AWS_PROFILE = ""

func GetAWSConfig(ctx context.Context) (aws.Config, error) {
	if AWS_PROFILE == "" {
		return config.LoadDefaultConfig(ctx)
	}

	return config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(AWS_PROFILE))
}
