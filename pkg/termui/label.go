package termui

import (
	"github.com/gdamore/tcell/v2"
)

// Label represents a text label component
type Label struct {
	Container
	text  string
	style tcell.Style
	align Alignment
	wrap  bool
}

// NewLabel creates a new label
func NewLabel(x, y int, text string) *Label {
	return &Label{
		Container: Container{
			x:       x,
			y:       y,
			width:   len(text),
			height:  1,
			visible: true,
		},
		text:  text,
		style: tcell.StyleDefault.Foreground(tcell.ColorWhite),
		align: AlignLeft,
		wrap:  false,
	}
}

// SetText sets the label text
func (l *Label) SetText(text string) *Label {
	l.text = text
	l.width = len(text)
	return l
}

// GetText returns the label text
func (l *Label) GetText() string {
	return l.text
}

// SetStyle sets the label style
func (l *Label) SetStyle(style tcell.Style) *Label {
	l.style = style
	return l
}

// Draw draws the label
func (l *Label) Draw(screen tcell.Screen) {
	if !l.visible {
		return
	}

	// Draw text
	for i, r := range l.text {
		if i >= l.width {
			break
		}
		screen.SetContent(l.x+i, l.y, r, nil, l.style)
	}
}

// HandleEvent handles events for the label
func (l *Label) HandleEvent(event tcell.Event) bool {
	// Labels don't typically handle events
	return false
}

// WithAlign sets the alignment of the label
func (l *Label) WithAlign(align Alignment) *Label {
	l.align = align
	return l
}

// WithWrap sets whether the label should wrap text
func (l *Label) WithWrap(wrap bool) *Label {
	l.wrap = wrap
	return l
}

// SetVisible sets whether the label is visible
func (l *Label) SetVisible(visible bool) {
	l.visible = visible
}

// IsVisible returns whether the label is visible
func (l *Label) IsVisible() bool {
	return l.visible
}

// GetBounds returns the bounds of the label
func (l *Label) GetBounds() (x, y, width, height int) {
	return l.x, l.y, l.width, l.height
}

// SetBounds sets the bounds of the label
func (l *Label) SetBounds(x, y, width, height int) {
	l.x, l.y, l.width, l.height = x, y, width, height
}

// Focus focuses the label
func (l *Label) Focus() {
	// Implementation needed
}

// Blur blurs the label
func (l *Label) Blur() {
	// Implementation needed
}

// IsFocused returns whether the label is focused
func (l *Label) IsFocused() bool {
	// Implementation needed
	return false
}

// SetOnFocus sets the focus callback
func (l *Label) SetOnFocus(callback func()) {
	// Implementation needed
}

// SetOnBlur sets the blur callback
func (l *Label) SetOnBlur(callback func()) {
	// Implementation needed
}

// GetX returns the x coordinate of the label
func (l *Label) GetX() int {
	return l.x
}

// GetY returns the y coordinate of the label
func (l *Label) GetY() int {
	return l.y
}

// GetSize returns the width and height of the label
func (l *Label) GetSize() (int, int) {
	return l.width, l.height
}

// SetPosition sets the position of the label
func (l *Label) SetPosition(x, y int) {
	l.x = x
	l.y = y
}

// Show shows the label
func (l *Label) Show() {
	l.visible = true
}

// Hide hides the label
func (l *Label) Hide() {
	l.visible = false
}
