package route53

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/tui/styles"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchHostedZone(), m.loading.spinner.Tick)
}

func New() Model {
	ti := styles.DefeaultTextInput()
	sp := styles.DefaultSpinner()

	hostedZoneColumns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Record count", Width: 15},
		{Title: "ID", Width: 50},
	}
	recordColumns := []table.Column{
		{Title: "Record Name", Width: 38},
		{Title: "Type", Width: 7},
		{Title: "Values", Width: 70},
		{Title: "TTL", Width: 8},
		{Title: "Weight", Width: 7},
	}

	return Model{
		textinput: ti,
		loading: loading{
			spinner: sp,
			msg:     FETCHING_HOSTED_ZONES_MSG,
		},
		tab: ROUTE53_HOSTED_ZONE_TAB,
		table: tables{
			hostedZone: table.New(table.WithColumns(hostedZoneColumns)),
			record:     table.New(table.WithColumns(recordColumns)),
		},
	}
}
