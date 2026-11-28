package screen

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/chrome"
	"gorm.io/gorm"
)

// Root is the Bubble Tea model that hosts the current screen. It owns screen
// changes so neither HomeScreen nor ExplorerScreen needs a reference to the
// other.
type Root struct {
	db *gorm.DB

	home     *HomeScreen
	explorer *ExplorerScreen
	current  tea.Model

	width, height int
}

// NewRoot builds the root model with the home screen active.
func NewRoot(db *gorm.DB) *Root {
	home := NewHomeScreen(db, 0, 0)
	return &Root{
		db:      db,
		home:    home,
		current: home,
	}
}

// SetSize resizes every screen that exists, including the inactive one, so a
// resize in the explorer is not lost when the user returns home.
func (r *Root) SetSize(w, h int) {
	r.width, r.height = w, h
	r.home.SetSize(w, h)
	if r.explorer != nil {
		r.explorer.SetSize(w, h)
	}
}

func (r *Root) Init() tea.Cmd {
	return nil
}

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case chrome.OpenCollectionMsg:
		// The popup is about to be replaced, but closing it keeps the home
		// screen clean for the next visit.
		r.home.pm.Close(chrome.PopupCollections)
		r.explorer = NewExplorerScreen(r.db, m.Collection, r.width, r.height)
		r.current = r.explorer
		return r, nil
	case BackToHomeMsg:
		r.current = r.home
		return r, nil
	case tea.WindowSizeMsg:
		r.SetSize(m.Width, m.Height)
	}

	var cmd tea.Cmd
	r.current, cmd = r.current.Update(msg)
	return r, cmd
}

func (r *Root) View() string {
	return r.current.View()
}
