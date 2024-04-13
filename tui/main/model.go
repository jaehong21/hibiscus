package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/tui/aws/ecr"
	"github.com/jaehong21/hibiscus/tui/aws/route53"
)

type model struct {
	resourceType int
	ecr          tea.Model
	route53      tea.Model
}

func New() model {
	return model{
		resourceType: ECR_TYPE,

		ecr:     ecr.New(),
		route53: route53.New(),
	}
}

// https://github.com/charmbracelet/bubbletea/blob/491eda41276c3419d519bc8c622725fa587b7e37/tea.go#L513
// NOTE: needed to initialize the ecr model
// p.Run() call the Init() method of the model
// but for our custom models, we need to call the Init() method manually

func (m model) Init() tea.Cmd {
	return tea.Batch(m.ecr.Init(), m.route53.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		ecrCmd     tea.Cmd
		route53Cmd tea.Cmd
	)

	switch rs := m.resourceType; rs {
	case ECR_TYPE:
		m.ecr, ecrCmd = m.ecr.Update(msg)

	case ROUTE53_TYPE:
		m.route53, route53Cmd = m.route53.Update(msg)

	default:

	}

	return m, tea.Batch(ecrCmd, route53Cmd)
}

func (m model) View() string {
	var s string

	switch rs := m.resourceType; rs {
	case ECR_TYPE:
		s += m.ecr.View()

	case ROUTE53_TYPE:
		s += m.route53.View()

	default:

	}

	return s
}
