package termui

import (
	"github.com/gdamore/tcell/v2"
)

// Alignment represents text alignment options
type Alignment int

// Alignment constants
const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Widget is the base interface for all UI components
type Widget interface {
	// Draw draws the component on the screen
	Draw(screen tcell.Screen)

	// HandleEvent handles an event and returns true if it was handled
	HandleEvent(event tcell.Event) bool

	// GetX returns the x coordinate of the component
	GetX() int

	// GetY returns the y coordinate of the component
	GetY() int

	// GetSize returns the width and height of the component
	GetSize() (int, int)

	// SetPosition sets the position of the component
	SetPosition(x, y int)

	// Show shows the component
	Show()

	// Hide hides the component
	Hide()

	// IsVisible returns whether the component is visible
	IsVisible() bool

	// GetBounds returns the bounds of the component
	GetBounds() (x, y, width, height int)

	// SetBounds sets the bounds of the component
	SetBounds(x, y, width, height int)

	// SetVisible sets the visibility of the component
	SetVisible(visible bool)
}
