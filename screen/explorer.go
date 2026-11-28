package screen

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/mhbidhan/lazycurl/theme"
	"github.com/mhbidhan/lazycurl/ui"
	"gorm.io/gorm"
)

// Rows the explorer reserves around the tree viewport. The filter field is
// reserved even while hidden so opening it cannot resize the tree.
const (
	explorerHeaderRows = 2
	explorerSearchRows = 2
	explorerFooterRows = 2
)

// Popup IDs owned by the explorer. Popup managers are per screen, so these do
// not collide with the collections popup's internal IDs.
const (
	explorerHelpPopup   chrome.PopupID = "explorer_help"
	explorerNamePopup   chrome.PopupID = "explorer_name"
	explorerDeletePopup chrome.PopupID = "explorer_delete"
)

type explorerState int

const (
	stateTree explorerState = iota
	stateName
	stateDelete
	stateSearch
)

// explorerMarkMode records what p will do with the marked node.
type explorerMarkMode int

const (
	markNone explorerMarkMode = iota
	markMove
	markCopy
)

// BackToHomeMsg asks Root to leave the explorer and restore the home screen.
type BackToHomeMsg struct{}

// ExplorerScreen is a full-screen tree browser for one collection.
type ExplorerScreen struct {
	db         *gorm.DB
	collection store.Collection
	width      int
	height     int

	keybind *keybind.ExplorerKeys
	tree    *explorerTree
	pm      *chrome.PopupManager
	input   textinput.Model

	state   explorerState
	err     error
	message string

	// marked is the id of the node waiting for p; markMode says whether p
	// moves it or copies it. markNone means nothing is marked.
	marked   uint
	markMode explorerMarkMode

	// pending carries the create/rename target while the name prompt is open.
	// pendingType is the fallback type for a create: a appends a request
	// unless the name ends in /, while A always creates a directory.
	pendingType   store.NodeType
	pendingParent *uint
	pendingAfter  uint
	pendingID     uint
}

// NewExplorerScreen builds the explorer for a collection and loads its tree.
func NewExplorerScreen(db *gorm.DB, coll store.Collection, w, h int) *ExplorerScreen {
	keys := keybind.NewExplorerKeys()
	s := &ExplorerScreen{
		db:         db,
		collection: coll,
		keybind:    keys,
		tree:       newExplorerTree(max(w, 1), max(h, 1)),
		pm:         chrome.NewPopupManager(),
		input:      newExplorerSearchInput(),
	}
	s.tree.setTitle("Collection: " + coll.Name)

	s.pm.Register(explorerHelpPopup, chrome.NewHelpPopup(keys.GetKeybinds(), 30, 12))
	s.pm.Register(explorerNamePopup, chrome.NewNamePromptPopup("Name: ", ""))

	s.SetSize(w, h)
	s.reload()
	return s
}

// SetSize sizes the tree against the whole screen, reserving rows for the
// title, the filter, the help line and the status line.
func (s *ExplorerScreen) SetSize(w, h int) {
	s.width, s.height = max(w, 1), max(h, 1)
	treeH := max(s.height-explorerHeaderRows-explorerSearchRows-explorerFooterRows, 1)
	s.tree.SetSize(s.width, treeH)
	s.input.Width = max(s.width-len(explorerSearchPrompt)-1, 1)

	contentW, contentH := ui.PopupContentSize(s.width, s.height, 40, 12)
	s.pm.SetSize(contentW, contentH)
}

// reload re-reads the collection's nodes and redraws the tree. Every edit ends
// here so the model and the view cannot drift apart.
func (s *ExplorerScreen) reload() error {
	nodes, err := store.GetNodes(s.db, s.collection.ID)
	s.tree.load(nodes)
	return err
}

func (s *ExplorerScreen) Init() tea.Cmd {
	return nil
}

