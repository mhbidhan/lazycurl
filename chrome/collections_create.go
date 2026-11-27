package chrome

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/store"
)

// openCreate shows the name prompt for a brand new collection.
func (c *CollectionsPopup) openCreate() {
	c.state = stateCreate
	// Open now rather than on the next message, so the sub-popup is already
	// active when View runs for this transition.
	c.pm.Open(newCollectionPopup)
}

// updateCreate consumes the name prompt's answer. Creating needs no
// confirmation: nothing existing is overwritten.
func (c *CollectionsPopup) updateCreate(msg tea.Msg) tea.Cmd {
	ev, ok := msg.(NamePromptEvent)
	if !ok {
		return c.pm.Update(msg)
	}

	if name := strings.TrimSpace(ev.Value); !ev.Cancel && name != "" {
		_ = store.SaveCollection(c.db, store.NewCollection(name))
		c.reload()
	}

	c.state = stateList
	c.pm.Close(newCollectionPopup)
	c.list.syncView()

	return nil
}
