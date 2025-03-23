package elb

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/tui/styles"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchLoadBalancers(), m.loading.spinner.Tick)
}

func New() Model {
	ti := styles.DefeaultTextInput()
	sp := styles.DefaultSpinner()

	loadBalancerColumns := []table.Column{
		{Title: "Name"},
		{Title: "Type"},
		{Title: "DNS Name"},
		{Title: "State"},
		{Title: "Created at"},
	}
	listenerColumns := []table.Column{
		{Title: "Protocol"},
		{Title: "Port"},
		{Title: "Default Action"},
	}
	ruleColumns := []table.Column{
		{Title: "Priority"},
		{Title: "Condition Type"},
		{Title: "Value"},
		{Title: "Action Type"},
		{Title: "Target"},
	}

	return Model{
		textinput: ti,
		loading: loading{
			spinner: sp,
			msg:     FETCHING_LOADBALANCERS_MSG,
		},
		tab: ELB_LOADBALANCER_TAB,
		table: tables{
			loadBalancer: table.New(table.WithColumns(loadBalancerColumns)),
			listener:     table.New(table.WithColumns(listenerColumns)),
			rule:         table.New(table.WithColumns(ruleColumns)),
		},
	}
}
