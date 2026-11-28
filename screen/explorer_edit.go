package screen

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/store"
)

// openNamePrompt registers a fresh prompt because NamePromptPopup keeps the
// initial value it was built with.
func (s *ExplorerScreen) openNamePrompt(prompt, initial string) {
	s.pm.Register(explorerNamePopup, chrome.NewNamePromptPopup(prompt, initial))
	s.pm.Open(explorerNamePopup)
	s.state = stateName
}

func (s *ExplorerScreen) clearPending() {
	s.pendingType = ""
	s.pendingParent = nil
	s.pendingAfter = 0
	s.pendingID = 0
}

// createNameAndType applies the one-prompt create rule: a trailing slash makes
// a directory and is stripped from the stored name. Otherwise the fallback
// type wins, which is a request for a and a directory for A.
func createNameAndType(input string, fallback store.NodeType) (string, store.NodeType) {
	name := strings.TrimSpace(input)
	typ := fallback
	if strings.HasSuffix(name, "/") {
		typ = store.NodeTypeDirectory
		name = strings.TrimSuffix(name, "/")
	}
	return strings.TrimSpace(name), typ
}

func (s *ExplorerScreen) updateName(msg tea.Msg) tea.Cmd {
	ev, ok := msg.(chrome.NamePromptEvent)
	if !ok {
		return s.pm.Update(msg)
	}
	s.pm.Close(explorerNamePopup)

	if ev.Cancel {
		s.state = stateTree
		s.clearPending()
		return nil
	}
	name := strings.TrimSpace(ev.Value)
	if name == "" {
		s.state = stateTree
		s.clearPending()
		s.message = "Name cannot be empty"
		return nil
	}

	if s.pendingID != 0 {
		id := s.pendingID
		err := store.RenameNode(s.db, id, name)
		s.clearPending()
		s.state = stateTree
		err = errors.Join(err, s.reload())
		s.tree.revealID(id)
		return s.report(err)
	}

	name, typ := createNameAndType(name, s.pendingType)
	if name == "" {
		s.state = stateTree
		s.clearPending()
		s.message = "Name cannot be empty"
		return nil
	}

	parentID, afterID := s.pendingParent, s.pendingAfter
	s.clearPending()
	s.state = stateTree
	node, err := store.CreateNode(s.db, s.collection.ID, parentID, afterID, name, typ)
	err = errors.Join(err, s.reload())
	if err == nil {
		s.tree.revealID(node.ID)
	}
	return s.report(err)
}

// addChild creates a request under the selected directory, or at the collection
// root when the selection is empty or a request. A trailing / in the name
// makes it a directory instead.
func (s *ExplorerScreen) addChild() {
	s.clearPending()
	s.pendingType = store.NodeTypeRequest
	if row, ok := s.tree.current(); ok && row.Node.Type == store.NodeTypeDirectory {
		id := row.Node.ID
		s.pendingParent = &id
	}
	s.openNamePrompt("New node name: ", "")
}

// addSibling creates a directory after the selected node. With no selection it
// adds a root-level directory.
func (s *ExplorerScreen) addSibling() {
	s.clearPending()
	s.pendingType = store.NodeTypeDirectory
	if row, ok := s.tree.current(); ok {
		if row.Node.ParentID != nil {
			parentID := *row.Node.ParentID
			s.pendingParent = &parentID
		}
		s.pendingAfter = row.Node.ID
	}
	s.openNamePrompt("New node name: ", "")
}

func (s *ExplorerScreen) askRename() {
	row, ok := s.tree.current()
	if !ok {
		return
	}
	s.clearPending()
	s.pendingID = row.Node.ID
	s.openNamePrompt("Rename: ", row.Node.Name)
}

func (s *ExplorerScreen) askDelete() {
	row, ok := s.tree.current()
	if !ok {
		return
	}
	s.clearPending()
	s.pendingID = row.Node.ID
	s.pm.Register(explorerDeletePopup, chrome.NewConfirmPopup(fmt.Sprintf("Delete %q?", row.Node.Name)))
	s.pm.Open(explorerDeletePopup)
	s.state = stateDelete
}

func (s *ExplorerScreen) updateDelete(msg tea.Msg) tea.Cmd {
	ev, ok := msg.(chrome.ConfirmEvent)
	if !ok {
		return s.pm.Update(msg)
	}
	s.pm.Close(explorerDeletePopup)
	s.state = stateTree

	var err error
	if ev.Confirmed && s.pendingID != 0 {
		err = store.DeleteNode(s.db, s.pendingID)
	}
	s.clearPending()
	err = errors.Join(err, s.reload())
	return s.report(err)
}

func (s *ExplorerScreen) markCurrent(mode explorerMarkMode) {
	row, ok := s.tree.current()
	if !ok || mode == markNone {
		return
	}
	s.marked = row.Node.ID
	s.markMode = mode
	s.err = nil
	verb := "move"
	if mode == markCopy {
		verb = "copy"
	}
	s.message = fmt.Sprintf("Marked %q to %s; move to target and press p", row.Node.Name, verb)
}

// pasteMarked moves or copies the marked node inside the selected directory,
// or after the selected non-directory node. Root is the target when nothing is
// selected. A rejected paste keeps the mark so the target can be changed.
func (s *ExplorerScreen) pasteMarked() tea.Cmd {
	if s.marked == 0 || s.markMode == markNone {
		s.message = "Nothing marked to paste"
		return nil
	}

	var parentID *uint
	var afterID uint
	if row, ok := s.tree.current(); ok {
		if row.Node.Type == store.NodeTypeDirectory {
			id := row.Node.ID
			parentID = &id
		} else {
			if row.Node.ParentID != nil {
				parent := *row.Node.ParentID
				parentID = &parent
			}
			afterID = row.Node.ID
		}
	}

	id := s.marked
	mode := s.markMode
	var copied store.Node
	var err error
	if mode == markCopy {
		copied, err = store.CopyNode(s.db, id, parentID, afterID)
	} else {
		err = store.MoveNode(s.db, id, parentID, afterID)
	}
	if err == nil {
		s.marked = 0
		s.markMode = markNone
	}
	err = errors.Join(err, s.reload())

	// Reveal the result: the moved node for x, the new clone for c.
	reveal := id
	if copied.ID != 0 {
		reveal = copied.ID
	}
	s.tree.revealID(reveal)
	return s.report(err)
}

func (s *ExplorerScreen) reorder(delta int) tea.Cmd {
	id := s.tree.currentID()
	if id == 0 {
		return nil
	}
	err := store.ReorderNode(s.db, id, delta)
	err = errors.Join(err, s.reload())
	s.tree.revealID(id)
	return s.report(err)
}
