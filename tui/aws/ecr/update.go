package ecr

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// TODO: dynamic column width size

		// m.table.ecrRepo.SetWidth(msg.Width - 2)
		// m.table.ecrRepoImage.SetWidth(msg.Width - 2)
		m.table.ecrRepo.SetHeight(msg.Height - 21)
		m.table.ecrRepoImage.SetHeight(msg.Height - 21)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(): // Quit
			return m, tea.Quit

		case "Q", "q", tea.KeyEsc.String():
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
