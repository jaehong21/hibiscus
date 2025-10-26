package elb

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jaehong21/hibiscus/internal/aws/elbv2"
	"github.com/jaehong21/hibiscus/tviewapp/hibiscus"
)

type tab int

const (
	lbTab tab = iota
	listenerTab
	ruleTab
)

// Service implements the hibiscus.Service interface for the ELB view.
type Service struct {
	ctx hibiscus.ServiceContext

	layout *tview.Flex
	pages  *tview.Pages
	filter *tview.InputField

	lbTable       *tview.Table
	listenerTable *tview.Table
	ruleTable     *tview.Table

	current tab

	loadBalancers         []types.LoadBalancer
	filteredLoadBalancers []types.LoadBalancer
	listeners             []types.Listener
	rules                 []types.Rule

	selectedLoadBalancerArn  string
	selectedLoadBalancerName string
	selectedListenerArn      string

	active bool
}

func New(ctx hibiscus.ServiceContext) hibiscus.Service {
	svc := &Service{ctx: ctx, current: lbTab}

	svc.filter = tview.NewInputField().
		SetLabel("Filter (/): ").
		SetFieldBackgroundColor(tcell.ColorBlack)

	svc.lbTable = buildTable("Load balancers")
	svc.listenerTable = buildTable("Listeners")
	svc.ruleTable = buildTable("Rules")

	svc.pages = tview.NewPages()
	svc.pages.AddPage("lbs", svc.lbTable, true, true)
	svc.pages.AddPage("listeners", svc.listenerTable, true, false)
	svc.pages.AddPage("rules", svc.ruleTable, true, false)

	svc.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(svc.filter, 1, 0, true).
		AddItem(svc.pages, 0, 1, true)

	svc.filter.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			svc.applyFilter(strings.TrimSpace(svc.filter.GetText()))
		case tcell.KeyEsc:
			svc.exitFilterMode()
		}
	})

	return svc
}

func (s *Service) Name() string { return "elb" }
func (s *Service) Title() string {
	return "Elastic Load Balancing – load balancers › listeners › rules"
}
func (s *Service) Primitive() tview.Primitive {
	return s.layout
}

func (s *Service) Init() {
	s.loadLoadBalancers()
}

func (s *Service) Activate() {
	s.active = true
	s.focusCurrentTable()
}

func (s *Service) Deactivate() {
	s.active = false
}

func (s *Service) Refresh() {
	switch s.current {
	case listenerTab:
		if s.selectedLoadBalancerArn != "" {
			s.loadListeners(s.selectedLoadBalancerArn)
			return
		}
	case ruleTab:
		if s.selectedListenerArn != "" {
			s.loadRules(s.selectedListenerArn)
			return
		}
	}
	s.loadLoadBalancers()
}

func (s *Service) EnterFilterMode() bool {
	if !s.canFocus() {
		return false
	}
	if s.current != lbTab {
		s.ctx.SetStatus("Filtering only applies to load balancers")
		return false
	}
	s.ctx.App.SetFocus(s.filter)
	return true
}

func (s *Service) InFilterMode() bool {
	return s.filter.HasFocus()
}

func (s *Service) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event == nil {
		return nil
	}

	switch event.Key() {
	case tcell.KeyEsc:
		if s.filter.HasFocus() {
			s.exitFilterMode()
			return nil
		}
		switch s.current {
		case ruleTab:
			s.showListenerTab()
			return nil
		case listenerTab:
			s.showLoadBalancerTab()
			return nil
		}
	case tcell.KeyEnter:
		if s.filter.HasFocus() {
			s.applyFilter(strings.TrimSpace(s.filter.GetText()))
			return nil
		}
		switch s.current {
		case lbTab:
			s.openSelectedLoadBalancer()
			return nil
		case listenerTab:
			s.openSelectedListener()
			return nil
		}
	}

	return event
}

