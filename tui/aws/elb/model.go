package elb

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
)

type Model struct {
	width  int
	height int

	textinput textinput.Model
	loading   loading

	tab   int
	table tables

	// Selected resources ARNs
	selectedLoadBalancerName string
	selectedLoadBalancerArn  string
	selectedListenerArn      string

	focused bool
	msg     string
	err     error
}

type tables struct {
	loadBalancer table.Model
	listener     table.Model
	rule         table.Model
}

type loading struct {
	spinner spinner.Model
	msg     string
}

func (m Model) Focused() bool {
	return m.focused
}
