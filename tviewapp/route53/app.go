package route53

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jaehong21/hibiscus/internal/aws/route53"
)

// App wires the Route53 data provider into a tview based interface.
type App struct {
	app *tview.Application

	header       *tview.TextView
	pages        *tview.Pages
	zoneFilter   *tview.InputField
	recordFilter *tview.InputField
	zoneTable    *tview.Table
	recordTable  *tview.Table
	statusBar    *tview.TextView
	errorBar     *tview.TextView

	layout tview.Primitive

	hostedZones    []types.HostedZone
	filteredZones  []types.HostedZone
	zoneRowToID    map[int]string
	currentZoneID  string
	hasCurrentZone bool

	records         []types.ResourceRecordSet
	filteredRecords []types.ResourceRecordSet

	focusables   []tview.Primitive
	focusIdx     int
	recordRowMap map[int]int

	modalVisible bool

	mu                  sync.Mutex
	loadHostedZonesOnce sync.Once
}

// NewApp builds the Route53 PoC UI and kicks off data loading.
func NewApp() *App {
	app := tview.NewApplication()

	r := &App{
		app:          app,
		header:       newHeader(),
		zoneFilter:   newInput("Hosted zone filter: "),
		recordFilter: newInput("Record filter: "),
		zoneTable:    newTable(),
		recordTable:  newTable(),
		statusBar:    newStatusBar(),
		errorBar:     newErrorBar(),
		zoneRowToID:  map[int]string{},
		recordRowMap: map[int]int{},
	}

	r.zoneFilter.SetChangedFunc(func(text string) {
		r.applyZoneFilter(text)
	}).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			r.setFocus(r.zoneTable)
		}
	})

	r.recordFilter.SetChangedFunc(func(text string) {
		r.applyRecordFilter(text)
	}).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			r.setFocus(r.recordTable)
		}
	})

	r.zoneTable.SetSelectable(true, false).
		SetSelectedFunc(func(row, _ int) {
			r.handleZoneSelection(row)
		})

	r.recordTable.SetSelectable(true, false)

	r.focusables = []tview.Primitive{
		r.zoneTable,
		r.zoneFilter,
		r.recordTable,
		r.recordFilter,
	}

	r.layout = r.buildLayout()
	r.pages = tview.NewPages().
		AddPage("main", r.layout, true, true)

	r.app.SetInputCapture(r.globalShortcuts)
	r.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		if len(r.hostedZones) == 0 {
			r.setStatus("Fetching hosted zones...")
		}

		r.loadHostedZonesOnce.Do(func() {
			go r.loadHostedZones()
		})

		return false
	})

	return r
}

// Run starts the application event loop.
func (a *App) Run() error {
	a.app.SetRoot(a.pages, true)
	a.app.SetFocus(a.zoneTable)
	return a.app.EnableMouse(true).Run()
}

func (a *App) buildLayout() tview.Primitive {
	hostedZonePanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.zoneFilter, 1, 0, false).
		AddItem(tview.NewBox().SetBorder(false), 0, 0, false).
		AddItem(wrapWithBorder("Hosted Zones", a.zoneTable), 0, 1, true)

	recordPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.recordFilter, 1, 0, false).
		AddItem(tview.NewBox().SetBorder(false), 0, 0, false).
		AddItem(wrapWithBorder("Records", a.recordTable), 0, 1, false)

	content := tview.NewFlex().
		AddItem(hostedZonePanel, 0, 2, true).
		AddItem(recordPanel, 0, 3, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 3, 0, false).
		AddItem(content, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false).
		AddItem(a.errorBar, 2, 0, false)

	return layout
}