func (s *Service) exitFilterMode() {
	s.filter.SetText("")
	switch s.current {
	case listenerTab:
		s.setFocus(s.listenerTable)
	case ruleTab:
		s.setFocus(s.ruleTable)
	default:
		s.setFocus(s.lbTable)
	}
}

func (s *Service) applyFilter(query string) {
	if s.current != lbTab {
		s.filter.SetText("")
		s.exitFilterMode()
		return
	}

	query = strings.ToLower(strings.TrimSpace(query))
	s.filteredLoadBalancers = s.filteredLoadBalancers[:0]
	if query == "" {
		s.filteredLoadBalancers = append(s.filteredLoadBalancers, s.loadBalancers...)
		s.renderLoadBalancers()
		s.exitFilterMode()
		return
	}

	for _, lb := range s.loadBalancers {
		if lb.LoadBalancerName == nil {
			continue
		}
		if strings.Contains(strings.ToLower(*lb.LoadBalancerName), query) {
			s.filteredLoadBalancers = append(s.filteredLoadBalancers, lb)
		}
	}

	s.renderLoadBalancers()
	s.filter.SetText("")
	s.exitFilterMode()
}

func (s *Service) openSelectedLoadBalancer() {
	row, _ := s.lbTable.GetSelection()
	if row <= 0 || row-1 >= len(s.filteredLoadBalancers) {
		return
	}
	lb := s.filteredLoadBalancers[row-1]
	if lb.LoadBalancerArn == nil {
		return
	}
	s.selectedLoadBalancerArn = *lb.LoadBalancerArn
	s.selectedLoadBalancerName = valueOr(lb.LoadBalancerName)
	s.loadListeners(s.selectedLoadBalancerArn)
}

func (s *Service) openSelectedListener() {
	row, _ := s.listenerTable.GetSelection()
	if row <= 0 || row-1 >= len(s.listeners) {
		return
	}
	listener := s.listeners[row-1]
	if listener.ListenerArn == nil {
		return
	}
	s.selectedListenerArn = *listener.ListenerArn
	s.loadRules(s.selectedListenerArn)
}

func (s *Service) loadLoadBalancers() {
	s.ctx.SetStatus("Fetching load balancers...")
	s.ctx.SetError(nil)

	go func() {
		lbs, err := elbv2.DescribeLoadBalancers()
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("describe load balancers: %w", err))
				return
			}
			s.loadBalancers = lbs
			s.filteredLoadBalancers = append([]types.LoadBalancer(nil), lbs...)
			s.renderLoadBalancers()
			s.showLoadBalancerTab()
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d load balancers", len(lbs)))
		})
	}()
}

func (s *Service) loadListeners(loadBalancerArn string) {
	if loadBalancerArn == "" {
		return
	}
	s.ctx.SetStatus("Fetching listeners...")
	s.ctx.SetError(nil)

	arn := loadBalancerArn
	go func() {
		listeners, err := elbv2.DescribeListeners(&arn)
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("describe listeners: %w", err))
				return
			}
			s.listeners = listeners
			s.renderListeners()
			s.showListenerTab()
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d listeners", len(listeners)))
		})
	}()
}

func (s *Service) loadRules(listenerArn string) {
	if listenerArn == "" {
		return
	}
	s.ctx.SetStatus("Fetching listener rules...")
	s.ctx.SetError(nil)

	arn := listenerArn
	go func() {
		rules, err := elbv2.DescribeRules(&arn)
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("describe rules: %w", err))
				return
			}
			s.rules = rules
			s.renderRules()
			s.showRuleTab()
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d rules", len(rules)))
		})
	}()
}

