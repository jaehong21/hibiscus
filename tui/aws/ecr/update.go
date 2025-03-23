package ecr

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// Message to clear the success message after a delay
type clearMsgMsg struct{}

// Command to clear the message after a delay
func clearMessageAfterDelay() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearMsgMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Enable dynamic column width sizing
		tableWidth := msg.Width - 2

		// Set overall table widths
		m.table.ecrRepo.SetWidth(tableWidth)
		m.table.ecrRepoImage.SetWidth(tableWidth)

		// Set table heights
		m.table.ecrRepo.SetHeight(msg.Height - 12)
		m.table.ecrRepoImage.SetHeight(msg.Height - 12)

		// Update repository table column widths with new proportional columns
		nameWidth := int(float64(tableWidth) * 0.20)    // 20% of width
		uriWidth := int(float64(tableWidth) * 0.50)     // 50% of width
		createdWidth := int(float64(tableWidth) * 0.25) // 25% of width

		ecrRepoColumns := []table.Column{
			{Title: "Repository name", Width: nameWidth},
			{Title: "URI", Width: uriWidth},
			{Title: "Created at", Width: createdWidth},
		}
		m.table.ecrRepo.SetColumns(ecrRepoColumns)

		// Update image table column widths with new proportional columns
		tagWidth := int(float64(tableWidth) * 0.14)    // 14% of width
		pushedWidth := int(float64(tableWidth) * 0.22) // 22% of width
		sizeWidth := int(float64(tableWidth) * 0.10)   // 10% of width
		digestWidth := int(float64(tableWidth) * 0.49) // 49% of width

		ecrRepoImageColumns := []table.Column{
			{Title: "Tag", Width: tagWidth},
			{Title: "Pushed at", Width: pushedWidth},
			{Title: "Size", Width: sizeWidth},
			{Title: "Digest", Width: digestWidth},
		}
		m.table.ecrRepoImage.SetColumns(ecrRepoImageColumns)

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
				if m.tab == ECR_REPO_TAB {
					m.table.ecrRepo.Focus()
				} else if m.tab == ECR_IMAGE_TAB {
					m.table.ecrRepoImage.Focus()
				}
				return m, nil
			}

			if m.tab == ECR_IMAGE_TAB { // TAB: ECR_IMAGE_TAB -> ECR_REPO_TAB
				m.tab = ECR_REPO_TAB
				m.table.ecrRepo.Focus()
				m.table.ecrRepoImage.Blur()
				return m, nil
			}

		case "R", "r":
			if m.table.ecrRepo.Focused() { // Refresh ECR_REPO_TAB
				m.loading.msg = FETCHING_REPOS_MSG
				return m, tea.Batch(fetchRepos(), m.loading.spinner.Tick)
			}
			if m.table.ecrRepoImage.Focused() { // Refresh ECR_IMAGE_TAB
				m.loading.msg = FETCHING_IMAGES_MSG
				return m, tea.Batch(fetchImages(&m.table.ecrRepo.SelectedRow()[0]), m.loading.spinner.Tick)
			}

		case "/": // Search
			if !m.textinput.Focused() {
				m.textinput.Focus()
				m.table.ecrRepo.Blur()
				m.table.ecrRepoImage.Blur()
				return m, nil
			}

		case tea.KeyEnter.String():
			if m.table.ecrRepo.Focused() { // TAB: ECR_REPO_TAB -> ECR_IMAGE_TAB
				m.tab = ECR_IMAGE_TAB
				m.loading.msg = FETCHING_IMAGES_MSG

				m.table.ecrRepo.Blur()
				m.table.ecrRepoImage.Focus()
				return m, tea.Batch(
					fetchImages(&m.table.ecrRepo.SelectedRow()[0]),
					m.loading.spinner.Tick,
				)
			}

			if m.textinput.Focused() {
				if m.tab == ECR_REPO_TAB { // Filter ECR_REPO
					m.loading.msg = FILTERING_REPOS_MSG
					m.textinput.Blur()
					m.table.ecrRepo.Focus()
					return m, filterRepos(m.textinput.Value())
				}
				if m.tab == ECR_IMAGE_TAB { // Filter ECR_IMAGE
					m.loading.msg = FILTERING_IMAGES_MSG
					m.textinput.Blur()
					m.table.ecrRepoImage.Focus()
					return m, filterImages(&m.table.ecrRepo.SelectedRow()[0], m.textinput.Value())
				}
			}

		case "C", "c", "Y", "y":
			if m.table.ecrRepo.Focused() { // Copy repository URI when in ECR_REPO_TAB
				repoURI := m.table.ecrRepo.SelectedRow()[1]
				err := clipboard.WriteAll(repoURI)
				if err != nil {
					m.err = err
				} else {
					m.msg = "Repository URI copied to clipboard"
				}
				return m, clearMessageAfterDelay()
			}
			if m.table.ecrRepoImage.Focused() { // Copy full image URI with tag when in ECR_IMAGE_TAB
				repoURI := m.table.ecrRepo.SelectedRow()[1]
				imageTag := m.table.ecrRepoImage.SelectedRow()[0]
				fullImageURI := repoURI + ":" + imageTag
				err := clipboard.WriteAll(fullImageURI)
				if err != nil {
					m.err = err
				} else {
					m.msg = "Full image URI copied to clipboard"
				}
				return m, clearMessageAfterDelay()
			}

		}

	case fetchReposMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.table.ecrRepo.SetRows(msg.Rows)
		m.table.ecrRepo.Focus()
		return m, nil

	case fetchImagesMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.table.ecrRepoImage.SetRows(msg.Rows)
		m.table.ecrRepoImage.Focus()
		return m, nil

	case filterReposMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.textinput.Reset()
		m.table.ecrRepo.SetRows(msg.Rows)
		m.table.ecrRepo.Focus()
		return m, nil

	case filterImagesMsg:
		m.loading.msg = ""
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.textinput.Reset()
		m.table.ecrRepoImage.SetRows(msg.Rows)
		m.table.ecrRepoImage.Focus()
		return m, nil

	case clearMsgMsg:
		m.msg = ""
		// Force a layout update by simulating a window resize with the current dimensions
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			}
		}
	}

	if m.loading.msg != "" { // if loading
		var cmd tea.Cmd
		m.loading.spinner, cmd = m.loading.spinner.Update(msg)
		return m, cmd
	}

	if m.textinput.Focused() {
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		return m, tea.Batch(cmd, m.textinput.Cursor.BlinkCmd())
	}

	if m.table.ecrRepo.Focused() || m.table.ecrRepoImage.Focused() {
		var (
			ecrRepoCmd      tea.Cmd
			ecrRepoImageCmd tea.Cmd
		)
		m.table.ecrRepo, ecrRepoCmd = m.table.ecrRepo.Update(msg)
		m.table.ecrRepoImage, ecrRepoImageCmd = m.table.ecrRepoImage.Update(msg)
		return m, tea.Batch(ecrRepoCmd, ecrRepoImageCmd)
	}

	return m, nil
}
