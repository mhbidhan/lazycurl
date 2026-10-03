package screen

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Root struct {
	currentScreen tea.Model
	homeScreen    *HomeScreen
}

func NewRoot() *Root {
	s := NewHomeScreen(0, 0)
	return &Root{
		currentScreen: s,
		homeScreen:    s,
	}
}

func (r *Root) Init() tea.Cmd {
	return nil
}

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		if r.homeScreen != nil {
			r.homeScreen.SetSize(ws.Width, ws.Height)
		}
	}

	var cmd tea.Cmd
	r.currentScreen, cmd = r.currentScreen.Update(msg)
	return r, cmd
}

func (r *Root) View() string {
	return r.currentScreen.View()
}