func (a *App) globalShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if a.modalVisible {
		if event.Key() == tcell.KeyEsc {
			a.hideModal()
			return nil
		}
		return event
	}

	switch {
	case event.Key() == tcell.KeyCtrlC,
		event.Key() == tcell.KeyCtrlQ,
		event.Rune() == 'q' || event.Rune() == 'Q':
		a.app.Stop()
		return nil

	case event.Key() == tcell.KeyTAB:
		a.focusIdx = (a.focusIdx + 1) % len(a.focusables)
		a.app.SetFocus(a.focusables[a.focusIdx])
		return nil

	case event.Key() == tcell.KeyBacktab:
		a.focusIdx = (a.focusIdx - 1 + len(a.focusables)) % len(a.focusables)
		a.app.SetFocus(a.focusables[a.focusIdx])
		return nil

	case event.Rune() == '/':
		if a.app.GetFocus() == a.zoneTable {
			a.setFocus(a.zoneFilter)
		} else if a.app.GetFocus() == a.recordTable {
			a.setFocus(a.recordFilter)
		}
		return nil

	case event.Rune() == 'R' || event.Rune() == 'r':
		if a.app.GetFocus() == a.zoneTable {
			go a.loadHostedZones()
			return nil
		}
		if a.app.GetFocus() == a.recordTable && a.hasCurrentZone {
			go a.loadRecords(a.currentZoneID)
			return nil
		}

	case event.Rune() == 'e' || event.Rune() == 'E':
		if a.app.GetFocus() == a.recordTable && a.hasCurrentZone {
			row, _ := a.recordTable.GetSelection()
			a.openEditRecordModal(row)
			return nil
		}
	}

	return event
}

func (a *App) setFocus(p tview.Primitive) {
	for idx, prim := range a.focusables {
		if prim == p {
			a.focusIdx = idx
			break
		}
	}
	a.app.SetFocus(p)
}

func (a *App) loadHostedZones() {
	a.setStatus("Fetching hosted zones...")
	a.setError(nil)

	hostedZones, err := route53.ListHostedZones()
	a.app.QueueUpdateDraw(func() {
		if err != nil {
			a.setError(fmt.Errorf("failed to load hosted zones: %w", err))
			a.zoneTable.Clear()
			return
		}

		a.hostedZones = hostedZones
		a.applyZoneFilter(a.zoneFilter.GetText())
		a.setStatus(fmt.Sprintf("Loaded %d hosted zones", len(hostedZones)))
	})
}

func (a *App) applyZoneFilter(query string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	query = strings.TrimSpace(strings.ToLower(query))
	a.filteredZones = a.filteredZones[:0]

	if query == "" {
		a.filteredZones = append(a.filteredZones, a.hostedZones...)
	} else {
		for _, zone := range a.hostedZones {
			if strings.Contains(strings.ToLower(*zone.Name), query) ||
				strings.Contains(strings.ToLower(*zone.Id), query) {
				a.filteredZones = append(a.filteredZones, zone)
			}
		}
	}

	a.refreshZoneTable()
}

func (a *App) refreshZoneTable() {
	a.zoneTable.Clear()
	a.zoneRowToID = map[int]string{}

	headers := []string{"Name", "Record count", "ID"}
	for col, title := range headers {
		a.zoneTable.SetCell(0, col, headerCell(title))
	}

	for idx, zone := range a.filteredZones {
		row := idx + 1
		displayName := strings.TrimSuffix(*zone.Name, ".")
		count := "-"
		if zone.ResourceRecordSetCount != nil {
			count = fmt.Sprintf("%d", *zone.ResourceRecordSetCount)
		}

		a.zoneTable.SetCell(row, 0, tableCell(displayName))
		a.zoneTable.SetCell(row, 1, tableCell(count))
		a.zoneTable.SetCell(row, 2, tableCell(*zone.Id))

		a.zoneRowToID[row] = *zone.Id
	}

	if len(a.filteredZones) == 0 {
		a.zoneTable.SetCell(1, 0, tableCell("No hosted zones found").SetSelectable(false))
	} else {
		a.zoneTable.Select(1, 0)
	}
}

func (a *App) handleZoneSelection(row int) {
	zoneID, ok := a.zoneRowToID[row]
	if !ok {
		return
	}

	a.hasCurrentZone = true
	a.currentZoneID = zoneID
	go a.loadRecords(zoneID)
}

func (a *App) loadRecords(zoneID string) {
	a.setStatus(fmt.Sprintf("Fetching records for %s...", zoneID))
	a.setError(nil)

	id := zoneID

	records, err := route53.ListRecords(&id)
	a.app.QueueUpdateDraw(func() {
		if err != nil {
			a.setError(fmt.Errorf("failed to load records: %w", err))
			a.recordTable.Clear()
			return
		}

		a.records = records
		a.applyRecordFilter(a.recordFilter.GetText())
		a.setStatus(fmt.Sprintf("Loaded %d record sets", len(records)))
	})
}

func (a *App) applyRecordFilter(query string) {
	query = strings.TrimSpace(strings.ToLower(query))
	a.filteredRecords = a.filteredRecords[:0]

	if query == "" {
		a.filteredRecords = append(a.filteredRecords, a.records...)
	} else {
		for _, record := range a.records {
			if recordMatches(record, query) {
				a.filteredRecords = append(a.filteredRecords, record)
			}
		}
	}

	a.refreshRecordTable()
}

