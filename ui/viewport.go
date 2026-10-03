package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Chrome is a fixed-height block rendered above or below the viewport.
// Height must match the number of lines View returns, otherwise the
// reservation made for it in Update will be wrong.
type Chrome interface {
	Height() int
	View() string
}

type Viewport struct {
	model  viewport.Model
	header Chrome
	footer Chrome
}

func NewViewport(width, height int, header, footer Chrome) *Viewport {
	return &Viewport{
		model:  viewport.New(width, height),
		header: header,
		footer: footer,
	}
}

func (v *Viewport) Width() int {
	return v.model.Width
}

func (v *Viewport) Height() int {
	return v.model.Height
}

func (v *Viewport) ContentWidth() int {
	return v.model.Width - v.model.Style.GetHorizontalFrameSize()
}

func (v *Viewport) ContentHeight() int {
	return v.model.Height - v.model.Style.GetVerticalFrameSize()
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

		v.model.Width = max(size.Width, 1)
		v.model.Height = max(h, 1)
	}

	v.model, cmd = v.model.Update(msg)
	return cmd
}

func (v *Viewport) SetContent(content string) {
	v.model.SetContent(content)
}

func (v *Viewport) View() string {
	var b strings.Builder

	if v.header != nil {
		b.WriteString(v.header.View())
		b.WriteRune('\n')
	}

	b.WriteString(v.model.View())

	if v.footer != nil {
		b.WriteRune('\n')
		b.WriteString(v.footer.View())
	}

	return b.String()
}
