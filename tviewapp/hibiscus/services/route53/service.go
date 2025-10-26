package route53

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	awsr53 "github.com/jaehong21/hibiscus/internal/aws/route53"
	"github.com/jaehong21/hibiscus/tviewapp/hibiscus"
)

type tab int

const (
	zoneTab tab = iota
	recordTab
)

// Service implements the hibiscus.Service interface for Amazon Route53.
type Service struct {
	ctx hibiscus.ServiceContext

	layout    *tview.Flex
	pages     *tview.Pages
	filter    *tview.InputField
	zoneTable *tview.Table
	recTable  *tview.Table

	current tab

	zones           []types.HostedZone
	filteredZones   []types.HostedZone
	records         []types.ResourceRecordSet
	filteredRecords []types.ResourceRecordSet
	recordRowMap    map[int]int

	currentZoneID   string
	currentZoneName string

	mu sync.Mutex
}

func New(ctx hibiscus.ServiceContext) hibiscus.Service {
	svc := &Service{
		ctx:          ctx,
		current:      zoneTab,
		recordRowMap: map[int]int{},
	}

	svc.filter = tview.NewInputField().
		SetLabel("Filter (/): ").
		SetFieldBackgroundColor(tcell.ColorBlack)

	svc.zoneTable = buildTable("Route53 hosted zones")
	svc.recTable = buildTable("Hosted zone records")

	svc.pages = tview.NewPages()
	svc.pages.AddPage("zones", svc.zoneTable, true, true)
	svc.pages.AddPage("records", svc.recTable, true, false)

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

func (s *Service) Name() string  { return "route53" }
func (s *Service) Title() string { return "Amazon Route53 – hosted zones › records" }
func (s *Service) Primitive() tview.Primitive {
	return s.layout
}

func (s *Service) Init() {
	s.loadHostedZones()
}

func (s *Service) Activate() {
	if s.ctx.App == nil {
		return
	}
	if s.current == recordTab {
		s.ctx.App.SetFocus(s.recTable)
	} else {
		s.ctx.App.SetFocus(s.zoneTable)
	}
}

func (s *Service) Deactivate() {}

func (s *Service) Refresh() {
	if s.current == recordTab && s.currentZoneID != "" {
		s.loadRecords(s.currentZoneID)
		return
	}
	s.loadHostedZones()
}

func (s *Service) EnterFilterMode() bool {
	if s.ctx.App == nil {
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
		if s.current == recordTab {
			s.showZoneTab()
			return nil
		}
	case tcell.KeyEnter:
		if s.zoneTable.HasFocus() {
			s.openSelectedZone()
			return nil
		}
	}

	return event
}

func (s *Service) exitFilterMode() {
	s.filter.SetText("")
	if s.ctx.App == nil {
		return
	}
	if s.current == recordTab {
		s.ctx.App.SetFocus(s.recTable)
	} else {
		s.ctx.App.SetFocus(s.zoneTable)
	}
}

func (s *Service) openSelectedZone() {
	row, _ := s.zoneTable.GetSelection()
	if row <= 0 || row-1 >= len(s.filteredZones) {
		return
	}
	zone := s.filteredZones[row-1]
	if zone.Id == nil {
		return
	}
	s.currentZoneID = aws.ToString(zone.Id)
	s.currentZoneName = aws.ToString(zone.Name)
	s.loadRecords(s.currentZoneID)
}

func (s *Service) loadHostedZones() {
	s.ctx.SetStatus("Fetching hosted zones...")
	s.ctx.SetError(nil)

	go func() {
		zones, err := awsr53.ListHostedZones()
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("list hosted zones: %w", err))
				return
			}
			s.mu.Lock()
			s.zones = zones
			s.filteredZones = append([]types.HostedZone(nil), zones...)
			s.mu.Unlock()
			s.renderZones()
			s.showZoneTab()
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d hosted zones", len(zones)))
		})
	}()
}

func (s *Service) loadRecords(zoneID string) {
	if zoneID == "" {
		return
	}
	s.ctx.SetStatus(fmt.Sprintf("Fetching records for %s...", s.currentZoneName))
	s.ctx.SetError(nil)

	id := zoneID
	go func() {
		records, err := awsr53.ListRecords(&id)
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("list records: %w", err))
				return
			}
			s.mu.Lock()
			s.records = records
			s.filteredRecords = append([]types.ResourceRecordSet(nil), records...)
			s.mu.Unlock()
			s.renderRecords()
			s.showRecordTab()
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d records", len(records)))
		})
	}()
}

func (s *Service) applyFilter(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	if s.current == recordTab {
		if s.currentZoneID == "" {
			return
		}
		s.filteredRecords = s.filteredRecords[:0]
		if query == "" {
			s.filteredRecords = append(s.filteredRecords, s.records...)
		} else {
			for _, record := range s.records {
				if recordMatches(record, query) {
					s.filteredRecords = append(s.filteredRecords, record)
				}
			}
		}
		s.renderRecords()
	} else {
		s.filteredZones = s.filteredZones[:0]
		if query == "" {
			s.filteredZones = append(s.filteredZones, s.zones...)
		} else {
			for _, zone := range s.zones {
				if zoneMatches(zone, query) {
					s.filteredZones = append(s.filteredZones, zone)
				}
			}
		}
		s.renderZones()
	}

	s.filter.SetText("")
	s.exitFilterMode()
}