func (a *App) refreshRecordTable() {
	a.recordTable.Clear()
	a.recordRowMap = map[int]int{}

	headers := []string{"Record Name", "Type", "Value / Alias", "TTL", "Weight"}
	for col, title := range headers {
		a.recordTable.SetCell(0, col, headerCell(title))
	}

	if len(a.filteredRecords) == 0 {
		msg := "Select a hosted zone to load records"
		if a.hasCurrentZone && len(a.records) > 0 {
			msg = "No records matched this filter"
		}
		a.recordTable.SetCell(1, 0, tableCell(msg).SetSelectable(false))
		return
	}

	row := 1
	for idx, record := range a.filteredRecords {
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
			a.recordTable.SetCell(row, 0, tableCell(strings.TrimSuffix(*record.Name, ".")))
			a.recordTable.SetCell(row, 1, tableCell(string(record.Type)))
			a.recordTable.SetCell(row, 2, tableCell(val))
			a.recordTable.SetCell(row, 3, tableCell(ttl))
			a.recordTable.SetCell(row, 4, tableCell(weight))
			a.recordRowMap[row] = idx
			row++
		}
	}
}

func newHeader() *tview.TextView {
	text := `[yellow]Route53 (tview PoC)[-]  Tab: cycle focus  /: focus filter  R: refresh  Ctrl+C: quit`
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	tv.SetBorder(false)
	tv.SetTextAlign(tview.AlignLeft)
	return tv
}

func newInput(label string) *tview.InputField {
	return tview.NewInputField().
		SetLabel(label).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)
}

func newTable() *tview.Table {
	table := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)
	table.SetBorder(true)
	table.SetBorderColor(tcell.ColorDimGray)
	return table
}

func newStatusBar() *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[lightgreen]Ready[-]")
	tv.SetBorder(false)
	return tv
}

func newErrorBar() *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true)
	tv.SetBorder(false)
	return tv
}

func wrapWithBorder(title string, content tview.Primitive) tview.Primitive {
	return tview.NewFrame(content).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText(title, true, tview.AlignLeft, tcell.ColorLightCyan)
}

func headerCell(title string) *tview.TableCell {
	return tview.NewTableCell(title).
		SetTextColor(tcell.ColorLightCyan).
		SetSelectable(false).
		SetAlign(tview.AlignCenter)
}

func tableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetMaxWidth(0).
		SetExpansion(1)
}

func formatRecordValues(record types.ResourceRecordSet) []string {
	if len(record.ResourceRecords) > 0 {
		values := make([]string, len(record.ResourceRecords))
		for i, rr := range record.ResourceRecords {
			values[i] = strings.TrimSpace(*rr.Value)
		}
		return values
	}

	if record.AliasTarget != nil {
		var aliasType string
		switch {
		case route53.IsCloudFrontAlias(record.AliasTarget.HostedZoneId):
			aliasType = "CloudFront"
		case route53.IsELBAlias(record.AliasTarget.HostedZoneId):
			aliasType = "ELB"
		default:
			aliasType = "Alias"
		}

		value := fmt.Sprintf("%s (%s) -> %s",
			aliasType,
			deref(record.AliasTarget.HostedZoneId),
			deref(record.AliasTarget.DNSName),
		)
		return []string{value}
	}

	return []string{"-"}
}

func recordMatches(record types.ResourceRecordSet, query string) bool {
	if query == "" {
		return true
	}

	name := strings.ToLower(strings.TrimSuffix(*record.Name, "."))
	if strings.Contains(name, query) {
		return true
	}

	if len(record.ResourceRecords) > 0 {
		for _, rr := range record.ResourceRecords {
			if strings.Contains(strings.ToLower(*rr.Value), query) {
				return true
			}
		}
	}

	if record.AliasTarget != nil {
		alias := strings.ToLower(fmt.Sprintf("%s %s", deref(record.AliasTarget.HostedZoneId), deref(record.AliasTarget.DNSName)))
		if strings.Contains(alias, query) {
			return true
		}
	}

	return false
}