func (s *Service) renderLoadBalancers() {
	table := s.lbTable
	table.Clear()

	headers := []string{"Name", "Type", "DNS name", "State", "Created"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.filteredLoadBalancers) == 0 {
		table.SetCell(1, 0, tableCell("No load balancers found").SetSelectable(false))
		return
	}

	for idx, lb := range s.filteredLoadBalancers {
		table.SetCell(idx+1, 0, tableCell(valueOr(lb.LoadBalancerName)))
		lbType := "Unknown"
		if lb.Type != "" {
			lbType = string(lb.Type)
		}
		table.SetCell(idx+1, 1, tableCell(lbType))
		table.SetCell(idx+1, 2, tableCell(valueOr(lb.DNSName)))
		state := "Unknown"
		if lb.State != nil && lb.State.Code != "" {
			state = string(lb.State.Code)
		}
		table.SetCell(idx+1, 3, tableCell(state))
		created := ""
		if lb.CreatedTime != nil {
			created = lb.CreatedTime.Local().Format("2006-01-02 15:04:05")
		}
		table.SetCell(idx+1, 4, tableCell(created))
	}

	table.Select(1, 0)
}

func (s *Service) renderListeners() {
	table := s.listenerTable
	table.Clear()

	headers := []string{"Protocol", "Port", "Default action"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.listeners) == 0 {
		table.SetCell(1, 0, tableCell("No listeners found").SetSelectable(false))
		return
	}

	for idx, listener := range s.listeners {
		protocol := string(listener.Protocol)
		port := ""
		if listener.Port != nil {
			port = fmt.Sprintf("%d", *listener.Port)
		}
		action := summarizeAction(listener.DefaultActions)

		table.SetCell(idx+1, 0, tableCell(protocol))
		table.SetCell(idx+1, 1, tableCell(port))
		table.SetCell(idx+1, 2, tableCell(action))
	}

	table.Select(1, 0)
}

func (s *Service) renderRules() {
	table := s.ruleTable
	table.Clear()

	headers := []string{"Priority", "Condition", "Value", "Action", "Target"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.rules) == 0 {
		table.SetCell(1, 0, tableCell("No rules found").SetSelectable(false))
		return
	}

	for idx, rule := range s.rules {
		priority := "default"
		if rule.Priority != nil {
			priority = *rule.Priority
		}
		condType, condValue := summarizeCondition(rule.Conditions)
		actionType, target := summarizeRuleAction(rule.Actions)

		table.SetCell(idx+1, 0, tableCell(priority))
		table.SetCell(idx+1, 1, tableCell(condType))
		table.SetCell(idx+1, 2, tableCell(condValue))
		table.SetCell(idx+1, 3, tableCell(actionType))
		table.SetCell(idx+1, 4, tableCell(target))
	}

	table.Select(1, 0)
}

func (s *Service) showLoadBalancerTab() {
	s.current = lbTab
	s.pages.SwitchToPage("lbs")
	s.lbTable.SetTitle("Load balancers")
	s.setFocus(s.lbTable)
}

func (s *Service) showListenerTab() {
	s.current = listenerTab
	title := "Listeners"
	if s.selectedLoadBalancerName != "" {
		title = fmt.Sprintf("Listeners for %s", s.selectedLoadBalancerName)
	}
	s.listenerTable.SetTitle(title)
	s.pages.SwitchToPage("listeners")
	s.setFocus(s.listenerTable)
}

func (s *Service) showRuleTab() {
	s.current = ruleTab
	title := "Rules"
	if s.selectedLoadBalancerName != "" {
		title = fmt.Sprintf("Rules for %s", s.selectedLoadBalancerName)
	}
	s.ruleTable.SetTitle(title)
	s.pages.SwitchToPage("rules")
	s.setFocus(s.ruleTable)
}

