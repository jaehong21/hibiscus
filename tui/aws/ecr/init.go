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
		{Title: "Repository name"},
		{Title: "URI"},
		{Title: "Created at"},
	}
	ecrRepoImageColumns := []table.Column{
		{Title: "Tag"},
		{Title: "Pushed at"},
		{Title: "Size"},
		{Title: "Digest"},
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
