package elb

import (
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/internal/aws/elbv2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate table width with some padding
		tableWidth := msg.Width - 2

		// Set overall table widths
		m.table.loadBalancer.SetWidth(tableWidth)
		m.table.listener.SetWidth(tableWidth)
		m.table.rule.SetWidth(tableWidth)

		// Set table heights
		m.table.loadBalancer.SetHeight(msg.Height - 15)
		m.table.listener.SetHeight(msg.Height - 15)
		m.table.rule.SetHeight(msg.Height - 15)

		// Update loadBalancer table column widths with proportional columns
		nameWidth := int(float64(tableWidth) * 0.14)    // 14% of width
		typeWidth := int(float64(tableWidth) * 0.09)    // 9% of width
		dnsWidth := int(float64(tableWidth) * 0.43)     // 43% of width
		stateWidth := int(float64(tableWidth) * 0.09)   // 9% of width
		createdWidth := int(float64(tableWidth) * 0.20) // 20% of width

		loadBalancerColumns := []table.Column{
			{Title: "Name", Width: nameWidth},
			{Title: "Type", Width: typeWidth},
			{Title: "DNS Name", Width: dnsWidth},
			{Title: "State", Width: stateWidth},
			{Title: "Created at", Width: createdWidth},
		}
		m.table.loadBalancer.SetColumns(loadBalancerColumns)

		// Update listener table column widths with proportional columns
		protocolWidth := int(float64(tableWidth) * 0.19) // 19% of width
		portWidth := int(float64(tableWidth) * 0.14)     // 14% of width
		actionWidth := int(float64(tableWidth) * 0.62)   // 62% of width

		listenerColumns := []table.Column{
			{Title: "Protocol", Width: protocolWidth},
			{Title: "Port", Width: portWidth},
			{Title: "Default Action", Width: actionWidth},
		}
		m.table.listener.SetColumns(listenerColumns)

		// Update rule table column widths with proportional columns
		priorityWidth := int(float64(tableWidth) * 0.09)   // 9% of width
		condTypeWidth := int(float64(tableWidth) * 0.14)   // 14% of width
		valueWidth := int(float64(tableWidth) * 0.34)      // 34% of width
		actionTypeWidth := int(float64(tableWidth) * 0.14) // 14% of width
		targetWidth := int(float64(tableWidth) * 0.24)     // 24% of width

		ruleColumns := []table.Column{
			{Title: "Priority", Width: priorityWidth},
			{Title: "Condition Type", Width: condTypeWidth},
			{Title: "Value", Width: valueWidth},
			{Title: "Action Type", Width: actionTypeWidth},
			{Title: "Target", Width: targetWidth},
		}
		m.table.rule.SetColumns(ruleColumns)

		// Focus the active table based on current tab
		if m.tab == ELB_LOADBALANCER_TAB {
			m.table.loadBalancer.Focus()
			m.table.listener.Blur()
			m.table.rule.Blur()
		} else if m.tab == ELB_LISTENER_TAB {
			m.table.loadBalancer.Blur()
			m.table.listener.Focus()
			m.table.rule.Blur()
		} else if m.tab == ELB_RULE_TAB {
			m.table.loadBalancer.Blur()
			m.table.listener.Blur()
			m.table.rule.Focus()
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "R", "r":
			// Refresh current tab data
			if m.tab == ELB_LOADBALANCER_TAB {
				m.loading.msg = FETCHING_LOADBALANCERS_MSG
				return m, tea.Batch(fetchLoadBalancers(), m.loading.spinner.Tick)
			} else if m.tab == ELB_LISTENER_TAB {
				m.loading.msg = FETCHING_LISTENERS_MSG
				return m, tea.Batch(fetchListeners(&m.selectedLoadBalancerArn), m.loading.spinner.Tick)
			} else if m.tab == ELB_RULE_TAB {
				m.loading.msg = FETCHING_RULES_MSG
				return m, tea.Batch(fetchRules(&m.selectedListenerArn), m.loading.spinner.Tick)
			}

		case "/":
			// Activate search/filter mode
			m.textinput.Focus()
			// Blur all tables when searching
			m.table.loadBalancer.Blur()
			m.table.listener.Blur()
			m.table.rule.Blur()
			return m, nil

		case "esc":
			// If textinput is focused, unfocus it
			if m.textinput.Focused() {
				m.textinput.Blur()

				// Focus the appropriate table
				if m.tab == ELB_LOADBALANCER_TAB {
					m.table.loadBalancer.Focus()
				} else if m.tab == ELB_LISTENER_TAB {
					m.table.listener.Focus()
				} else if m.tab == ELB_RULE_TAB {
					m.table.rule.Focus()
				}

				return m, nil
			}

			// Otherwise handle navigating back
			if m.tab == ELB_RULE_TAB {
				m.tab = ELB_LISTENER_TAB
				m.loading.msg = FETCHING_LISTENERS_MSG

				// Update focus state
				m.table.rule.Blur()
				m.table.listener.Focus()

				loadBalancerArn := m.selectedLoadBalancerArn
				return m, tea.Batch(fetchListeners(&loadBalancerArn), m.loading.spinner.Tick)
			} else if m.tab == ELB_LISTENER_TAB {
				m.tab = ELB_LOADBALANCER_TAB
				m.loading.msg = FETCHING_LOADBALANCERS_MSG

				// Update focus state
				m.table.listener.Blur()
				m.table.loadBalancer.Focus()

				return m, tea.Batch(fetchLoadBalancers(), m.loading.spinner.Tick)
			}

		case "enter":
			// If textinput is focused, submit the search
			if m.textinput.Focused() {
				m.textinput.Blur()
				query := m.textinput.Value()
				m.textinput.Reset()

				if query == "" {
					// If search is empty, fetch all items
					if m.tab == ELB_LOADBALANCER_TAB {
						m.loading.msg = FETCHING_LOADBALANCERS_MSG
						m.table.loadBalancer.Focus()
						return m, tea.Batch(fetchLoadBalancers(), m.loading.spinner.Tick)
					}
					// We don't handle empty searches for other tabs here

					// Focus the appropriate table
					if m.tab == ELB_LOADBALANCER_TAB {
						m.table.loadBalancer.Focus()
					} else if m.tab == ELB_LISTENER_TAB {
						m.table.listener.Focus()
					} else if m.tab == ELB_RULE_TAB {
						m.table.rule.Focus()
					}

					return m, nil
				}

				// Handle filtering based on tab
				if m.tab == ELB_LOADBALANCER_TAB {
					m.loading.msg = FILTERING_LOADBALANCERS_MSG
					m.table.loadBalancer.Focus()
					return m, tea.Batch(filterLoadBalancers(query), m.loading.spinner.Tick)
				}
				// We don't implement filtering for other tabs in this implementation

				// Focus the appropriate table
				if m.tab == ELB_LOADBALANCER_TAB {
					m.table.loadBalancer.Focus()
				} else if m.tab == ELB_LISTENER_TAB {
					m.table.listener.Focus()
				} else if m.tab == ELB_RULE_TAB {
					m.table.rule.Focus()
				}

				return m, nil
			}

			// Navigate to selected item
			if m.tab == ELB_LOADBALANCER_TAB && len(m.table.loadBalancer.Rows()) > 0 {
				// Get the selected load balancer and fetch its listeners
				selectedRow := m.table.loadBalancer.SelectedRow()
				loadBalancers, err := getLbArnByName(selectedRow[0]) // First column is name
				if err != nil || len(loadBalancers) == 0 {
					m.err = err
					return m, nil
				}

				// Store the selected load balancer ARN and fetch listeners
				m.selectedLoadBalancerName = selectedRow[0]
				m.selectedLoadBalancerArn = *loadBalancers[0].LoadBalancerArn
				m.tab = ELB_LISTENER_TAB
				m.loading.msg = FETCHING_LISTENERS_MSG

				// Update focus state
				m.table.loadBalancer.Blur()
				m.table.listener.Focus()

				return m, tea.Batch(fetchListeners(&m.selectedLoadBalancerArn), m.loading.spinner.Tick)

			} else if m.tab == ELB_LISTENER_TAB && len(m.table.listener.Rows()) > 0 {
				// Get selected listener and fetch its rules
				idx := m.table.listener.Cursor()
				if idx >= 0 && idx < len(m.table.listener.Rows()) {
					// Get listeners for the current load balancer
					listeners, err := getListenersByLoadBalancerArn(&m.selectedLoadBalancerArn)
					if err != nil || len(listeners) == 0 {
						m.err = err
						return m, nil
					}

					// Make sure we have enough listeners to match the selected index
					if idx < len(listeners) {
						// Store the selected listener ARN and fetch rules
						m.selectedListenerArn = *listeners[idx].ListenerArn
						m.tab = ELB_RULE_TAB
						m.loading.msg = FETCHING_RULES_MSG

						// Update focus state
						m.table.listener.Blur()
						m.table.rule.Focus()

						return m, tea.Batch(fetchRules(&m.selectedListenerArn), m.loading.spinner.Tick)
					}
				}
			}
		}

	// Handle results from commands
	case fetchLoadBalancersMsg:
		m.loading.msg = ""
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.table.loadBalancer.SetRows(msg.Rows)
		m.table.loadBalancer.Focus()
		m.err = nil
		return m, nil

	case filterLoadBalancersMsg:
		m.loading.msg = ""
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.table.loadBalancer.SetRows(msg.Rows)
		m.table.loadBalancer.Focus()
		m.err = nil
		return m, nil

	case fetchListenersMsg:
		m.loading.msg = ""
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.table.listener.SetRows(msg.Rows)
		m.table.listener.Focus()
		m.err = nil
		return m, nil

	case fetchRulesMsg:
		m.loading.msg = ""
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.table.rule.SetRows(msg.Rows)
		m.table.rule.Focus()
		m.err = nil
		return m, nil
	}

	// Handle updates for sub-components
	newSpinner, cmd := m.loading.spinner.Update(msg)
	m.loading.spinner = newSpinner
	cmds = append(cmds, cmd)

	// Update tables based on current tab
	var cmd2 tea.Cmd

	// Update text input if needed
	if m.textinput.Focused() {
		newInput, cmd := m.textinput.Update(msg)
		m.textinput = newInput
		cmds = append(cmds, cmd)
	} else {
		// Only update tables if they're focused
		if m.tab == ELB_LOADBALANCER_TAB && m.table.loadBalancer.Focused() {
			m.table.loadBalancer, cmd2 = m.table.loadBalancer.Update(msg)
			cmds = append(cmds, cmd2)
		} else if m.tab == ELB_LISTENER_TAB && m.table.listener.Focused() {
			m.table.listener, cmd2 = m.table.listener.Update(msg)
			cmds = append(cmds, cmd2)
		} else if m.tab == ELB_RULE_TAB && m.table.rule.Focused() {
			m.table.rule, cmd2 = m.table.rule.Update(msg)
			cmds = append(cmds, cmd2)
		}
	}

	return m, tea.Batch(cmds...)
}

// Helper functions for looking up resources
func getLbArnByName(name string) ([]types.LoadBalancer, error) {
	loadBalancers, err := elbv2.DescribeLoadBalancers()
	if err != nil {
		return nil, err
	}

	var result []types.LoadBalancer
	for _, lb := range loadBalancers {
		if *lb.LoadBalancerName == name {
			result = append(result, lb)
		}
	}

	return result, nil
}

func getListenersByLoadBalancerArn(loadBalancerArn *string) ([]types.Listener, error) {
	return elbv2.DescribeListeners(loadBalancerArn)
}