func (s *Service) renderZones() {
	table := s.zoneTable
	table.Clear()

	headers := []string{"Name", "Record count", "ID"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.filteredZones) == 0 {
		table.SetCell(1, 0, tableCell("No hosted zones found").SetSelectable(false))
		return
	}

	for idx, zone := range s.filteredZones {
		name := aws.ToString(zone.Name)
		recordCount := ""
		if zone.ResourceRecordSetCount != nil {
			recordCount = fmt.Sprintf("%d", *zone.ResourceRecordSetCount)
		}
		table.SetCell(idx+1, 0, tableCell(name))
		table.SetCell(idx+1, 1, tableCell(recordCount))
		table.SetCell(idx+1, 2, tableCell(aws.ToString(zone.Id)))
	}

	table.Select(1, 0)
}

func (s *Service) renderRecords() {
	table := s.recTable
	table.Clear()
	s.recordRowMap = map[int]int{}

	headers := []string{"Record name", "Type", "Value", "TTL", "Weight"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.filteredRecords) == 0 {
		msg := "Select a hosted zone"
		if s.currentZoneID != "" && len(s.records) > 0 {
			msg = "No records match this filter"
		}
		table.SetCell(1, 0, tableCell(msg).SetSelectable(false))
		return
	}

	row := 1
	for idx, record := range s.filteredRecords {
		values := formatRecordValues(record)
		ttl := ""
		if record.TTL != nil {
			ttl = fmt.Sprintf("%d", *record.TTL)
		}
		weight := "-"
		if record.Weight != nil {
			weight = fmt.Sprintf("%d", *record.Weight)
		}

		for _, val := range values {
			table.SetCell(row, 0, tableCell(trimDot(aws.ToString(record.Name))))
			table.SetCell(row, 1, tableCell(string(record.Type)))
			table.SetCell(row, 2, tableCell(val))
			table.SetCell(row, 3, tableCell(ttl))
			table.SetCell(row, 4, tableCell(weight))
			s.recordRowMap[row] = idx
			row++
		}
	}

	table.Select(1, 0)
}

func (s *Service) showZoneTab() {
	s.current = zoneTab
	s.pages.SwitchToPage("zones")
	s.zoneTable.SetTitle("Route53 hosted zones")
	if s.ctx.App != nil {
		s.ctx.App.SetFocus(s.zoneTable)
	}
}

func (s *Service) showRecordTab() {
	s.current = recordTab
	title := "Hosted zone records"
	if s.currentZoneName != "" {
		title = fmt.Sprintf("Records for %s", strings.TrimSuffix(s.currentZoneName, "."))
	}
	s.recTable.SetTitle(title)
	s.pages.SwitchToPage("records")
	if s.ctx.App != nil {
		s.ctx.App.SetFocus(s.recTable)
	}
}

func zoneMatches(zone types.HostedZone, query string) bool {
	name := strings.ToLower(aws.ToString(zone.Name))
	id := strings.ToLower(aws.ToString(zone.Id))
	return strings.Contains(name, query) || strings.Contains(id, query)
}

func recordMatches(record types.ResourceRecordSet, query string) bool {
	if query == "" {
		return true
	}
	if strings.Contains(strings.ToLower(trimDot(aws.ToString(record.Name))), query) {
		return true
	}
	if len(record.ResourceRecords) > 0 {
		for _, rr := range record.ResourceRecords {
			if rr.Value != nil && strings.Contains(strings.ToLower(*rr.Value), query) {
				return true
			}
		}
	}
	if record.AliasTarget != nil {
		alias := strings.ToLower(fmt.Sprintf("%s %s", aws.ToString(record.AliasTarget.HostedZoneId), aws.ToString(record.AliasTarget.DNSName)))
		if strings.Contains(alias, query) {
			return true
		}
	}
	return false
}

func formatRecordValues(record types.ResourceRecordSet) []string {
	if len(record.ResourceRecords) > 0 {
		values := make([]string, 0, len(record.ResourceRecords))
		for _, rr := range record.ResourceRecords {
			if rr.Value != nil {
				values = append(values, strings.TrimSpace(*rr.Value))
			}
		}
		if len(values) > 0 {
			return values
		}
	}

	if record.AliasTarget != nil {
		aliasType := "Alias"
		switch {
		case awsr53.IsCloudFrontAlias(record.AliasTarget.HostedZoneId):
			aliasType = "CloudFront"
		case awsr53.IsELBAlias(record.AliasTarget.HostedZoneId):
			aliasType = "ELB"
		}
		value := fmt.Sprintf("%s (%s) -> %s", aliasType, aws.ToString(record.AliasTarget.HostedZoneId), aws.ToString(record.AliasTarget.DNSName))
		return []string{value}
	}

	return []string{"-"}
}

func trimDot(value string) string {
	return strings.TrimSuffix(value, ".")
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
