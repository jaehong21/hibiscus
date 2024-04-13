package ecr

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	. "github.com/jaehong21/hibiscus/tui/styles"
	"github.com/muesli/reflow/wordwrap"
)

func (m Model) View() string {
	var s string

	s += "\n" + m.textinput.View() + "\n"

	s += m.loadingRender()
	s += m.tableRender()
	s += m.errorRender(m.width - 10)

	s += m.footerRender()

	// TODO: help render with keybindings
	// s += m.helpRender()

	return s
}

func (m Model) loadingRender() string {
	var s string

	s += "\n"
	if m.loading.msg != "" {
		s += m.loading.spinner.View() + " " + m.loading.msg
	}
	s += "\n"

	return s
}

func (m Model) tableRender() string {
	var s string

	st := table.DefaultStyles()
	st.Header = TableHeaderStyle
	st.Selected = TableSelectedStyle

	switch m.tab {
	case ECR_REPO_TAB:
		m.table.ecrRepo.SetStyles(st)
		s += TableBaseStyle.Render(m.table.ecrRepo.View()) + "\n"

	case ECR_IMAGE_TAB:
		m.table.ecrRepoImage.SetStyles(st)
		s += TableBaseStyle.Render(m.table.ecrRepoImage.View()) + "\n"
	}

	return s
}

func (m Model) errorRender(width int) string {
	var s string
	if m.err != nil {
		s += "\n" + wordwrap.String(SpinnerStyle.Render("🤬 Error: "+m.err.Error()), width) + "\n\n"
	}
	return s
}

func (m Model) footerRender() string {
	var s string

	if m.tab == ECR_REPO_TAB {
		s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.ecrRepo.Rows()))) + " "
		s += " " + TabSelectedStyle.Render("<repository>")
	} else if m.tab == ECR_IMAGE_TAB {
		s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.ecrRepoImage.Rows()))) + " "
		s += " " + TabBaseStyle.Render("<repository: "+m.table.ecrRepo.SelectedRow()[0]+">")
		s += " " + TabSelectedStyle.Render("<image>")
	}

	s += "\n\n"

	return s
}

func (m Model) helpRender() string {
	var s string
	alignWidth := "17" // Must be string

	items := []string{
		fmt.Sprintf("%-"+alignWidth+"s Quit", "<crtl + c>"),
		fmt.Sprintf("%-"+alignWidth+"s Search", "</>"),
		fmt.Sprintf("%-"+alignWidth+"s Refresh", "<R(r)>"),
	}

	if m.tab == ECR_REPO_TAB {
		items = append(items,
			fmt.Sprintf("%-"+alignWidth+"s Select", "<enter>"),
			fmt.Sprintf("%-"+alignWidth+"s Copy URI", "<C(c)>, <Y(y)>"),
		)
	}

	if m.tab == ECR_IMAGE_TAB {
		items = append(items,
			fmt.Sprintf("%-"+alignWidth+"s Back", "<esc>, <Q(q)>"),
			fmt.Sprintf("%-"+alignWidth+"s Copy Tag", "<C(c)>, <Y(y)>"),
		)
	}

	for _, item := range items {
		s += " " + HelpStyle.Render(item) + "\n"
	}

	return s
}
