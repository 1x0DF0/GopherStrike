package termui

import (
	"github.com/gdamore/tcell/v2"
)

// StyleManager handles styles for UI components
type StyleManager struct {
	styles map[string]tcell.Style
}

// NewStyleManager creates a new style manager with default styles
func NewStyleManager() *StyleManager {
	sm := &StyleManager{
		styles: make(map[string]tcell.Style),
	}

	// Initialize default styles
	sm.styles["default"] = tcell.StyleDefault
	sm.styles["title"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlue).
		Bold(true)
	sm.styles["border"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite)
	sm.styles["selected"] = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorYellow)
	sm.styles["focused"] = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorGreen)
	sm.styles["button"] = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorSilver)
	sm.styles["button.focused"] = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorYellow)
	sm.styles["error"] = tcell.StyleDefault.
		Foreground(tcell.ColorRed).
		Bold(true)
	sm.styles["success"] = tcell.StyleDefault.
		Foreground(tcell.ColorGreen).
		Bold(true)
	sm.styles["warning"] = tcell.StyleDefault.
		Foreground(tcell.ColorYellow).
		Bold(true)
	sm.styles["info"] = tcell.StyleDefault.
		Foreground(tcell.ColorTeal)
	sm.styles["disabled"] = tcell.StyleDefault.
		Foreground(tcell.ColorGray)
	sm.styles["dialog"] = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorSilver)
	sm.styles["dialog.title"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorNavy).
		Bold(true)
	sm.styles["progress"] = tcell.StyleDefault.
		Foreground(tcell.ColorGreen).
		Background(tcell.ColorBlack)
	sm.styles["progress.bar"] = tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorGreen)
	sm.styles["menu"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlue)
	sm.styles["menu.selected"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorNavy).
		Bold(true)

	// Tool category styles
	sm.styles["category.scanner"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlue)
	sm.styles["category.recon"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDarkGreen)
	sm.styles["category.vuln"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorMaroon)
	sm.styles["category.reporting"] = tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorPurple)

	return sm
}

// Get returns a style by name
func (s *StyleManager) Get(name string) tcell.Style {
	if style, ok := s.styles[name]; ok {
		return style
	}
	return s.styles["default"]
}

// Set sets a style by name
func (s *StyleManager) Set(name string, style tcell.Style) {
	s.styles[name] = style
}

// Derives a new style from an existing one with modifications
func (s *StyleManager) Derive(base string, fg, bg tcell.Color, attrs ...tcell.AttrMask) tcell.Style {
	style := s.Get(base)

	if fg != tcell.ColorDefault {
		style = style.Foreground(fg)
	}

	if bg != tcell.ColorDefault {
		style = style.Background(bg)
	}

	for _, attr := range attrs {
		if attr == tcell.AttrBold {
			style = style.Bold(true)
		} else if attr == tcell.AttrBlink {
			style = style.Blink(true)
		} else if attr == tcell.AttrReverse {
			style = style.Reverse(true)
		} else if attr == tcell.AttrUnderline {
			style = style.Underline(true)
		} else if attr == tcell.AttrDim {
			style = style.Dim(true)
		} else if attr == tcell.AttrItalic {
			style = style.Italic(true)
		} else if attr == tcell.AttrStrikeThrough {
			style = style.StrikeThrough(true)
		}
	}

	return style
}
