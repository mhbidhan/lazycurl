package screen

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/ui"
)

const explorerSearchPrompt = "/ "

func newExplorerSearchInput() textinput.Model {
	input := textinput.New()
	input.CharLimit = 50
	input.Prompt = explorerSearchPrompt
	return input
}

// openSearch starts filtering from an empty field, so the whole tree is visible
// again the moment the filter opens.
func (s *ExplorerScreen) openSearch() {
	s.state = stateSearch
	s.input.SetValue("")
	s.input.CursorEnd()
	s.input.Focus()
	s.tree.setQuery("")
}

func (s *ExplorerScreen) closeSearch(clear bool) {
	s.state = stateTree
	s.input.Blur()
	if clear {
		s.tree.setQuery("")
	}
}

func (s *ExplorerScreen) updateSearch(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, s.keybind.Submit):
			s.closeSearch(false)
			return nil
		case key.Matches(keyMsg, s.keybind.Cancel):
			s.closeSearch(true)
			return nil
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.tree.setQuery(s.input.Value())
	return cmd
}

// searchView keeps the filter rows even while hidden so the tree never jumps
// when the user presses /.
func (s *ExplorerScreen) searchView() string {
	if s.state != stateSearch {
		return ui.GapVertical(explorerSearchRows)
	}
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Render(s.input.View()),
		ui.GapVertical(explorerSearchRows-1),
	)
}
