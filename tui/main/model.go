package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/config"
	"github.com/jaehong21/hibiscus/tui/aws/ecr"
	"github.com/jaehong21/hibiscus/tui/aws/elb"
	"github.com/jaehong21/hibiscus/tui/aws/route53"
)

// Available AWS services for navigation
var availableServices = []string{
	"ecr",
	"route53",
	"elb",
}

type model struct {
	ecr            tea.Model
	route53        tea.Model
	elb            tea.Model
	navInput       textinput.Model
	showNav        bool
	navError       string
	suggestions    []string
	selectedIndex  int
	lastWindowSize tea.WindowSizeMsg // Store the last window size
}

func New(config *config.Config) model {
	navInput := textinput.New()
	navInput.Prompt = ": "
	navInput.Placeholder = "type service name (ecr, route53, ...)"
	navInput.CharLimit = 30

	return model{
		ecr:            ecr.New(),
		route53:        route53.New(),
		elb:            elb.New(),
		navInput:       navInput,
		showNav:        false,
		navError:       "",
		suggestions:    []string{},
		selectedIndex:  0,
		lastWindowSize: tea.WindowSizeMsg{}, // Initialize with zero values
	}
}

// Helper function to create a command that sends the last window size
func (m model) resendWindowSize() tea.Cmd {
	return func() tea.Msg {
		return m.lastWindowSize
	}
}

// https://github.com/charmbracelet/bubbletea/blob/491eda41276c3419d519bc8c622725fa587b7e37/tea.go#L513
// NOTE: needed to initialize the ecr model
// p.Run() call the Init() method of the model
// but for our custom models, we need to call the Init() method manually

func (m model) Init() tea.Cmd {
	// Entrypoint to tab in Application

	// No need to set tab key here, as it's already loaded from config file
	// If there was no saved config, it defaults to HOME_TAB in the Initialize function

	return tea.Batch(
		m.ecr.Init(),
		m.route53.Init(),
		m.elb.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Store window size when received
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.lastWindowSize = windowMsg
	}

	// Handle navigation input
	if m.showNav {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				// Exit navigation mode
				m.showNav = false
				m.navInput.Reset()
				m.navError = ""
				m.suggestions = []string{}
				m.selectedIndex = 0
				return m, nil

			case "enter":
				// Try to navigate to the service
				service := m.navInput.Value()
				// If a suggestion is selected, use that instead
				if m.selectedIndex < len(m.suggestions) && m.selectedIndex >= 0 {
					service = m.suggestions[m.selectedIndex]
				}

				// Reset navigation state
				m.showNav = false
				m.navInput.Reset()
				m.navError = ""
				m.suggestions = []string{}
				m.selectedIndex = 0

				switch service {
				case "ecr":
					config.SetTabKey(config.ECR_TAB)
					// Reinitialize the ECR model and send window size
					return m, tea.Batch(
						m.ecr.Init(),
						m.resendWindowSize(), // Send the stored window size
					)
				case "route53":
					config.SetTabKey(config.ROUTE53_TAB)
					// Reinitialize the Route53 model and send window size
					return m, tea.Batch(
						m.route53.Init(),
						m.resendWindowSize(), // Send the stored window size
					)
				case "elb":
					config.SetTabKey(config.ELB_TAB)
					// Reinitialize the ELB model and send window size
					return m, tea.Batch(
						m.elb.Init(),
						m.resendWindowSize(), // Send the stored window size
					)
				default:
					m.navError = "Service not supported: " + service
					return m, nil
				}

			case "tab", "down":
				// Cycle through suggestions
				if len(m.suggestions) > 0 {
					m.selectedIndex = (m.selectedIndex + 1) % len(m.suggestions)
				}
				return m, nil

			case "shift+tab", "up":
				// Cycle through suggestions backwards
				if len(m.suggestions) > 0 {
					m.selectedIndex = (m.selectedIndex - 1 + len(m.suggestions)) % len(m.suggestions)
				}
				return m, nil
			}

			// Filter suggestions
			var cmd tea.Cmd
			m.navInput, cmd = m.navInput.Update(msg)
			cmds = append(cmds, cmd)

			// Update suggestions based on current input
			input := m.navInput.Value()
			m.suggestions = []string{}
			if input != "" {
				for _, service := range availableServices {
					if strings.Contains(service, strings.ToLower(input)) {
						m.suggestions = append(m.suggestions, service)
					}
				}
			} else {
				// Show all available services when input is empty
				m.suggestions = make([]string, len(availableServices))
				copy(m.suggestions, availableServices)
			}
			m.selectedIndex = 0
			return m, tea.Batch(cmds...)
		}
	} else {
		// Normal mode
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case ":":
				// Enter navigation mode
				m.showNav = true
				m.navInput.Focus()
				m.navError = ""
				// Populate suggestions with all available services immediately
				m.suggestions = make([]string, len(availableServices))
				copy(m.suggestions, availableServices)
				m.selectedIndex = 0
				return m, nil
			}
		}
	}

	// Regular tab handling
	tab := config.GetConfig().TabKey

	switch tab {
	case config.ECR_TAB:
		var ecrCmd tea.Cmd
		m.ecr, ecrCmd = m.ecr.Update(msg)
		cmds = append(cmds, ecrCmd)

	case config.ROUTE53_TAB:
		var route53Cmd tea.Cmd
		m.route53, route53Cmd = m.route53.Update(msg)
		cmds = append(cmds, route53Cmd)

	case config.ELB_TAB:
		var elbCmd tea.Cmd
		m.elb, elbCmd = m.elb.Update(msg)
		cmds = append(cmds, elbCmd)

	default:

	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s string

	// Show navigation if active
	if m.showNav {
		s += "\n " + m.navInput.View() + "\n"

		// Show suggestions
		if len(m.suggestions) > 0 {
			s += "\n Suggestions:\n"
			for i, suggestion := range m.suggestions {
				if i == m.selectedIndex {
					s += " > " + suggestion + "\n"
				} else {
					s += "   " + suggestion + "\n"
				}
			}
		}

		// Show error if any
		if m.navError != "" {
			s += "\n Error: " + m.navError + "\n"
		}

		return s
	}

	// Regular view
	tab := config.GetConfig().TabKey

	switch tab {
	case config.ECR_TAB:
		s += m.ecr.View()

	case config.ROUTE53_TAB:
		s += m.route53.View()

	case config.ELB_TAB:
		s += m.elb.View()

	default:

	}

	return s
}
