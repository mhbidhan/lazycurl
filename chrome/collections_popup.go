package chrome

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/mhbidhan/lazycurl/ui"
	"gorm.io/gorm"
)

// state is which part of the popup has the keyboard. Every state but stateList
// either owns an input or has a sub-popup open in pm.
type collectionsState int

const (
	stateList collectionsState = iota
	stateCreate
	stateSearch
	stateConfirmDelete
	// stateRename covers both halves of a rename: the name prompt and then the
	// confirmation. Both live in pm, so the state only has to mean "a sub-popup
	// is open".
	stateRename
)

// Ids of the popups the collections popup nests inside itself.
const (
	newCollectionPopup  PopupID = "new_collection"
	editCollectionPopup PopupID = "edit_collection"
	confirmPopup        PopupID = "confirm"
)

// CollectionsPopup browses collections and owns the flows that change them.
// Everything it does beyond browsing lives in its own file, named after the
// flow: collections_list.go for the list widget itself, then
// collections_search.go, collections_create.go, collections_delete.go and
// collections_rename.go.
type CollectionsPopup struct {
	db      *gorm.DB
	keybind *keybind.CollectionKeys
	pm      *PopupManager

	// width and height are the area inside this popup's own frame. Kept because
	// a sub-popup is centered in all of it, which is more room than the list
	// gives up to the search field.
	width, height int

	state collectionsState
	list  *collectionList
	input textinput.Model

	// target is the collection a flow is asking about: the one pending
	// deletion, or the one being renamed. It is held here because the answer
	// arrives later as a message, by which time the cursor can no longer be
	// trusted to point at the same row.
	target *store.Collection
	// pendingName is the name the rename confirmation is asking to commit.
	pendingName string
}

func NewCollectionsPopup(db *gorm.DB, w, h int) *CollectionsPopup {
	pm := NewPopupManager()
	namePromptPopup := NewNamePromptPopup("New collection: ", "")
	namePromptPopup.SetSize(w, h)
	pm.Register(newCollectionPopup, namePromptPopup)

	c := &CollectionsPopup{
		db:      db,
		keybind: keybind.NewCollectionKeys(),
		pm:      pm,

		state: stateList,
		list:  newCollectionList(w, h),
		input: newSearchInput(),
	}

	c.reload()

	return c
}

// reload re-reads the rows from the store and redraws. Every flow that changes
// the data ends here.
func (c *CollectionsPopup) reload() {
	rows, err := store.GetCollections(c.db)
	if err != nil {
		rows = nil
	}
	c.list.setRows(rows)
	c.list.syncView()
}

func (c *CollectionsPopup) Init() tea.Cmd {
	return nil
}

// Update is a dispatch table. The handling for each state belongs to the file
// named after that flow, so this is the whole of the routing.
func (c *CollectionsPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch c.state {
	case stateList:
		return c, c.updateList(msg)
	case stateCreate:
		return c, c.updateCreate(msg)
	case stateSearch:
		return c, c.updateSearch(msg)
	case stateConfirmDelete:
		return c, c.updateDelete(msg)
	case stateRename:
		return c, c.updateRename(msg)
	}

	return c, nil
}

// updateList is the popup's key table. Everything it does not match goes to the
// list itself, which is what makes the mouse wheel work.
func (c *CollectionsPopup) updateList(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return c.list.update(msg)
	}

	switch {
	case key.Matches(keyMsg, c.keybind.Escape):
		return func() tea.Msg { return DismissPopup{} }
	case key.Matches(keyMsg, c.keybind.CursorDown):
		c.list.moveDown()
		return nil
	case key.Matches(keyMsg, c.keybind.CursorUp):
		c.list.moveUp()
		return nil
	case key.Matches(keyMsg, c.keybind.New):
		c.openCreate()
		return textinput.Blink
	case key.Matches(keyMsg, c.keybind.Search):
		c.openSearch()
		return textinput.Blink
	case key.Matches(keyMsg, c.keybind.Delete):
		c.askDelete()
		return nil
	case key.Matches(keyMsg, c.keybind.Edit):
		c.askRename()
		return nil
	}

	return c.list.update(msg)
}

// SetSize lays out the popup around the area inside its own frame. The list gets
// what the search field does not need, and the sub-popups get what is left once
// their frame is accounted for, so every level of nesting stays inside the one
// above it.
func (c *CollectionsPopup) SetSize(w, h int) {
	c.width, c.height = max(w, 1), max(h, 1)

	// Room for the filter field is reserved whether or not a search is running,
	// which keeps the list from changing height when one opens.
	c.list.SetSize(c.width, max(c.height-searchFieldRows, 1))
	c.input.Width = fieldWidth(c.width, searchPrompt)

	// A sub-popup brings its own frame, so it is sized against what is left of
	// this popup's content area once that frame is subtracted.
	c.pm.SetSize(
		max(c.width-ui.PopupFrameWidth, 1),
		max(c.height-ui.PopupFrameHeight, 1),
	)
}

func (c *CollectionsPopup) View() string {
	var content string
	help := c.keybind.ShortHelp()

	if _, m := c.pm.Active(); m != nil {
		// Crop before centering for the same reason overlay does: a sub-popup
		// a column too wide would otherwise wrap and break its own frame. The
		// area is the whole content area, not the list's, so a short sub-popup
		// is not cropped for being shorter than the list.

		sub := lipgloss.NewStyle().
			MaxWidth(c.width).
			MaxHeight(c.height).
			Render(m.View())

		content = lipgloss.JoinVertical(lipgloss.Top, sub, ui.GapVertical(1), help)

		return lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Width(c.width).
			Height(c.height).
			Render(content)
	}

	if c.state == stateSearch {
		content = lipgloss.JoinVertical(lipgloss.Top, c.searchView(), ui.GapVertical(1), help)
		return ui.PopupBox(content)
	}

	content = lipgloss.JoinVertical(lipgloss.Top, c.list.view(), ui.GapVertical(1), help)
	return ui.PopupBox(content)
}
