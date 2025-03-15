package termui

import (
	"github.com/gdamore/tcell/v2"
)

// Modal represents a modal dialog box
type Modal struct {
	Container
	title         string
	content       string
	buttons       []*Button
	titleStyle    tcell.Style
	contentStyle  tcell.Style
	borderStyle   tcell.Style
	backdropStyle tcell.Style
	showBackdrop  bool
	onClose       func()
	focusedButton int
}

// NewModal creates a new modal dialog
func NewModal(x, y, width, height int) *Modal {
	return &Modal{
		Container: Container{
			x:       x,
			y:       y,
			width:   width,
			height:  height,
			visible: false,
		},
		title:         "",
		content:       "",
		buttons:       []*Button{},
		titleStyle:    tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true),
		contentStyle:  tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		borderStyle:   tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		backdropStyle: tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlack),
		showBackdrop:  true,
		focusedButton: 0,
	}
}

// SetTitle sets the modal title
func (m *Modal) SetTitle(title string) *Modal {
	m.title = title
	return m
}

// SetContent sets the modal content text
func (m *Modal) SetContent(content string) *Modal {
	m.content = content
	return m
}

// AddButton adds a button to the modal
func (m *Modal) AddButton(button *Button) *Modal {
	m.buttons = append(m.buttons, button)
	return m
}

// SetStyles sets the styles for the modal
func (m *Modal) SetStyles(titleStyle, contentStyle, borderStyle, backdropStyle tcell.Style) *Modal {
	m.titleStyle = titleStyle
	m.contentStyle = contentStyle
	m.borderStyle = borderStyle
	m.backdropStyle = backdropStyle
	return m
}

// SetShowBackdrop sets whether to show a backdrop behind the modal
func (m *Modal) SetShowBackdrop(show bool) *Modal {
	m.showBackdrop = show
	return m
}

// SetOnClose sets the callback function when the modal is closed
func (m *Modal) SetOnClose(onClose func()) *Modal {
	m.onClose = onClose
	return m
}

// Show shows the modal
func (m *Modal) Show() {
	m.visible = true
}

// Hide hides the modal
func (m *Modal) Hide() {
	m.visible = false
	if m.onClose != nil {
		m.onClose()
	}
}

// Draw draws the modal
func (m *Modal) Draw(screen tcell.Screen) {
	if !m.visible {
		return
	}

	screenWidth, screenHeight := screen.Size()

	// Draw backdrop if enabled
	if m.showBackdrop {
		for y := 0; y < screenHeight; y++ {
			for x := 0; x < screenWidth; x++ {
				screen.SetContent(x, y, ' ', nil, m.backdropStyle)
			}
		}
	}

	// Draw modal border
	for y := m.y; y < m.y+m.height; y++ {
		for x := m.x; x < m.x+m.width; x++ {
			// Draw corners
			if x == m.x && y == m.y {
				screen.SetContent(x, y, '┌', nil, m.borderStyle) // Top-left corner
			} else if x == m.x+m.width-1 && y == m.y {
				screen.SetContent(x, y, '┐', nil, m.borderStyle) // Top-right corner
			} else if x == m.x && y == m.y+m.height-1 {
				screen.SetContent(x, y, '└', nil, m.borderStyle) // Bottom-left corner
			} else if x == m.x+m.width-1 && y == m.y+m.height-1 {
				screen.SetContent(x, y, '┘', nil, m.borderStyle) // Bottom-right corner
			} else if y == m.y || y == m.y+m.height-1 {
				screen.SetContent(x, y, '─', nil, m.borderStyle) // Horizontal borders
			} else if x == m.x || x == m.x+m.width-1 {
				screen.SetContent(x, y, '│', nil, m.borderStyle) // Vertical borders
			} else {
				screen.SetContent(x, y, ' ', nil, m.contentStyle) // Interior
			}
		}
	}

	// Draw title if set
	if m.title != "" {
		// Draw title separator
		for x := m.x + 1; x < m.x+m.width-1; x++ {
			screen.SetContent(x, m.y+2, '─', nil, m.borderStyle)
		}

		// Draw title text
		for i, r := range " " + m.title + " " {
			if m.x+1+i >= m.x+m.width-1 {
				break
			}
			screen.SetContent(m.x+1+i, m.y+1, r, nil, m.titleStyle)
		}
	}

	// Calculate content area
	contentStartY := m.y + 2
	if m.title != "" {
		contentStartY = m.y + 3
	}

	// Draw content text
	contentX := m.x + 2
	contentY := contentStartY + 1
	contentWidth := m.width - 4

	// Split content by newlines and word wrap
	lines := []string{}
	currentLine := ""

	for _, r := range m.content {
		if r == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(r)
			if len(currentLine) >= contentWidth {
				lines = append(lines, currentLine)
				currentLine = ""
			}
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Draw content lines
	for i, line := range lines {
		if contentY+i >= m.y+m.height-2 { // -2 for space for buttons
			break
		}
		for j, r := range line {
			if contentX+j >= m.x+m.width-2 {
				break
			}
			screen.SetContent(contentX+j, contentY+i, r, nil, m.contentStyle)
		}
	}

	// Draw buttons
	if len(m.buttons) > 0 {
		buttonY := m.y + m.height - 3
		buttonSpacing := 2
		totalButtonsWidth := 0
		for _, button := range m.buttons {
			totalButtonsWidth += button.width
		}
		totalButtonsWidth += (len(m.buttons) - 1) * buttonSpacing

		buttonX := m.x + (m.width-totalButtonsWidth)/2

		for _, button := range m.buttons {
			button.x = buttonX
			button.y = buttonY
			button.Draw(screen)
			buttonX += button.width + buttonSpacing
		}
	}
}

// HandleEvent handles events for the modal
func (m *Modal) HandleEvent(event tcell.Event) bool {
	if !m.visible {
		return false
	}

	// Handle button navigation
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			if len(m.buttons) > 0 {
				m.focusedButton = (m.focusedButton + 1) % len(m.buttons)
				return true
			}
		case tcell.KeyBacktab, tcell.KeyLeft:
			if len(m.buttons) > 0 {
				m.focusedButton = (m.focusedButton - 1)
				if m.focusedButton < 0 {
					m.focusedButton = len(m.buttons) - 1
				}
				return true
			}
		case tcell.KeyEnter:
			if len(m.buttons) > 0 && m.focusedButton >= 0 && m.focusedButton < len(m.buttons) {
				// Trigger button action
				button := m.buttons[m.focusedButton]
				if button.onClick != nil {
					button.onClick()
				}
				return true
			}
		case tcell.KeyEscape:
			m.Hide()
			return true
		}
	}

	// Pass event to focused button
	if len(m.buttons) > 0 && m.focusedButton >= 0 && m.focusedButton < len(m.buttons) {
		return m.buttons[m.focusedButton].HandleEvent(event)
	}

	return false
}
