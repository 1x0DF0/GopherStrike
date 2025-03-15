package termui

import (
	"github.com/gdamore/tcell/v2"
)

// SelectOption represents an option in a select list
type SelectOption struct {
	Text     string
	Value    interface{}
	Selected bool
}

// SelectList represents an interactive list of selectable options
type SelectList struct {
	Container
	options         []SelectOption
	selectedIndex   int
	itemHeight      int
	scrollOffset    int
	maxVisibleItems int
	onChange        func(index int, option SelectOption)
	normalStyle     tcell.Style
	selectedStyle   tcell.Style
}

// NewSelectList creates a new select list component
func NewSelectList(x, y, width, height int) *SelectList {
	s := &SelectList{
		Container: Container{
			x:      x,
			y:      y,
			width:  width,
			height: height,
		},
		itemHeight:      1,
		selectedIndex:   0,
		scrollOffset:    0,
		maxVisibleItems: 0,
		normalStyle:     tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		selectedStyle:   tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
	}
	return s
}

// SetOptions sets the options for the select list
func (s *SelectList) SetOptions(options []SelectOption) *SelectList {
	s.options = options
	s.maxVisibleItems = s.height / s.itemHeight
	if s.selectedIndex >= len(options) {
		s.selectedIndex = len(options) - 1
		if s.selectedIndex < 0 {
			s.selectedIndex = 0
		}
	}
	return s
}

// GetSelectedIndex returns the currently selected index
func (s *SelectList) GetSelectedIndex() int {
	return s.selectedIndex
}

// GetSelectedOption returns the currently selected option
func (s *SelectList) GetSelectedOption() (SelectOption, bool) {
	if len(s.options) == 0 || s.selectedIndex < 0 || s.selectedIndex >= len(s.options) {
		return SelectOption{}, false
	}
	return s.options[s.selectedIndex], true
}

// SetSelectedIndex sets the selected index
func (s *SelectList) SetSelectedIndex(index int) *SelectList {
	if index >= 0 && index < len(s.options) {
		s.selectedIndex = index
		s.ensureSelectedVisible()
	}
	return s
}

// SetOnChange sets the callback function when selection changes
func (s *SelectList) SetOnChange(onChange func(index int, option SelectOption)) *SelectList {
	s.onChange = onChange
	return s
}

// SetStyles sets the normal and selected styles
func (s *SelectList) SetStyles(normalStyle, selectedStyle tcell.Style) *SelectList {
	s.normalStyle = normalStyle
	s.selectedStyle = selectedStyle
	return s
}

// ensureSelectedVisible makes sure the selected item is visible
func (s *SelectList) ensureSelectedVisible() {
	if s.selectedIndex < s.scrollOffset {
		s.scrollOffset = s.selectedIndex
	} else if s.selectedIndex >= s.scrollOffset+s.maxVisibleItems {
		s.scrollOffset = s.selectedIndex - s.maxVisibleItems + 1
	}
}

// Draw draws the select list
func (s *SelectList) Draw(screen tcell.Screen) {
	if !s.visible {
		return
	}

	// Clear the component area
	for y := s.y; y < s.y+s.height; y++ {
		for x := s.x; x < s.x+s.width; x++ {
			screen.SetContent(x, y, ' ', nil, s.normalStyle)
		}
	}

	// Calculate visible options
	endVisible := s.scrollOffset + s.maxVisibleItems
	if endVisible > len(s.options) {
		endVisible = len(s.options)
	}

	// Draw visible options
	for i := s.scrollOffset; i < endVisible; i++ {
		y := s.y + (i-s.scrollOffset)*s.itemHeight
		option := s.options[i]

		// Use selected style if this is the selected option
		style := s.normalStyle
		if i == s.selectedIndex {
			style = s.selectedStyle
		}

		// Draw option text
		for x, r := range option.Text {
			if x >= s.width {
				break
			}
			screen.SetContent(s.x+x, y, r, nil, style)
		}

		// Fill the rest of the line with the style background
		for x := len(option.Text); x < s.width; x++ {
			screen.SetContent(s.x+x, y, ' ', nil, style)
		}
	}
}

// HandleEvent handles events for the select list
func (s *SelectList) HandleEvent(event tcell.Event) bool {
	if !s.visible {
		return false
	}

	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyUp:
			if s.selectedIndex > 0 {
				s.selectedIndex--
				s.ensureSelectedVisible()
				if s.onChange != nil {
					s.onChange(s.selectedIndex, s.options[s.selectedIndex])
				}
				return true
			}
		case tcell.KeyDown:
			if s.selectedIndex < len(s.options)-1 {
				s.selectedIndex++
				s.ensureSelectedVisible()
				if s.onChange != nil {
					s.onChange(s.selectedIndex, s.options[s.selectedIndex])
				}
				return true
			}
		case tcell.KeyEnter:
			if len(s.options) > 0 && s.onChange != nil {
				s.onChange(s.selectedIndex, s.options[s.selectedIndex])
			}
			return true
		}
	}

	return false
}
