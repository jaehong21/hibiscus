package hibiscus

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jaehong21/hibiscus/config"
)

// ServiceFactory allows the shell to lazily construct services once the
// tview application and shared context are available.
type ServiceFactory func(ctx ServiceContext) Service

// App is the top-level Hibiscus controller built on tview. It handles the
// global layout, keybindings, and service switching while each service owns
// its local UI and behaviors.
type App struct {
	cfg *config.Config
	app *tview.Application

	header    *tview.TextView
	statusBar *tview.TextView
	errorBar  *tview.TextView

	pages   *tview.Pages
	content *tview.Pages

	services map[string]Service
	order    []string
	current  Service

	palette *commandPalette
}

// New wires the provided services into a single application instance.
func New(cfg *config.Config, factories []ServiceFactory) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	if len(factories) == 0 {
		return nil, fmt.Errorf("at least one service must be registered")
	}

	tviewApp := tview.NewApplication()
	header := tview.NewTextView().SetDynamicColors(true)
	status := tview.NewTextView().SetDynamicColors(true)
	status.SetText("[lightgreen]Ready[-]")
	errorBar := tview.NewTextView().SetDynamicColors(true)

	content := tview.NewPages()
	pages := tview.NewPages()

	hib := &App{
		cfg:       cfg,
		app:       tviewApp,
		header:    header,
		statusBar: status,
		errorBar:  errorBar,
		content:   content,
		pages:     pages,
		services:  make(map[string]Service),
	}

	ctx := ServiceContext{
		App:       tviewApp,
		SetStatus: hib.setStatus,
		SetError:  hib.setError,
	}

	// Instantiate services in the provided order so the palette matches the
	// user's mental model from the Bubble Tea version.
	for _, factory := range factories {
		svc := factory(ctx)
		if svc == nil {
			continue
		}
		name := strings.ToLower(svc.Name())
		hib.services[name] = svc
		hib.order = append(hib.order, name)
		content.AddPage(name, svc.Primitive(), true, false)
	}

	if len(hib.services) == 0 {
		return nil, fmt.Errorf("no services were constructed")
	}

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 2, 0, false).
		AddItem(content, 0, 1, true).
		AddItem(status, 1, 0, false).
		AddItem(errorBar, 1, 0, false)

	pages.AddPage("main", layout, true, true)

	hib.palette = newCommandPalette(tviewApp, pages, hib.order, hib.switchService, func() {
		if hib.current != nil {
			hib.current.Activate()
		}
	})

	tviewApp.SetInputCapture(hib.handleGlobalInput)
	hib.updateHeader()

	// Kick off service initialization once everything is wired up.
	for _, name := range hib.order {
		if svc, ok := hib.services[name]; ok {
			svc.Init()
		}
	}

	return hib, nil
}

// Run starts the event loop.
func (a *App) Run() error {
	if a.current == nil {
		initial := serviceNameFromTabKey(a.cfg.TabKey)
		if _, ok := a.services[initial]; !ok {
			if len(a.order) > 0 {
				initial = a.order[0]
			}
		}
		a.switchService(initial)
	}

	a.app.SetRoot(a.pages, true)
	return a.app.EnableMouse(true).Run()
}

func (a *App) handleGlobalInput(event *tcell.EventKey) *tcell.EventKey {
	if event == nil {
		return nil
	}

	if a.palette != nil && a.palette.Visible() {
		return event
	}

	filtering := a.current != nil && a.current.InFilterMode()

	if event.Key() == tcell.KeyCtrlC {
		a.app.Stop()
		return nil
	}

	if filtering {
		// Let the focused filter field consume the keystroke; Esc will bubble
		// down to the service to exit filter mode.
		return event
	}

	switch {
	case event.Rune() == ':':
		if a.palette != nil {
			a.palette.Show()
			return nil
		}
	case event.Rune() == '/':
		if a.current != nil && a.current.EnterFilterMode() {
			return nil
		}
	case event.Rune() == 'R' || event.Rune() == 'r':
		if a.current != nil {
			a.current.Refresh()
			return nil
		}
	}

	if a.current != nil {
		return a.current.HandleInput(event)
	}

	return event
}

func (a *App) switchService(name string) {
	name = strings.ToLower(strings.TrimSpace(name))
	svc, ok := a.services[name]
	if !ok {
		return
	}

	if a.current != nil {
		a.current.Deactivate()
	}

	a.current = svc
	a.content.SwitchToPage(name)
	a.updateHeader()
	a.setStatus(fmt.Sprintf("Showing %s", svc.Title()))
	config.SetTabKey(serviceNameToTabKey(name))
	svc.Activate()
}

func (a *App) updateHeader() {
	title := ""
	if a.current != nil {
		title = fmt.Sprintf("%s", a.current.Title())
	} else {
		title = "Select a service with :"
	}

	helper := "[: ]command  [/]filter  [R]refresh  [Esc]back  [Ctrl+C]quit"
	a.header.SetText(fmt.Sprintf("[yellow]Hibiscus[-] – %s  %s", title, helper))
}

func (a *App) setStatus(msg string) {
	if msg == "" {
		msg = "Ready"
	}
	a.statusBar.SetText(fmt.Sprintf("[lightgreen]%s[-]", msg))
}

func (a *App) setError(err error) {
	if err == nil {
		a.errorBar.SetText("")
		return
	}
	a.errorBar.SetText(fmt.Sprintf("[red]Error: %s[-]", err.Error()))
}

func serviceNameFromTabKey(tab int) string {
	switch tab {
	case config.ROUTE53_TAB:
		return "route53"
	case config.ELB_TAB:
		return "elb"
	case config.ECR_TAB:
		fallthrough
	default:
		return "ecr"
	}
}

func serviceNameToTabKey(name string) int {
	switch strings.ToLower(name) {
	case "route53":
		return config.ROUTE53_TAB
	case "elb":
		return config.ELB_TAB
	default:
		return config.ECR_TAB
	}
}
