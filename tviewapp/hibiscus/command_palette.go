package hibiscus

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const commandPalettePage = "command-palette"

// commandPalette renders the ':' navigation overlay.
type commandPalette struct {
	app      *tview.Application
	pages    *tview.Pages
	services []string
	onSelect func(string)
	onClose  func()

	layout   *tview.Flex
	input    *tview.InputField
	list     *tview.List
	visible  bool
	filtered []string
}

func newCommandPalette(app *tview.Application, pages *tview.Pages, services []string, onSelect func(string), onClose func()) *commandPalette {
	input := tview.NewInputField().
		SetLabel(": ").
		SetFieldBackgroundColor(tcell.ColorBlack)

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true)
	list.SetTitle("Services")

	cp := &commandPalette{
		app:      app,
		pages:    pages,
		services: services,
		onSelect: onSelect,
		onClose:  onClose,
		input:    input,
		list:     list,
	}

	cp.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(list, 0, 1, false)

	list.SetSelectedFunc(func(index int, mainText string, secondary string, shortcut rune) {
		cp.choose(mainText)
	})

	list.SetDoneFunc(func() {
		cp.Hide()
	})

	cp.layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cp.Hide()
			return nil
		}
		return event
	})

	input.SetChangedFunc(func(text string) {
		cp.updateSuggestions(text)
	})

	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			cp.choose(cp.input.GetText())
		case tcell.KeyEsc:
			cp.Hide()
		}
	})

	cp.updateSuggestions("")
	return cp
}

func (c *commandPalette) Show() {
	if c.visible {
		return
	}
	c.visible = true
	c.input.SetText("")
	c.updateSuggestions("")
	c.pages.AddPage(commandPalettePage, centerPrimitive(c.layout, 50, 12), true, true)
	c.app.SetFocus(c.input)
}

func (c *commandPalette) Hide() {
	if !c.visible {
		return
	}
	c.visible = false
	c.pages.RemovePage(commandPalettePage)
	if c.onClose != nil {
		c.onClose()
	}
}

func (c *commandPalette) Visible() bool {
	return c.visible
}

func (c *commandPalette) choose(name string) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" || !c.isValid(name) {
		if len(c.filtered) == 0 {
			return
		}
		idx := c.list.GetCurrentItem()
		if idx < 0 || idx >= len(c.filtered) {
			idx = 0
		}
		name = c.filtered[idx]
	}

	c.Hide()
	if c.onSelect != nil {
		c.onSelect(name)
	}
}

func (c *commandPalette) updateSuggestions(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	c.list.Clear()
	c.filtered = c.filtered[:0]

	for _, svc := range c.services {
		if query == "" || strings.Contains(svc, query) {
			c.list.AddItem(svc, "", 0, nil)
			c.filtered = append(c.filtered, svc)
		}
	}

	if len(c.filtered) == 0 {
		c.list.AddItem("No matches", "", 0, nil)
	}
	c.list.SetCurrentItem(0)
}

func (c *commandPalette) isValid(name string) bool {
	return slices.Contains(c.services, name)
}

func centerPrimitive(content tview.Primitive, width, height int) tview.Primitive {
	grid := tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(content, 1, 1, 1, 1, 0, 0, true)
	grid.SetBackgroundColor(tcell.ColorBlack)
	return grid
}
