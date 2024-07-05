package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/jaehong21/hibiscus/config"
)

func GetAWSConfig(ctx context.Context) (aws.Config, error) {
	profile := config.GetConfig().AwsProfile

	if profile == "" {
		return awsConfig.LoadDefaultConfig(ctx)
	}

	return awsConfig.LoadDefaultConfig(ctx, awsConfig.WithSharedConfigProfile(profile))
}
