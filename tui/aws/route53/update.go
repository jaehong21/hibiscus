package route53

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Enable dynamic column width sizing
		tableWidth := msg.Width - 2

		// Set overall table widths
		m.table.hostedZone.SetWidth(tableWidth)
		m.table.record.SetWidth(tableWidth)

		// Set table heights
		m.table.hostedZone.SetHeight(msg.Height - 12)
		m.table.record.SetHeight(msg.Height - 12)

		// Update hosted zone table column widths with new proportional columns
		nameWidth := int(float64(tableWidth) * 0.25)        // 25% of width
		recordCountWidth := int(float64(tableWidth) * 0.15) // 15% of width
		idWidth := int(float64(tableWidth) * 0.55)          // 55% of width

		// Create new hosted zone columns with dynamic widths
		hostedZoneColumns := []table.Column{
			{Title: "Name", Width: nameWidth},
			{Title: "Record count", Width: recordCountWidth},
			{Title: "ID", Width: idWidth},
		}
		m.table.hostedZone.SetColumns(hostedZoneColumns)

		// Update record table column widths with new proportional columns
		recordNameWidth := int(float64(tableWidth) * 0.30) // 30% of width
		typeWidth := int(float64(tableWidth) * 0.07)       // 7% of width
		valuesWidth := int(float64(tableWidth) * 0.45)     // 45% of width
		ttlWidth := int(float64(tableWidth) * 0.08)        // 8% of width
		weightWidth := int(float64(tableWidth) * 0.05)     // 5% of width

		// Create new record columns with dynamic widths
		recordColumns := []table.Column{
			{Title: "Record Name", Width: recordNameWidth},
			{Title: "Type", Width: typeWidth},
			{Title: "Values", Width: valuesWidth},
			{Title: "TTL", Width: ttlWidth},
			{Title: "Weight", Width: weightWidth},
		}
		m.table.record.SetColumns(recordColumns)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(): // Quit
			return m, tea.Quit

		case "Q", "q", tea.KeyEsc.String():
			if m.textinput.Focused() { // Exit from search input
				m.textinput.Blur()
				m.textinput.Reset()

				// Focus the appropriate table based on current tab
				if m.tab == ROUTE53_HOSTED_ZONE_TAB {
					m.table.hostedZone.Focus()
				} else if m.tab == ROUTE53_RECORD_TAB {
					m.table.record.Focus()
				}
				return m, nil
			}

			if m.tab == ROUTE53_RECORD_TAB { // TAB: ECR_IMAGE_TAB -> ECR_REPO_TAB
				m.tab = ROUTE53_HOSTED_ZONE_TAB
				m.table.hostedZone.Focus()
				m.table.record.Blur()
				return m, nil
			}

		case "R", "r":
			if m.table.hostedZone.Focused() { // Refresh ECR_REPO_TAB
				m.loading.msg = FETCHING_HOSTED_ZONES_MSG
				return m, tea.Batch(fetchHostedZone(), m.loading.spinner.Tick)
			}
			if m.table.record.Focused() { // Refresh ECR_IMAGE_TAB
				m.loading.msg = FETCHING_RECORDS_MSG
				return m, tea.Batch(fetchRecords(&m.table.hostedZone.SelectedRow()[2]), m.loading.spinner.Tick)
			}

		case "/": // Search
			if !m.textinput.Focused() {
				m.textinput.Focus()
				m.table.hostedZone.Blur()
				m.table.record.Blur()
				return m, nil
			}

		case tea.KeyEnter.String():
			if m.table.hostedZone.Focused() { // TAB: ECR_REPO_TAB -> ECR_IMAGE_TAB
				m.tab = ROUTE53_RECORD_TAB
				m.loading.msg = FETCHING_RECORDS_MSG

				m.table.hostedZone.Blur()
				m.table.record.Focus()
				return m, tea.Batch(
					fetchRecords(&m.table.hostedZone.SelectedRow()[2]),
					m.loading.spinner.Tick,
				)
			}

			if m.textinput.Focused() {
				if m.tab == ROUTE53_HOSTED_ZONE_TAB { // Filter ROUTE53_HOSTED_ZONE
					m.loading.msg = FILTERING_HOSTED_ZONES_MSG
					m.textinput.Blur()
					m.table.hostedZone.Focus()
					return m, filterHostedZone(m.textinput.Value())
				}
				if m.tab == ROUTE53_RECORD_TAB { // Filter ROUTE53_RECORD
					m.loading.msg = FILTERING_RECORDS_MSG
					m.textinput.Blur()
					m.table.record.Focus()
					return m, filterRecords(&m.table.hostedZone.SelectedRow()[2], m.textinput.Value())
				}
			}

		}

	case fetchHostedZoneMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.table.hostedZone.SetRows(msg.Rows)
		m.table.hostedZone.Focus()
		return m, nil

	case fetchRecordsMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.table.record.SetRows(msg.Rows)
		m.table.record.SetCursor(0)
		m.table.record.Focus()
		return m, nil

	case filterHostedZoneMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.textinput.Reset()
		m.table.hostedZone.SetRows(msg.Rows)
		m.table.hostedZone.Focus()
		return m, nil

	case filterRecordsMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.textinput.Reset()
		m.table.record.SetRows(msg.Rows)
		m.table.record.Focus()
		return m, nil

	}

	if m.loading.msg != "" { // if loading
		var cmd tea.Cmd
		m.loading.spinner, cmd = m.loading.spinner.Update(msg)
		// return m, tea.Batch(cmd, m.loading.spinner.Tick)
		return m, cmd
	}

	if m.textinput.Focused() {
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		return m, tea.Batch(cmd, m.textinput.Cursor.BlinkCmd())
	}

	if m.table.hostedZone.Focused() || m.table.record.Focused() {
		var (
			hostedZoneCmd tea.Cmd
			recordCmd     tea.Cmd
		)
		m.table.hostedZone, hostedZoneCmd = m.table.hostedZone.Update(msg)
		m.table.record, recordCmd = m.table.record.Update(msg)
		return m, tea.Batch(hostedZoneCmd, recordCmd)
	}

	return m, nil
}
