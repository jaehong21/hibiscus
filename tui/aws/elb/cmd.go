package elb

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/internal/aws/elbv2"
)

// Load Balancers
type fetchLoadBalancersMsg struct {
	Rows []table.Row
	Err  error
}

func fetchLoadBalancers() tea.Cmd {
	return func() tea.Msg {
		loadBalancers, err := elbv2.DescribeLoadBalancers()
		if err != nil {
			return fetchLoadBalancersMsg{Err: err}
		}

		return fetchLoadBalancersMsg{Rows: getLoadBalancerRows(&loadBalancers)}
	}
}

type filterLoadBalancersMsg struct {
	Rows []table.Row
	Err  error
}

func filterLoadBalancers(query string) tea.Cmd {
	return func() tea.Msg {
		loadBalancers, err := elbv2.DescribeLoadBalancers()
		if err != nil {
			return filterLoadBalancersMsg{Err: err}
		}

		var filtered []types.LoadBalancer
		for _, lb := range loadBalancers {
			if strings.Contains(strings.ToLower(*lb.LoadBalancerName), strings.ToLower(query)) {
				filtered = append(filtered, lb)
			}
		}

		return filterLoadBalancersMsg{Rows: getLoadBalancerRows(&filtered)}
	}
}

func getLoadBalancerRows(loadBalancers *[]types.LoadBalancer) []table.Row {
	rows := []table.Row{}
	for _, lb := range *loadBalancers {
		lbType := "Unknown"
		if lb.Type != "" {
			lbType = string(lb.Type)
		}

		state := "Unknown"
		if lb.State != nil && lb.State.Code != "" {
			state = string(lb.State.Code)
		}

		rows = append(rows, table.Row{
			*lb.LoadBalancerName,
			lbType,
			*lb.DNSName,
			state,
			lb.CreatedTime.Local().String(),
		})
	}

	return rows
}

// Listeners
type fetchListenersMsg struct {
	Rows []table.Row
	Err  error
}

func fetchListeners(loadBalancerArn *string) tea.Cmd {
	return func() tea.Msg {
		listeners, err := elbv2.DescribeListeners(loadBalancerArn)
		if err != nil {
			return fetchListenersMsg{Err: err}
		}

		return fetchListenersMsg{Rows: getListenerRows(&listeners)}
	}
}

func getListenerRows(listeners *[]types.Listener) []table.Row {
	rows := []table.Row{}
	for _, listener := range *listeners {
		defaultActionType := "Unknown"
		targetInfo := ""

		if len(listener.DefaultActions) > 0 {
			action := listener.DefaultActions[0]
			defaultActionType = string(action.Type)

			// Extract target info based on action type
			switch action.Type {
			case types.ActionTypeEnumForward:
				if action.ForwardConfig != nil && len(action.ForwardConfig.TargetGroups) > 0 {
					targets := []string{}
					for _, tg := range action.ForwardConfig.TargetGroups {
						if tg.TargetGroupArn != nil {
							arnParts := strings.Split(*tg.TargetGroupArn, "/")
							if len(arnParts) > 1 {
								targets = append(targets, arnParts[len(arnParts)-1])
							}
						}
					}
					targetInfo = strings.Join(targets, ", ")
				}
			case types.ActionTypeEnumRedirect:
				if action.RedirectConfig != nil {
					protocol := "Unknown"
					port := "Unknown"
					if action.RedirectConfig.Protocol != nil {
						protocol = *action.RedirectConfig.Protocol
					}
					if action.RedirectConfig.Port != nil {
						port = *action.RedirectConfig.Port
					}
					targetInfo = protocol + ":" + port
				}
			case types.ActionTypeEnumFixedResponse:
				if action.FixedResponseConfig != nil {
					if action.FixedResponseConfig.StatusCode != nil {
						targetInfo = *action.FixedResponseConfig.StatusCode
					}
				}
			}
		}

		rows = append(rows, table.Row{
			string(listener.Protocol),
			fmt.Sprintf("%d", *listener.Port),
			defaultActionType + " → " + targetInfo,
		})
	}

	return rows
}

// Rules
type fetchRulesMsg struct {
	Rows []table.Row
	Err  error
}

func fetchRules(listenerArn *string) tea.Cmd {
	return func() tea.Msg {
		rules, err := elbv2.DescribeRules(listenerArn)
		if err != nil {
			return fetchRulesMsg{Err: err}
		}

		return fetchRulesMsg{Rows: getRuleRows(&rules)}
	}
}

func getRuleRows(rules *[]types.Rule) []table.Row {
	rows := []table.Row{}
	for _, rule := range *rules {
		priority := "default"
		if rule.Priority != nil {
			priority = *rule.Priority
		}

		// Process conditions
		conditionType := ""
		conditionValue := ""
		if len(rule.Conditions) > 0 {
			cond := rule.Conditions[0]
			if cond.Field != nil {
				conditionType = *cond.Field

				// Extract values based on condition type
				if cond.PathPatternConfig != nil && len(cond.PathPatternConfig.Values) > 0 {
					conditionValue = strings.Join(cond.PathPatternConfig.Values, ", ")
				} else if cond.HostHeaderConfig != nil && len(cond.HostHeaderConfig.Values) > 0 {
					conditionValue = strings.Join(cond.HostHeaderConfig.Values, ", ")
				} else if len(cond.Values) > 0 {
					conditionValue = strings.Join(cond.Values, ", ")
				}
			}
		}

		// Process actions
		actionType := "Unknown"
		targetInfo := ""
		if len(rule.Actions) > 0 {
			action := rule.Actions[0]
			actionType = string(action.Type)

			// Extract target info based on action type
			switch action.Type {
			case types.ActionTypeEnumForward:
				if action.ForwardConfig != nil && len(action.ForwardConfig.TargetGroups) > 0 {
					targets := []string{}
					for _, tg := range action.ForwardConfig.TargetGroups {
						if tg.TargetGroupArn != nil {
							arnParts := strings.Split(*tg.TargetGroupArn, "/")
							if len(arnParts) > 1 {
								targets = append(targets, arnParts[len(arnParts)-1])
							}
						}
					}
					targetInfo = strings.Join(targets, ", ")
				}
			case types.ActionTypeEnumRedirect:
				if action.RedirectConfig != nil {
					protocol := "Unknown"
					port := "Unknown"
					if action.RedirectConfig.Protocol != nil {
						protocol = *action.RedirectConfig.Protocol
					}
					if action.RedirectConfig.Port != nil {
						port = *action.RedirectConfig.Port
					}
					targetInfo = protocol + ":" + port
				}
			case types.ActionTypeEnumFixedResponse:
				if action.FixedResponseConfig != nil {
					if action.FixedResponseConfig.StatusCode != nil {
						targetInfo = *action.FixedResponseConfig.StatusCode
					}
				}
			}
		}

		rows = append(rows, table.Row{
			priority,
			conditionType,
			conditionValue,
			actionType,
			targetInfo,
		})
	}

	return rows
}
