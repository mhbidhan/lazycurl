package chrome

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/store"
)

// askRename shows the name prompt for the selected collection, pre-filled with
// its current name so there is something to edit.
func (c *CollectionsPopup) askRename() {
	coll, ok := c.list.current()
	if !ok {
		return
	}

	c.target = &coll
	c.state = stateRename
	c.pm.Register(editCollectionPopup, NewNamePromptPopup(fmt.Sprintf("Rename %q to: ", coll.Name), coll.Name))
	c.pm.Open(editCollectionPopup)
}

// updateRename drives both halves of a rename: the name prompt, then the
// confirmation it leads to.
func (c *CollectionsPopup) updateRename(msg tea.Msg) tea.Cmd {
	if ev, ok := msg.(NamePromptEvent); ok {
		c.askRenameConfirm(ev)
		return nil
	}

	if ev, ok := msg.(ConfirmEvent); ok {
		c.pm.Close(confirmPopup)
		return c.finishRename(ev.Confirmed)
	}

	return c.pm.Update(msg)
}

// askRenameConfirm takes the typed name and asks before committing it. A
// cancelled, blank, or unchanged name goes straight back to the list: there is
// nothing to confirm.
func (c *CollectionsPopup) askRenameConfirm(ev NamePromptEvent) {
	c.pm.Close(editCollectionPopup)

	name := strings.TrimSpace(ev.Value)
	if ev.Cancel || c.target == nil || name == "" || name == c.target.Name {
		c.target = nil
		c.state = stateList
		c.list.syncView()
		return
	}

	c.pendingName = name
	c.pm.Register(confirmPopup, NewConfirmPopup(fmt.Sprintf("Rename %q to %q?", c.target.Name, name)))
	c.pm.Open(confirmPopup)
}

func (c *CollectionsPopup) finishRename(confirmed bool) tea.Cmd {
	var err error
	if confirmed && c.target != nil {
		coll := *c.target
		coll.Name = c.pendingName
		err = store.SaveCollection(c.db, &coll)
	}
	err = errors.Join(err, c.reload())

	c.target = nil
	c.pendingName = ""
	c.state = stateList

	return c.report(err)
}
