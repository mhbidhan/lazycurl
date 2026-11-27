package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/theme"
)

// What PopupBox adds around the content it wraps: a one-cell border on each
// side plus two cells of padding, so six columns and four rows.
const (
	PopupFrameWidth  = 6
	PopupFrameHeight = 4
)

// PopupBox draws content inside the rounded border that every popup shares.
func PopupBox(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Primary)).
		Padding(1, 2).
		Render(content)
}

// PopupContentSize is how much content a PopupBox can wrap without overflowing a
// w x h container: the requested size, capped by the space left over once the
// frame is subtracted. Owners of a PopupManager should size against this so the
// popup shrinks with the terminal instead of spilling out of it.
//
// The result never drops below one cell: a container too small for the frame
// cannot be made to fit, and rendering something degenerate beats rendering
// nothing.
func PopupContentSize(w, h, wantW, wantH int) (int, int) {
	return max(min(wantW, w-PopupFrameWidth), 1), max(min(wantH, h-PopupFrameHeight), 1)
}
