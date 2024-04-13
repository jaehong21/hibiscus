package ecr

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/internal/aws/ecr"
	"github.com/jaehong21/hibiscus/utils"
)

type fetchReposMsg struct {
	Rows []table.Row
	Err  error
}

func fetchRepos() tea.Cmd {
	return func() tea.Msg {
		repos, err := ecr.DescribeRepositories()
		if err != nil {
			return fetchReposMsg{Err: err}
		}

		return fetchReposMsg{Rows: getEcrRepoRows(&repos)}
	}
}

type filterReposMsg struct {
	Rows []table.Row
	Err  error
}

func filterRepos(query string) tea.Cmd {
	return func() tea.Msg {
		repos, err := ecr.DescribeRepositories()
		if err != nil {
			return filterReposMsg{Err: err}
		}

		var filtered []types.Repository
		for _, repo := range repos {
			if strings.Contains(strings.ToLower(*repo.RepositoryName), strings.ToLower(query)) {
				filtered = append(filtered, repo)
			}
		}

		return filterReposMsg{Rows: getEcrRepoRows(&filtered)}
	}
}

func getEcrRepoRows(repos *[]types.Repository) []table.Row {
	rows := []table.Row{}
	for _, repo := range *repos {
		rows = append(rows, table.Row{
			*repo.RepositoryName,
			*repo.RepositoryUri,
			// utils.GetAgeFromTime(repo.CreatedAt),
			(*repo.CreatedAt).Local().String(),
		})
	}

	return rows
}

type fetchImagesMsg struct {
	Rows []table.Row
	Err  error
}

func fetchImages(repositoryName *string) tea.Cmd {
	return func() tea.Msg {
		images, err := ecr.DescribeImages(repositoryName)
		if err != nil {
			return fetchImagesMsg{Err: err}
		}

		return fetchImagesMsg{Rows: getEcrImageRows(&images)}
	}
}

type filterImagesMsg struct {
	Rows []table.Row
	Err  error
}

func filterImages(repositoryName *string, query string) tea.Cmd {
	return func() tea.Msg {
		images, err := ecr.DescribeImages(repositoryName)
		if err != nil {
			return fetchImagesMsg{Err: err}
		}

		var filtered []types.ImageDetail
		for _, image := range images {
			for _, tag := range image.ImageTags {
				if strings.Contains(strings.ToLower(tag), strings.ToLower(query)) {
					filtered = append(filtered, image)
				}
			}
		}

		return filterImagesMsg{Rows: getEcrImageRows(&filtered)}
	}
}

func getEcrImageRows(images *[]types.ImageDetail) []table.Row {
	rows := []table.Row{}
	for _, image := range *images {
		var imageTags string
		if len(image.ImageTags) > 1 {
			imageTags = image.ImageTags[0]
		} else {
			imageTags = strings.Join(image.ImageTags, "\n")
		}

		rows = append(rows, table.Row{
			imageTags,
			(*image.ImagePushedAt).Local().String(),
			utils.GetSizeFromByte(image.ImageSizeInBytes),
			*image.ImageDigest,
		})
	}

	return rows
}
