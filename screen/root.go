package screen

import (
	tea "github.com/charmbracelet/bubbletea"
	"gorm.io/gorm"
)

// Root is the Bubble Tea model that hosts the current screen.
type Root struct {
	currentScreen tea.Model
}

// NewRoot builds the root model with the home screen active.
func NewRoot(db *gorm.DB) *Root {
	return &Root{
		currentScreen: NewHomeScreen(db, 0, 0),
	}
}

func (r *Root) Init() tea.Cmd {
	return nil
}

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	r.currentScreen, cmd = r.currentScreen.Update(msg)
	return r, cmd
}

func (r *Root) View() string {
	return r.currentScreen.View()
}