func (s *ExplorerScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		s.SetSize(size.Width, size.Height)
	}

	// Errors arrive after the flow that caused them has returned to the tree,
	// so they are handled before any state or popup dispatch.
	if ev, ok := msg.(chrome.OperationError); ok {
		s.err = ev.Err
		s.message = ""
		return s, nil
	}
	if _, ok := msg.(chrome.DismissPopup); ok {
		if id, _ := s.pm.Active(); id != "" {
			s.pm.Close(id)
			if s.state != stateSearch {
				s.state = stateTree
			}
		}
		return s, nil
	}

	// A state with its own prompt or popup owns every message until it ends.
	// Its handler forwards the ones it does not understand, so routing through
	// a global pm.IsOpen() check here would swallow the NamePromptEvent and
	// ConfirmEvent those handlers need to see.
	switch s.state {
	case stateName:
		return s, s.updateName(msg)
	case stateDelete:
		return s, s.updateDelete(msg)
	case stateSearch:
		return s, s.updateSearch(msg)
	}

	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		if s.pm.IsOpen() {
			return s, s.pm.Update(msg)
		}
		var cmd tea.Cmd
		s.tree.vp, cmd = s.tree.vp.Update(msg)
		return s, cmd
	}
	if s.pm.IsOpen() {
		return s, s.pm.Update(keyMsg)
	}
	if key.Matches(keyMsg, s.keybind.Help) {
		s.pm.Open(explorerHelpPopup)
		return s, nil
	}
	return s, s.updateTreeKeys(keyMsg)
}

// updateTreeKeys is the tree state's key table. It only runs when no popup and
// no prompt owns the keyboard.
func (s *ExplorerScreen) updateTreeKeys(msg tea.KeyMsg) tea.Cmd {
	kb := s.keybind
	switch {
	case key.Matches(msg, kb.CursorDown):
		s.tree.move(1)
	case key.Matches(msg, kb.CursorUp):
		s.tree.move(-1)
	case key.Matches(msg, kb.Collapse):
		s.tree.setExpanded(false)
	case key.Matches(msg, kb.Expand):
		s.tree.setExpanded(true)
	case key.Matches(msg, kb.Toggle):
		s.tree.toggle()
	case key.Matches(msg, kb.Top):
		s.tree.top()
	case key.Matches(msg, kb.Bottom):
		s.tree.bottom()
	case key.Matches(msg, kb.Search):
		s.openSearch()
		return textinput.Blink
	case key.Matches(msg, kb.AddChild):
		s.addChild()
	case key.Matches(msg, kb.AddSibling):
		s.addSibling()
	case key.Matches(msg, kb.Delete):
		s.askDelete()
	case key.Matches(msg, kb.Rename):
		s.askRename()
	case key.Matches(msg, kb.Cut):
		s.markCurrent(markMove)
	case key.Matches(msg, kb.Copy):
		s.markCurrent(markCopy)
	case key.Matches(msg, kb.Paste):
		return s.pasteMarked()
	case key.Matches(msg, kb.MoveDown):
		return s.reorder(1)
	case key.Matches(msg, kb.MoveUp):
		return s.reorder(-1)
	case key.Matches(msg, kb.Back):
		return func() tea.Msg { return BackToHomeMsg{} }
	}
	return nil
}

func (s *ExplorerScreen) statusView() string {
	msg := s.message
	if s.err != nil {
		msg = "Error: " + s.err.Error()
	}
	return lipgloss.NewStyle().
		MaxWidth(s.width).
		Height(1).
		Foreground(lipgloss.Color(theme.Error)).
		Render(msg)
}

// report clears a stale status on success and turns a failure into a message
// for Update, keeping View free of state changes.
func (s *ExplorerScreen) report(err error) tea.Cmd {
	s.message = ""
	if err == nil {
		s.err = nil
		return nil
	}
	return func() tea.Msg { return chrome.OperationError{Err: err} }
}

func (s *ExplorerScreen) View() string {
	if _, m := s.pm.Active(); m != nil {
		return overlay(s.width, s.height, m.View())
	}

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		s.tree.view(),
		s.searchView(),
		s.keybind.ShortHelp(),
		s.statusView(),
	)
	return fillScreen(s.width, s.height, content)
}

// fillScreen crops content to the terminal and anchors it at the top-left. The
// explorer is a full screen, not a centered popup.
func fillScreen(w, h int, content string) string {
	cropped := lipgloss.NewStyle().MaxWidth(w).MaxHeight(h).Render(content)
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Top, lipgloss.Left).
		Render(cropped)
}