func deref(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (a *App) setStatus(msg string) {
	a.statusBar.SetText(fmt.Sprintf("[lightgreen]%s[-]", msg))
}

func (a *App) setError(err error) {
	if err == nil {
		a.errorBar.SetText("")
		return
	}
	a.errorBar.SetText(fmt.Sprintf("[red]Error: %s[-]", err.Error()))
}

func (a *App) recordFromRow(row int) (*types.ResourceRecordSet, bool) {
	idx, ok := a.recordRowMap[row]
	if !ok {
		return nil, false
	}
	if idx < 0 || idx >= len(a.filteredRecords) {
		return nil, false
	}
	return &a.filteredRecords[idx], true
}

func (a *App) openEditRecordModal(row int) {
	if !a.hasCurrentZone {
		a.setStatus("Select a hosted zone first")
		return
	}

	record, ok := a.recordFromRow(row)
	if !ok {
		a.setStatus("Select a record to edit")
		return
	}

	if record.AliasTarget != nil || len(record.ResourceRecords) == 0 {
		a.setError(errors.New("editing alias records isn't supported in this PoC"))
		return
	}

	valueField := tview.NewInputField().
		SetLabel("Values (comma separated) ").
		SetText(strings.Join(recordValueStrings(*record), ", "))
	ttlField := tview.NewInputField().
		SetLabel("TTL (seconds) ")
	if record.TTL != nil {
		ttlField.SetText(fmt.Sprintf("%d", *record.TTL))
	}

	form := tview.NewForm().
		AddFormItem(valueField).
		AddFormItem(ttlField).
		AddButton("Save", func() {
			a.submitRecordEdit(*record, valueField.GetText(), ttlField.GetText())
		}).
		AddButton("Cancel", func() {
			a.hideModal()
		})
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf("Edit %s (%s)", strings.TrimSuffix(*record.Name, "."), record.Type))
	form.SetTitleAlign(tview.AlignLeft)
	form.SetButtonsAlign(tview.AlignCenter)

	helper := tview.NewTextView().
		SetDynamicColors(true).
		SetText("Separate multiple values with commas. Leave TTL blank to keep the current value.")
	helper.SetBorder(false)

	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(helper, 2, 0, false)

	a.showModal(content, 80, 14)
	a.app.SetFocus(valueField)
}

func (a *App) submitRecordEdit(original types.ResourceRecordSet, valuesRaw, ttlRaw string) {
	values, err := parseRecordValues(valuesRaw)
	if err != nil {
		a.setError(err)
		return
	}

	ttl, err := parseTTL(ttlRaw, original.TTL)
	if err != nil {
		a.setError(err)
		return
	}

	updated := original
	updated.ResourceRecords = make([]types.ResourceRecord, len(values))
	for i, val := range values {
		updated.ResourceRecords[i] = types.ResourceRecord{Value: aws.String(val)}
	}
	updated.TTL = ttl

	zoneID := a.currentZoneID
	a.setStatus("Updating record...")

	go func(record types.ResourceRecordSet, hostedZoneID string) {
		id := hostedZoneID
		err := route53.UpsertRecord(&id, record)
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.setError(fmt.Errorf("failed to update record: %w", err))
				a.setStatus("Update failed")
				return
			}

			a.setError(nil)
			a.setStatus("Record updated")
			a.hideModal()
			go a.loadRecords(hostedZoneID)
		})
	}(updated, zoneID)
}

func parseRecordValues(input string) ([]string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, errors.New("record value cannot be empty")
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return nil, errors.New("record value cannot be empty")
	}

	return values, nil
}

func parseTTL(input string, fallback *int64) (*int64, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("invalid TTL: %s", value)
	}

	return aws.Int64(parsed), nil
}

func recordValueStrings(record types.ResourceRecordSet) []string {
	values := make([]string, 0, len(record.ResourceRecords))
	for _, rr := range record.ResourceRecords {
		if rr.Value != nil {
			values = append(values, strings.TrimSpace(*rr.Value))
		}
	}
	return values
}

func (a *App) showModal(content tview.Primitive, width, height int) {
	if a.modalVisible {
		a.pages.RemovePage("modal")
	}
	a.modalVisible = true
	a.pages.AddPage("modal", centerPrimitive(content, width, height), true, true)
}

func (a *App) hideModal() {
	if !a.modalVisible {
		return
	}
	a.modalVisible = false
	a.pages.RemovePage("modal")
	a.setFocus(a.recordTable)
}

func centerPrimitive(content tview.Primitive, width, height int) tview.Primitive {
	grid := tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(content, 1, 1, 1, 1, 0, 0, true)
	grid.SetBackgroundColor(tcell.ColorBlack)
	return grid
}
