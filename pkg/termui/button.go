package termui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Button is an interactive button component
type Button struct {
	x, y, width, height int
	text                string
	style               tcell.Style
	focusedStyle        tcell.Style
	hoverStyle          tcell.Style
	align               Alignment
	visible             bool
	focused             bool
	disabled            bool
	padding             int
	hotkey              rune
	hotkeyStyle         tcell.Style
	// Callbacks
	onClick func()
	onFocus func()
	onBlur  func()
}

// NewButton creates a new button
func NewButton(x, y, width, height int, text string) *Button {
	return &Button{
		x:            x,
		y:            y,
		width:        width,
		height:       height,
		text:         text,
		style:        tcell.StyleDefault.Background(tcell.ColorSilver).Foreground(tcell.ColorBlack),
		focusedStyle: tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true),
		hoverStyle:   tcell.StyleDefault.Background(tcell.ColorTeal).Foreground(tcell.ColorWhite),
		hotkeyStyle:  tcell.StyleDefault.Underline(true).Bold(true),
		align:        AlignCenter,
		visible:      true,
		disabled:     false,
		padding:      1,
		hotkey:       0,
	}
}

// WithStyle sets the style of the button
func (b *Button) WithStyle(style tcell.Style) *Button {
	b.style = style
	return b
}

// WithFocusedStyle sets the focused style of the button
func (b *Button) WithFocusedStyle(style tcell.Style) *Button {
	b.focusedStyle = style
	return b
}

// WithHoverStyle sets the hover style of the button
func (b *Button) WithHoverStyle(style tcell.Style) *Button {
	b.hoverStyle = style
	return b
}

// WithAlign sets the alignment of the button
func (b *Button) WithAlign(align Alignment) *Button {
	b.align = align
	return b
}

// WithPadding sets the padding of the button
func (b *Button) WithPadding(padding int) *Button {
	b.padding = padding
	return b
}

// WithHotkey sets the hotkey of the button
func (b *Button) WithHotkey(key rune) *Button {
	b.hotkey = key
	return b
}

// WithHotkeyStyle sets the hotkey style of the button
func (b *Button) WithHotkeyStyle(style tcell.Style) *Button {
	b.hotkeyStyle = style
	return b
}

// SetText sets the text of the button
func (b *Button) SetText(text string) {
	b.text = text
}

// GetText returns the text of the button
func (b *Button) GetText() string {
	return b.text
}

// SetVisible sets whether the button is visible
func (b *Button) SetVisible(visible bool) {
	b.visible = visible
}

// IsVisible returns whether the button is visible
func (b *Button) IsVisible() bool {
	return b.visible
}

// SetDisabled sets whether the button is disabled
func (b *Button) SetDisabled(disabled bool) {
	b.disabled = disabled
}

// IsDisabled returns whether the button is disabled
func (b *Button) IsDisabled() bool {
	return b.disabled
}

// Click triggers the button's click action
func (b *Button) Click() {
	if !b.disabled && b.onClick != nil {
		b.onClick()
	}
}

// SetOnClick sets the click callback
func (b *Button) SetOnClick(callback func()) *Button {
	b.onClick = callback
	return b
}

// SetOnFocus sets the focus callback
func (b *Button) SetOnFocus(callback func()) {
	b.onFocus = callback
}

// SetOnBlur sets the blur callback
func (b *Button) SetOnBlur(callback func()) {
	b.onBlur = callback
}

// Draw draws the button
func (b *Button) Draw(screen tcell.Screen) {
	if !b.visible {
		return
	}

	// Select the appropriate style
	style := b.style
	if b.focused {
		style = b.focusedStyle
	}
	if b.disabled {
		style = style.Foreground(tcell.ColorGray)
	}

	// Draw background
	for y := b.y; y < b.y+b.height; y++ {
		for x := b.x; x < b.x+b.width; x++ {
			screen.SetContent(x, y, ' ', nil, style)
		}
	}

	// Center the text vertically
	textY := b.y + (b.height / 2)

	// Calculate text position based on alignment
	textWidth := runewidth.StringWidth(b.text)
	textX := b.x + b.padding // default left alignment

	if b.align == AlignCenter {
		textX = b.x + (b.width-textWidth)/2
	} else if b.align == AlignRight {
		textX = b.x + b.width - textWidth - b.padding
	}

	// Draw text
	for i, r := range b.text {
		if textX+i >= b.x+b.width {
			break
		}

		// Special styling for hotkey if it matches this character
		currentStyle := style
		if b.hotkey != 0 && r == b.hotkey {
			// Simply use the hotkey style directly
			currentStyle = b.hotkeyStyle
		}

		screen.SetContent(textX+i, textY, r, nil, currentStyle)
	}
}

// GetBounds returns the bounds of the button
func (b *Button) GetBounds() (x, y, width, height int) {
	return b.x, b.y, b.width, b.height
}

// SetBounds sets the bounds of the button
func (b *Button) SetBounds(x, y, width, height int) {
	b.x, b.y, b.width, b.height = x, y, width, height
}

// HandleEvent handles events
func (b *Button) HandleEvent(event tcell.Event) bool {
	if !b.visible || b.disabled {
		return false
	}

	switch ev := event.(type) {
	case *tcell.EventMouse:
		x, y := ev.Position()
		// Check if mouse event is within button bounds
		if x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height {
			if ev.Buttons() == tcell.ButtonPrimary {
				b.Click()
				return true
			}
		}
	case *tcell.EventKey:
		// Check if key event matches hotkey
		if b.hotkey != 0 && (ev.Rune() == b.hotkey || ev.Rune() == b.hotkey-32) { // Check both cases
			b.Click()
			return true
		}

		// If focused, handle enter/space to click
		if b.focused && (ev.Key() == tcell.KeyEnter || ev.Key() == ' ') {
			b.Click()
			return true
		}
	}
	return false
}

// Focus focuses the button
func (b *Button) Focus() {
	b.focused = true
	if b.onFocus != nil {
		b.onFocus()
	}
}

// Blur blurs the button
func (b *Button) Blur() {
	b.focused = false
	if b.onBlur != nil {
		b.onBlur()
	}
}

// IsFocused returns whether the button is focused
func (b *Button) IsFocused() bool {
	return b.focused
}

// SetHotkey sets the hotkey of the button
func (b *Button) SetHotkey(key rune) *Button {
	b.hotkey = key
	return b
}

// GetX returns the x coordinate of the button
func (b *Button) GetX() int {
	return b.x
}

// GetY returns the y coordinate of the button
func (b *Button) GetY() int {
	return b.y
}

// GetSize returns the width and height of the button
func (b *Button) GetSize() (int, int) {
	return b.width, b.height
}

// SetPosition sets the position of the button
func (b *Button) SetPosition(x, y int) {
	b.x = x
	b.y = y
}

// Show shows the button
func (b *Button) Show() {
	b.visible = true
}

// Hide hides the button
func (b *Button) Hide() {
	b.visible = false
}
