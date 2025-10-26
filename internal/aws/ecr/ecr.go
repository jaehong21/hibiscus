package ecr

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/jaehong21/hibiscus/internal/aws"
)

var client *ecr.Client

func DescribeRepositories() ([]types.Repository, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}

	// TODO: need to fetch all repositories when there are more than 1000
	maxResults := int32(1000)
	repositories, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
		MaxResults: &maxResults,
	})
	if err != nil {
		return nil, err
	}

	var result []types.Repository
	result = append(result, repositories.Repositories...)

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(*result[j].CreatedAt)
	})

	return result, nil
}

func DescribeImages(repositoryName *string) ([]types.ImageDetail, error) {
	if err := setupClient(); err != nil {
		return nil, err
	}
	var (
		result    []types.ImageDetail
		nextToken *string
		max       = int32(1000)
	)

	for {
		input := &ecr.DescribeImagesInput{
			RepositoryName: repositoryName,
			Filter: &types.DescribeImagesFilter{
				TagStatus: types.TagStatusTagged,
			},
			MaxResults: &max,
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		resp, err := client.DescribeImages(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		result = append(result, resp.ImageDetails...)

		if resp.NextToken == nil || len(*resp.NextToken) == 0 {
			break
		}
		nextToken = resp.NextToken
	}

	// sort by ImagePushedAt
	sort.Slice(result, func(i, j int) bool {
		return result[i].ImagePushedAt.After(*result[j].ImagePushedAt)
	})

	return result, nil
}

func setupClient() error {
	if client != nil {
		return nil
	}

	cfg, err := aws.GetAWSConfig(context.Background())
	if err != nil {
		return err
	}

	client = ecr.NewFromConfig(cfg)
	return nil
}
