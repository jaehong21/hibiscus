package ecr

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

	focused bool
	msg     string
	err     error
}

type tables struct {
	ecrRepo      table.Model
	ecrRepoImage table.Model
}

type loading struct {
	spinner spinner.Model
	msg     string
}

func (m Model) Focused() bool {
	return m.focused
}
