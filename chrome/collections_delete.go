package chrome

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/store"
)

// askDelete asks before deleting the selected collection. The confirm popup is
// registered per invocation because it carries the question.
func (c *CollectionsPopup) askDelete() {
	coll, ok := c.list.current()
	if !ok {
		return
	}

	c.target = &coll
	c.state = stateConfirmDelete
	c.pm.Register(confirmPopup, NewConfirmPopup(fmt.Sprintf("Delete %q?", coll.Name)))
	c.pm.Open(confirmPopup)
}

func (c *CollectionsPopup) updateDelete(msg tea.Msg) tea.Cmd {
	ev, ok := msg.(ConfirmEvent)
	if !ok {
		return c.pm.Update(msg)
	}

	c.pm.Close(confirmPopup)
	if ev.Confirmed && c.target != nil {
		_ = store.DeleteCollection(c.db, c.target.ID)
	}
	c.target = nil
	c.state = stateList
	c.reload()

	return nil
}
