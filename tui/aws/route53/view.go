package route53

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
	case ROUTE53_HOSTED_ZONE_TAB:
		m.table.hostedZone.SetStyles(st)
		s += TableBaseStyle.Render(m.table.hostedZone.View()) + "\n"

	case ROUTE53_RECORD_TAB:
		m.table.record.SetStyles(st)
		s += TableBaseStyle.Render(m.table.record.View()) + "\n"
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

	if m.tab == ROUTE53_HOSTED_ZONE_TAB {
		s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.hostedZone.Rows()))) + " "
		s += " " + TabSelectedStyle.Render("<hosted zone>")
	} else if m.tab == ROUTE53_RECORD_TAB {
		// s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.record.Rows()))) + " "
		s += " " + TabBaseStyle.Render("<hosted zone: "+m.table.hostedZone.SelectedRow()[0]+">")
		s += " " + TabSelectedStyle.Render("<record>")
	}

	s += "\n\n"

	return s
}

func (m Model) helpRender() string {
	var s string
	alignWidth := "17" // Must be string

	items := []string{
		fmt.Sprintf("%-"+alignWidth+"s Quit", "<crtl + c>"),
		fmt.Sprintf("%-"+alignWidth+"s Navigate", "<:>"),
		fmt.Sprintf("%-"+alignWidth+"s Search", "</>"),
		fmt.Sprintf("%-"+alignWidth+"s Refresh", "<R(r)>"),
	}

	if m.tab == ROUTE53_HOSTED_ZONE_TAB {
		items = append(items,
			fmt.Sprintf("%-"+alignWidth+"s Select", "<enter>"),
		)
	}

	if m.tab == ROUTE53_RECORD_TAB {
		items = append(items,
			fmt.Sprintf("%-"+alignWidth+"s Back", "<esc>, <Q(q)>"),
		)
	}

	for _, item := range items {
		s += " " + HelpStyle.Render(item) + "\n"
	}

	return s
}
