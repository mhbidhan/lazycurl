package chrome

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpPopup struct {
	vp viewport.Model
}

func NewHelpPopup(bindings []Keybind, w, h int) *HelpPopup {
	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true
	vp.SetContent(RenderKeybinds(bindings))
	return &HelpPopup{vp: vp}
}

func (h *HelpPopup) Init() tea.Cmd {
	return nil
}

func (h *HelpPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	h.vp, cmd = h.vp.Update(msg)
	return h, cmd
}

func (h *HelpPopup) SetSize(w, height int) {
	h.vp.Width = w
	h.vp.Height = height
}

func (h *HelpPopup) View() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2).
		Render(h.vp.View())
}
