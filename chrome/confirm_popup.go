package chrome

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
)

// ConfirmEvent is the answer to a ConfirmPopup.
type ConfirmEvent struct {
	Confirmed bool
}

// ConfirmPopup asks a yes/no question. Like the other popups it reports rather
// than acts: the answer comes back as a ConfirmEvent and closing is left to
// whoever opened it.
type ConfirmPopup struct {
	width, height int

	question string
	keybinds *keybind.ConfirmKeys
}

func NewConfirmPopup(question string) *ConfirmPopup {
	return &ConfirmPopup{
		question: question,
		keybinds: keybind.NewConfirmKeys(),
	}
}

func (c *ConfirmPopup) Init() tea.Cmd {
	return nil
}

func (c *ConfirmPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return c, nil
	}

	switch {
	case key.Matches(keyMsg, c.keybinds.Confirm):
		return c, func() tea.Msg {
			return ConfirmEvent{Confirmed: true}
		}
	case key.Matches(keyMsg, c.keybinds.Cancel):
		return c, func() tea.Msg {
			return ConfirmEvent{Confirmed: false}
		}
	}

	return c, nil
}

// SetSize records the content area to work with. The question is given the
// height left over once the help line is measured, rather than the whole area,
// so the frame stays inside its container.
func (c *ConfirmPopup) SetSize(w, h int) {
	c.width = max(w, 1)
	c.height = max(h, 1)
}

// View spends the available height on the question first, then the gap, then the
// help. The help is measured at the popup's width because it wraps on a narrow
// terminal, and the question takes whatever it ends up leaving.
func (c *ConfirmPopup) View() string {
	help := lipgloss.NewStyle().Width(c.width).Render(c.keybinds.ShortHelp())
	questionHeight := max(c.height-lipgloss.Height(help)-1, 1)

	question := lipgloss.NewStyle().
		Width(c.width).
		Height(questionHeight).
		Bold(true).
		Render(c.question)

	content := lipgloss.JoinVertical(lipgloss.Center, question, ui.GapVertical(1), help)

	return ui.PopupBox(content)
}