func summarizeAction(actions []types.Action) string {
	if len(actions) == 0 {
		return "-"
	}
	action := actions[0]
	info := string(action.Type)
	switch action.Type {
	case types.ActionTypeEnumForward:
		if action.ForwardConfig != nil && len(action.ForwardConfig.TargetGroups) > 0 {
			names := []string{}
			for _, tg := range action.ForwardConfig.TargetGroups {
				if tg.TargetGroupArn != nil {
					parts := strings.Split(*tg.TargetGroupArn, "/")
					names = append(names, parts[len(parts)-1])
				}
			}
			if len(names) > 0 {
				info = fmt.Sprintf("Forward → %s", strings.Join(names, ", "))
			} else {
				info = "Forward"
			}
		} else {
			info = "Forward"
		}
	case types.ActionTypeEnumRedirect:
		if action.RedirectConfig != nil {
			proto := valueOr(action.RedirectConfig.Protocol)
			port := valueOr(action.RedirectConfig.Port)
			info = fmt.Sprintf("Redirect → %s:%s", proto, port)
		}
	case types.ActionTypeEnumFixedResponse:
		if action.FixedResponseConfig != nil && action.FixedResponseConfig.StatusCode != nil {
			info = fmt.Sprintf("Fixed %s", *action.FixedResponseConfig.StatusCode)
		}
	}
	return info
}

func summarizeCondition(conditions []types.RuleCondition) (string, string) {
	if len(conditions) == 0 {
		return "-", "-"
	}
	cond := conditions[0]
	kind := valueOr(cond.Field)
	switch {
	case cond.PathPatternConfig != nil && len(cond.PathPatternConfig.Values) > 0:
		return kind, strings.Join(cond.PathPatternConfig.Values, ", ")
	case cond.HostHeaderConfig != nil && len(cond.HostHeaderConfig.Values) > 0:
		return kind, strings.Join(cond.HostHeaderConfig.Values, ", ")
	case len(cond.Values) > 0:
		return kind, strings.Join(cond.Values, ", ")
	default:
		return kind, "-"
	}
}

func summarizeRuleAction(actions []types.Action) (string, string) {
	if len(actions) == 0 {
		return "-", "-"
	}
	action := actions[0]
	target := "-"
	switch action.Type {
	case types.ActionTypeEnumForward:
		if action.ForwardConfig != nil && len(action.ForwardConfig.TargetGroups) > 0 {
			names := []string{}
			for _, tg := range action.ForwardConfig.TargetGroups {
				if tg.TargetGroupArn != nil {
					parts := strings.Split(*tg.TargetGroupArn, "/")
					names = append(names, parts[len(parts)-1])
				}
			}
			if len(names) > 0 {
				target = strings.Join(names, ", ")
			}
		}
	case types.ActionTypeEnumRedirect:
		if action.RedirectConfig != nil {
			proto := valueOr(action.RedirectConfig.Protocol)
			port := valueOr(action.RedirectConfig.Port)
			target = fmt.Sprintf("%s:%s", proto, port)
		}
	case types.ActionTypeEnumFixedResponse:
		if action.FixedResponseConfig != nil && action.FixedResponseConfig.StatusCode != nil {
			target = *action.FixedResponseConfig.StatusCode
		}
	}
	return string(action.Type), target
}

func headerCell(title string) *tview.TableCell {
	return tview.NewTableCell(title).
		SetSelectable(false).
		SetTextColor(tcell.ColorLightCyan).
		SetAlign(tview.AlignLeft)
}

func tableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetMaxWidth(0).
		SetExpansion(1)
}

func buildTable(title string) *tview.Table {
	tbl := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)
	tbl.SetBorder(true)
	tbl.SetTitle(title)
	tbl.SetBorderColor(tcell.ColorDimGray)
	return tbl
}

func valueOr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (s *Service) canFocus() bool {
	return s.ctx.App != nil && s.active
}

func (s *Service) setFocus(p tview.Primitive) {
	if !s.canFocus() || p == nil {
		return
	}
	s.ctx.App.SetFocus(p)
}

func (s *Service) focusCurrentTable() {
	switch s.current {
	case listenerTab:
		s.setFocus(s.listenerTable)
	case ruleTab:
		s.setFocus(s.ruleTable)
	default:
		s.setFocus(s.lbTable)
	}
}
