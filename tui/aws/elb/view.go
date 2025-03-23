package elb

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	. "github.com/jaehong21/hibiscus/tui/styles"
	"github.com/muesli/reflow/wordwrap"
)

func (m Model) View() string {
	var s string

	s += "\n" + m.textinput.View()

	s += m.loadingRender()
	s += m.errorRender(m.width - 10)
	s += m.tableRender()
	s += m.messageRender(m.width - 10)

	s += m.footerRender()

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
	case ELB_LOADBALANCER_TAB:
		m.table.loadBalancer.SetStyles(st)
		s += TableBaseStyle.Render(m.table.loadBalancer.View()) + "\n"

	case ELB_LISTENER_TAB:
		m.table.listener.SetStyles(st)
		s += TableBaseStyle.Render(m.table.listener.View()) + "\n"

	case ELB_RULE_TAB:
		m.table.rule.SetStyles(st)
		s += TableBaseStyle.Render(m.table.rule.View()) + "\n"
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

func (m Model) messageRender(width int) string {
	var s string
	if m.msg != "" {
		s += " " + wordwrap.String(SuccessStyle.Render("✓ "+m.msg), width) + "\n"
	} else {
		s += ""
	}
	return s
}

func (m Model) footerRender() string {
	var s string

	switch m.tab {
	case ELB_LOADBALANCER_TAB:
		s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.loadBalancer.Rows()))) + " "
		s += " " + TabSelectedStyle.Render("<load balancer>")
	case ELB_LISTENER_TAB:
		s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.listener.Rows()))) + " "
		s += " " + TabBaseStyle.Render("<load balancer: "+m.selectedLoadBalancerName+">")
		s += " " + TabSelectedStyle.Render("<listener>")
	case ELB_RULE_TAB:
		s += " " + HelpStyle.Render(fmt.Sprintf("Total items: %d", len(m.table.rule.Rows()))) + " "
		s += " " + TabBaseStyle.Render("<load balancer>")
		s += " " + TabBaseStyle.Render("<listener: "+m.selectedListenerArn+">")
		s += " " + TabSelectedStyle.Render("<rule>")
	}

	s += "\n\n"

	return s
}
