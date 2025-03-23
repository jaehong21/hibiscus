package styles

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

func DefeaultTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "`/` Search mode, `:` Command mode"
	ti.CharLimit = 100

	return ti
}

func DefaultSpinner() spinner.Model {
	sp := spinner.New()
	sp.Spinner = spinner.Meter
	sp.Spinner.FPS = time.Second / 7 / 2
	sp.Style = SpinnerStyle

	return sp
}
