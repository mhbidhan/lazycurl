package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Chrome is a fixed-height block rendered above or below the viewport.
// Height must match the number of lines View returns, otherwise the
// reservation made for it in Update will be wrong.
type Chrome interface {
	Height() int
	View() string
}

type Viewport struct {
	viewport viewport.Model
	header   Chrome
	footer   Chrome
}

func NewViewport(width, height int, header, footer Chrome) *Viewport {
	return &Viewport{
		viewport: viewport.New(width, height),
		header:   header,
		footer:   footer,
	}
}

func (v *Viewport) Width() int {
	return v.viewport.Width
}

func (v *Viewport) Height() int {
	return v.viewport.Height
}

// ContentWidth is the width available to content once the viewport style's
// padding, border and margins are subtracted.
func (v *Viewport) ContentWidth() int {
	return v.viewport.Width - v.viewport.Style.GetHorizontalFrameSize()
}

// ContentHeight is the height available to content once the viewport style's
// padding, border and margins are subtracted.
func (v *Viewport) ContentHeight() int {
	return v.viewport.Height - v.viewport.Style.GetVerticalFrameSize()
}

// SetStyle replaces the style applied around viewport content.
func (v *Viewport) SetStyle(style lipgloss.Style) {
	v.viewport.Style = style
}

func (v *Viewport) Init() tea.Cmd { return nil }

func (v *Viewport) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if size, ok := msg.(tea.WindowSizeMsg); ok {
		h := size.Height
		if v.header != nil {
			h -= v.header.Height()
		}
		if v.footer != nil {
			h -= v.footer.Height()
		}

		v.viewport.Width = max(size.Width, 1)
		v.viewport.Height = max(h, 1)
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

func (v *Viewport) SetContent(content string) {
	v.viewport.SetContent(content)
}

func (v *Viewport) View() string {
	var b strings.Builder

	if v.header != nil {
		b.WriteString(v.header.View())
		b.WriteRune('\n')
	}

	b.WriteString(v.viewport.View())

	if v.footer != nil {
		b.WriteRune('\n')
		b.WriteString(v.footer.View())
	}

	return b.String()
}
