package chrome

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/ui"
)

const searchPrompt = "/ "

// searchFieldRows is what the filter field and the gap under it take off the
// list's height while a search is on screen.
const searchFieldRows = 2

func newSearchInput() textinput.Model {
	input := textinput.New()
	input.CharLimit = 50
	input.Width = 30

	return input
}

// openSearch starts filtering: an empty field showing every row.
func (c *CollectionsPopup) openSearch() {
	c.state = stateSearch
	c.input.SetValue("")
	c.input.Focus()
	c.input.Prompt = searchPrompt
	c.list.clearQuery()
}

// closeSearch stops filtering but keeps whatever the user typed. Abandoning the
// search is clearQuery on top of this.
func (c *CollectionsPopup) closeSearch() {
	c.state = stateList
	c.input.Blur()
}

func (c *CollectionsPopup) updateSearch(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			// Keep the filter, just stop editing it.
			c.closeSearch()
			c.list.refresh()
			return nil
		case tea.KeyEsc, tea.KeyCtrlC:
			c.closeSearch()
			c.list.clearQuery()
			c.list.refresh()
			return nil
		}
	}

	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	c.list.setQuery(c.input.Value())
	c.list.refresh()

	return cmd
}

// searchView draws the filter field above the list.
func (c *CollectionsPopup) searchView() string {
	inputView := lipgloss.NewStyle().Render(c.input.View())

	return lipgloss.JoinVertical(lipgloss.Top, inputView, ui.GapVertical(1), c.list.view())
}
