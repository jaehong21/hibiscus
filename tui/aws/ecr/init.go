package ecr

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/tui/styles"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchRepos(), m.loading.spinner.Tick)
}

func New() Model {
	ti := styles.DefeaultTextInput()
	sp := styles.DefaultSpinner()

	ecrRepoColumns := []table.Column{
		{Title: "Repository name", Width: 28},
		{Title: "URI", Width: 75},
		{Title: "Created at", Width: 35},
	}
	ecrRepoImageColumns := []table.Column{
		{Title: "Tag", Width: 27},
		{Title: "Pushed at", Width: 33},
		{Title: "Size", Width: 10},
		{Title: "Digest", Width: 74},
	}

	return Model{
		textinput: ti,
		loading: loading{
			spinner: sp,
			msg:     FETCHING_REPOS_MSG,
		},
		tab: ECR_REPO_TAB,
		table: tables{
			ecrRepo:      table.New(table.WithColumns(ecrRepoColumns)),
			ecrRepoImage: table.New(table.WithColumns(ecrRepoImageColumns)),
		},
	}
}
