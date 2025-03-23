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
		{Title: "Name"},
		{Title: "Record count"},
		{Title: "ID"},
	}
	recordColumns := []table.Column{
		{Title: "Record Name"},
		{Title: "Type"},
		{Title: "Values"},
		{Title: "TTL"},
		{Title: "Weight"},
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
