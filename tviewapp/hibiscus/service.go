package hibiscus

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ServiceContext wires shared runtime helpers into individual services.
type ServiceContext struct {
	App       *tview.Application
	SetStatus func(string)
	SetError  func(error)
}

// Service describes the contract each AWS view must implement so the
// shell can mount and interact with it in a consistent fashion.
type Service interface {
	// Name returns the unique key used for navigation (e.g. "ecr").
	Name() string
	// Title is rendered in the header so users know which view is active.
	Title() string
	// Primitive returns the root component for the service. The instance is
	// mounted once and kept alive for the lifetime of the application.
	Primitive() tview.Primitive
	// Init is called exactly once during boot to kick off initial data loads.
	Init()
	// Activate is invoked whenever the service becomes visible so it can
	// claim focus or refresh context-sensitive UI.
	Activate()
	// Deactivate lets the service release focus or persist state if needed.
	Deactivate()
	// Refresh requests the service to reload its current dataset.
	Refresh()
	// EnterFilterMode should focus the service's filter input (if any) and
	// return true when it took action.
	EnterFilterMode() bool
	// HandleInput gives the service a chance to consume key events that
	// aren't handled globally. Return nil to consume the event.
	HandleInput(event *tcell.EventKey) *tcell.EventKey
}
