package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/config"
	"github.com/jaehong21/hibiscus/tui/aws/ecr"
	"github.com/jaehong21/hibiscus/tui/aws/route53"
)

type model struct {
	ecr     tea.Model
	route53 tea.Model
}

func New(config *config.Config) model {
	return model{
		ecr:     ecr.New(),
		route53: route53.New(),
	}
}

// https://github.com/charmbracelet/bubbletea/blob/491eda41276c3419d519bc8c622725fa587b7e37/tea.go#L513
// NOTE: needed to initialize the ecr model
// p.Run() call the Init() method of the model
// but for our custom models, we need to call the Init() method manually

func (m model) Init() tea.Cmd {
	config.SetTabKey(config.ECR_TAB)

	return tea.Batch(m.ecr.Init(), m.route53.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	tab := config.GetConfig().TabKey

	switch tab {
	case config.ECR_TAB:
		var ecrCmd tea.Cmd
		m.ecr, ecrCmd = m.ecr.Update(msg)
		cmds = append(cmds, ecrCmd)

	case config.ROUTE53_TAB:
		var route53Cmd tea.Cmd
		m.route53, route53Cmd = m.route53.Update(msg)
		cmds = append(cmds, route53Cmd)

	default:

	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s string

	tab := config.GetConfig().TabKey

	switch tab {
	case config.ECR_TAB:
		s += m.ecr.View()

	case config.ROUTE53_TAB:
		s += m.route53.View()

	default:

	}

	return s
}
