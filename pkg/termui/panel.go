package termui

import (
	"github.com/gdamore/tcell/v2"
)

// Panel represents a container with a title and a border that can be collapsed
type Panel struct {
	Container
	title         string
	children      []Widget
	titleStyle    tcell.Style
	borderStyle   tcell.Style
	contentStyle  tcell.Style
	isCollapsed   bool
	collapsible   bool
	titleAlign    Alignment // 0: left, 1: center, 2: right
	childrenFocus int       // Index of focused child, -1 if none
}

// NewPanel creates a new panel
func NewPanel(x, y, width, height int) *Panel {
	return &Panel{
		Container: Container{
			x:       x,
			y:       y,
			width:   width,
			height:  height,
			visible: true,
		},
		title:         "",
		children:      []Widget{},
		titleStyle:    tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true),
		borderStyle:   tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		contentStyle:  tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		isCollapsed:   false,
		collapsible:   true,
		titleAlign:    AlignLeft,
		childrenFocus: -1,
	}
}

// SetTitle sets the panel title
func (p *Panel) SetTitle(title string) *Panel {
	p.title = title
	return p
}

// SetStyles sets the styles for the panel
func (p *Panel) SetStyles(titleStyle, borderStyle, contentStyle tcell.Style) *Panel {
	p.titleStyle = titleStyle
	p.borderStyle = borderStyle
	p.contentStyle = contentStyle
	return p
}

// SetCollapsible sets whether the panel can be collapsed
func (p *Panel) SetCollapsible(collapsible bool) *Panel {
	p.collapsible = collapsible
	return p
}

// SetCollapsed sets whether the panel is collapsed
func (p *Panel) SetCollapsed(collapsed bool) *Panel {
	p.isCollapsed = collapsed
	return p
}

// ToggleCollapsed toggles the collapsed state of the panel
func (p *Panel) ToggleCollapsed() *Panel {
	if p.collapsible {
		p.isCollapsed = !p.isCollapsed
	}
	return p
}

// SetTitleAlign sets the alignment of the panel title
func (p *Panel) SetTitleAlign(align Alignment) *Panel {
	if align >= AlignLeft && align <= AlignRight {
		p.titleAlign = align
	}
	return p
}

// AddChild adds a child component to the panel
func (p *Panel) AddChild(child Widget) *Panel {
	p.children = append(p.children, child)
	return p
}

// Draw draws the panel and its children
func (p *Panel) Draw(screen tcell.Screen) {
	if !p.visible {
		return
	}

	// If collapsed, only draw the title bar
	maxY := p.y + p.height
	if p.isCollapsed {
		maxY = p.y + 1
	}

	// Draw border
	for y := p.y; y < maxY; y++ {
		for x := p.x; x < p.x+p.width; x++ {
			// Border characters
			r := ' '
			style := p.contentStyle

			// Draw the borders
			if y == p.y || y == maxY-1 || x == p.x || x == p.x+p.width-1 {
				style = p.borderStyle
				// Top border
				if y == p.y {
					if x == p.x {
						r = '┌' // Top-left corner
					} else if x == p.x+p.width-1 {
						r = '┐' // Top-right corner
					} else {
						r = '─' // Top border
					}
				} else if y == maxY-1 {
					// Bottom border
					if x == p.x {
						r = '└' // Bottom-left corner
					} else if x == p.x+p.width-1 {
						r = '┘' // Bottom-right corner
					} else {
						r = '─' // Bottom border
					}
				} else if x == p.x {
					// Left border
					r = '│' // Left border
				} else if x == p.x+p.width-1 {
					// Right border
					r = '│' // Right border
				}
			}

			screen.SetContent(x, y, r, nil, style)
		}
	}

	// Draw title if present
	if p.title != "" {
		titleWidth := len(p.title) + 2
		titleX := p.x + 1 // Default left alignment

		// Calculate title position based on alignment
		if p.titleAlign == AlignCenter {
			titleX = p.x + (p.width-titleWidth)/2
		} else if p.titleAlign == AlignRight {
			titleX = p.x + p.width - titleWidth - 1
		}

		// Draw title background
		for i := 0; i < titleWidth; i++ {
			screen.SetContent(titleX+i, p.y, ' ', nil, p.titleStyle)
		}

		// Draw title text
		for i, r := range p.title {
			screen.SetContent(titleX+1+i, p.y, r, nil, p.titleStyle)
		}
	}

	// Draw children if not collapsed
	if !p.isCollapsed {
		for _, child := range p.children {
			child.Draw(screen)
		}
	}
}

// HandleEvent handles events for the panel and its children
func (p *Panel) HandleEvent(event tcell.Event) bool {
	if !p.visible {
		return false
	}

	// Check for panel collapse toggle
	if p.collapsible {
		switch ev := event.(type) {
		case *tcell.EventMouse:
			x, y := ev.Position()
			if y == p.y && x >= p.x && x < p.x+p.width {
				if ev.Buttons() == tcell.ButtonPrimary {
					p.ToggleCollapsed()
					return true
				}
			}
		}
	}

	// If collapsed, don't pass events to children
	if p.isCollapsed {
		return false
	}

	// If a child is focused, pass event to it first
	if p.childrenFocus >= 0 && p.childrenFocus < len(p.children) {
		if p.children[p.childrenFocus].HandleEvent(event) {
			return true
		}
	}

	// Pass event to all other children
	for i, child := range p.children {
		if i != p.childrenFocus {
			if child.HandleEvent(event) {
				return true
			}
		}
	}

	return false
}

// GetBounds returns the panel's bounds
func (p *Panel) GetBounds() (x, y, width, height int) {
	return p.x, p.y, p.width, p.height
}

// SetBounds sets the panel's bounds
func (p *Panel) SetBounds(x, y, width, height int) {
	p.x = x
	p.y = y
	p.width = width
	p.height = height
}

// GetX returns the x coordinate of the panel
func (p *Panel) GetX() int {
	return p.x
}

// GetY returns the y coordinate of the panel
func (p *Panel) GetY() int {
	return p.y
}

// GetSize returns the width and height of the panel
func (p *Panel) GetSize() (int, int) {
	return p.width, p.height
}

// SetPosition sets the position of the panel
func (p *Panel) SetPosition(x, y int) {
	p.x = x
	p.y = y
}

// Show shows the panel
func (p *Panel) Show() {
	p.visible = true
}

// Hide hides the panel
func (p *Panel) Hide() {
	p.visible = false
}

// SetVisible sets the visibility of the panel
func (p *Panel) SetVisible(visible bool) {
	p.visible = visible
}
