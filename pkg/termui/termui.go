// Package termui provides a rich terminal UI library for building interactive
// command-line applications with advanced visual elements.
package termui

import (
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Application represents a terminal UI application
type Application struct {
	screen        tcell.Screen
	focus         Widget
	widgets       []Widget
	running       bool
	screenLock    sync.Mutex
	quitting      chan struct{}
	mu            sync.Mutex
	eventHandlers []EventHandler
	styles        *StyleManager
	focusManager  *FocusManager
}

// EventHandler is a function that processes events
type EventHandler func(event tcell.Event) bool

// Drawable represents an object that can be drawn on the screen
type Drawable interface {
	Draw(screen tcell.Screen)
	GetBounds() (x, y, width, height int)
	SetBounds(x, y, width, height int)
	HandleEvent(event tcell.Event) bool
}

// Focusable represents an object that can receive focus
type Focusable interface {
	Drawable
	Focus()
	Blur()
	IsFocused() bool
}

// FocusManager manages focus between UI components
type FocusManager struct {
	focusables []Focusable
	current    int
	mu         sync.Mutex
}

// NewFocusManager creates a new focus manager
func NewFocusManager() *FocusManager {
	return &FocusManager{
		focusables: []Focusable{},
		current:    -1,
	}
}

// AddFocusable adds a focusable component
func (f *FocusManager) AddFocusable(component Focusable) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.focusables = append(f.focusables, component)
	if f.current == -1 && len(f.focusables) > 0 {
		f.current = 0
		f.focusables[0].Focus()
	}
}

// RemoveFocusable removes a focusable component
func (f *FocusManager) RemoveFocusable(component Focusable) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, c := range f.focusables {
		if c == component {
			f.focusables = append(f.focusables[:i], f.focusables[i+1:]...)
			if f.current == i {
				f.current = -1
				if len(f.focusables) > 0 {
					f.current = 0
					f.focusables[0].Focus()
				}
			} else if f.current > i {
				f.current--
			}
			break
		}
	}
}

// Next focuses the next component
func (f *FocusManager) Next() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.focusables) == 0 {
		return
	}
	if f.current >= 0 {
		f.focusables[f.current].Blur()
	}
	f.current = (f.current + 1) % len(f.focusables)
	f.focusables[f.current].Focus()
}

// Previous focuses the previous component
func (f *FocusManager) Previous() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.focusables) == 0 {
		return
	}
	if f.current >= 0 {
		f.focusables[f.current].Blur()
	}
	f.current = (f.current - 1 + len(f.focusables)) % len(f.focusables)
	f.focusables[f.current].Focus()
}

// GetCurrent returns the currently focused component
func (f *FocusManager) GetCurrent() Focusable {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.current >= 0 && f.current < len(f.focusables) {
		return f.focusables[f.current]
	}
	return nil
}

// NewApplication creates a new terminal UI application
func NewApplication() (*Application, error) {
	// Initialize tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	// Set default style
	screen.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite))

	return &Application{
		screen:        screen,
		widgets:       make([]Widget, 0),
		quitting:      make(chan struct{}),
		eventHandlers: make([]EventHandler, 0),
		styles:        NewStyleManager(),
		focusManager:  NewFocusManager(),
	}, nil
}

// AddWidget adds a widget to the application
func (app *Application) AddWidget(widget Widget) {
	app.widgets = append(app.widgets, widget)
}

// SetFocus sets the focused widget
func (app *Application) SetFocus(widget Widget) {
	app.focus = widget
}

// Run starts the application
func (app *Application) Run() {
	app.running = true

	// Handle events in a goroutine
	go func() {
		for app.running {
			ev := app.screen.PollEvent()

			app.screenLock.Lock()

			// Check if we should quit
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					app.Quit()
					app.screenLock.Unlock()
					return
				}
			case *tcell.EventResize:
				app.screen.Sync()
			}

			// Pass the event to the focused widget first
			if app.focus != nil {
				if app.focus.HandleEvent(ev) {
					app.screenLock.Unlock()
					continue
				}
			}

			// If the focused widget didn't handle it, try all widgets
			for _, widget := range app.widgets {
				if widget.HandleEvent(ev) {
					break
				}
			}

			app.screenLock.Unlock()
		}
	}()

	// Main draw loop
	for app.running {
		app.screenLock.Lock()

		// Clear the screen
		app.screen.Clear()

		// Draw all widgets
		for _, widget := range app.widgets {
			widget.Draw(app.screen)
		}

		// Update the screen
		app.screen.Show()

		app.screenLock.Unlock()

		// Sleep a bit to avoid consuming too much CPU
		select {
		case <-app.quitting:
			return
		default:
			// Sleep for ~60fps
			time.Sleep(16 * time.Millisecond)
		}
	}
}

// Quit stops the application
func (app *Application) Quit() {
	app.running = false
	close(app.quitting)
	app.screen.Fini()
	os.Exit(0)
}

// GetSize returns the screen size
func (app *Application) GetSize() (int, int) {
	return app.screen.Size()
}

// RegisterEventHandler registers a global event handler
func (a *Application) RegisterEventHandler(handler EventHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventHandlers = append(a.eventHandlers, handler)
}

// Styles returns the style manager
func (a *Application) Styles() *StyleManager {
	return a.styles
}

// FocusManager returns the focus manager
func (a *Application) FocusManager() *FocusManager {
	return a.focusManager
}

// GetScreen returns the underlying screen
func (a *Application) GetScreen() tcell.Screen {
	return a.screen
}
