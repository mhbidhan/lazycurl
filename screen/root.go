package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
)

type Root struct {
	kb       *keybind.RootKeys
	viewport *ui.Viewport
}

func NewRoot() *Root {
	kb := keybind.NewRootKeys()
	v := ui.NewViewport(20, 20, nil, nil)

	return &Root{
		kb:       kb,
		viewport: v,
	}
}

func (r *Root) Init() tea.Cmd {
	r.setContent()
	return nil
}

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	redraw := false

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, r.kb.Quit):
			return r, tea.Quit
		}
	}

	if _, ok := msg.(tea.WindowSizeMsg); ok {
		redraw = true
	}

	cmd := r.viewport.Update(msg)

	if redraw {
		r.setContent()
	}

	return r, cmd
}

func (r *Root) setContent() {
	container := lipgloss.NewStyle().
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center).
		Width(r.viewport.ContentWidth()).
		Height(r.viewport.ContentHeight())

	body := r.kb.ShortHelp()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		ui.Logo,
		ui.GapVertical(1),
		body,
	)

	r.viewport.SetContent(container.Render(content))
}

func (r *Root) View() string {
	return r.viewport.View()
}
