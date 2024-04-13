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
	repositories, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{})
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
	images, err := client.DescribeImages(context.TODO(), &ecr.DescribeImagesInput{
		RepositoryName: repositoryName,
	})
	if err != nil {
		return nil, err
	}

	var result []types.ImageDetail
	for _, image := range images.ImageDetails {
		// list images with tag only
		if len(image.ImageTags) > 0 {
			result = append(result, image)
		}
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
