package styles

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// https://github.com/charmbracelet/lipgloss/blob/master/examples/table/languages/main.go

const TABLE_BLANK_WIDTH = 12

const (
	white     = lipgloss.Color("#fff")
	whiteText = lipgloss.Color("229")
	black     = lipgloss.Color("#000")
	grayText  = lipgloss.Color("240")
	gray      = lipgloss.Color("245")
	lightGray = lipgloss.Color("241")

	violet  = lipgloss.Color("57")
	skyblue = lipgloss.Color("#87cdfa")
	mint    = lipgloss.Color("#98e7c0")

	red    = lipgloss.Color("196")
	pink   = lipgloss.Color("205")
	orange = lipgloss.Color("208")
	yellow = lipgloss.Color("228")
	green  = lipgloss.Color("2")
	purple = lipgloss.Color("99")
)

var SpinnerStyle = lipgloss.NewStyle().Foreground(pink)

var (
	TabBaseStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(whiteText).Background(violet).
			Bold(true)

	TabSelectedStyle = TabBaseStyle.Copy().
				Foreground(black).Background(mint)
)

var (
	TableBaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(grayText)

	TableHeaderStyle = table.DefaultStyles().Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(grayText).
				BorderBottom(true).
				Bold(false)

	TableSelectedStyle = table.DefaultStyles().Selected.
				Foreground(whiteText).Background(violet).
				Bold(false)
)

var HelpStyle = lipgloss.NewStyle().
	Foreground(gray)
