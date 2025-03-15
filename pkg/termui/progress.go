package termui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// ProgressBar represents a progress indicator
type ProgressBar struct {
	Container
	progress     float64 // 0.0 to 1.0
	showPercent  bool
	barStyle     tcell.Style
	emptyStyle   tcell.Style
	percentStyle tcell.Style
	caption      string
	captionStyle tcell.Style
}

// NewProgressBar creates a new progress bar
func NewProgressBar(x, y, width, height int) *ProgressBar {
	return &ProgressBar{
		Container: Container{
			x:       x,
			y:       y,
			width:   width,
			height:  height,
			visible: true,
		},
		progress:     0.0,
		showPercent:  true,
		barStyle:     tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorGreen),
		emptyStyle:   tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorGray),
		percentStyle: tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		captionStyle: tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
	}
}

// SetProgress sets the progress value (0.0 to 1.0)
func (p *ProgressBar) SetProgress(progress float64) *ProgressBar {
	if progress < 0.0 {
		progress = 0.0
	} else if progress > 1.0 {
		progress = 1.0
	}
	p.progress = progress
	return p
}

// SetShowPercent sets whether to show the percentage
func (p *ProgressBar) SetShowPercent(show bool) *ProgressBar {
	p.showPercent = show
	return p
}

// SetCaption sets the caption text
func (p *ProgressBar) SetCaption(caption string) *ProgressBar {
	p.caption = caption
	return p
}

// SetStyles sets the styles for the progress bar
func (p *ProgressBar) SetStyles(barStyle, emptyStyle, percentStyle, captionStyle tcell.Style) *ProgressBar {
	p.barStyle = barStyle
	p.emptyStyle = emptyStyle
	p.percentStyle = percentStyle
	p.captionStyle = captionStyle
	return p
}

// Draw draws the progress bar
func (p *ProgressBar) Draw(screen tcell.Screen) {
	if !p.visible {
		return
	}

	// Calculate bar width
	barWidth := int(float64(p.width) * p.progress)
	if barWidth > p.width {
		barWidth = p.width
	}

	// Draw caption if set
	if p.caption != "" && p.height > 1 {
		for i, r := range p.caption {
			if i >= p.width {
				break
			}
			screen.SetContent(p.x+i, p.y, r, nil, p.captionStyle)
		}
	}

	// Determine the y position for the bar
	barY := p.y
	if p.caption != "" && p.height > 1 {
		barY = p.y + 1
	}

	// Draw progress bar
	for x := 0; x < p.width; x++ {
		if x < barWidth {
			screen.SetContent(p.x+x, barY, ' ', nil, p.barStyle)
		} else {
			screen.SetContent(p.x+x, barY, ' ', nil, p.emptyStyle)
		}
	}

	// Draw percentage if enabled
	if p.showPercent {
		percent := fmt.Sprintf(" %d%% ", int(p.progress*100))
		// Center the percentage text
		textX := p.x + (p.width-len(percent))/2
		for i, r := range percent {
			if textX+i >= p.x+p.width {
				break
			}
			screen.SetContent(textX+i, barY, r, nil, p.percentStyle)
		}
	}
}

// HandleEvent handles events for the progress bar
func (p *ProgressBar) HandleEvent(event tcell.Event) bool {
	// Progress bars don't typically handle events
	return false
}
