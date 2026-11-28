package chrome

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/mhbidhan/lazycurl/theme"
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

// Rows under the list in every state: the status line and the help line.
// PopupBox supplies the padding below them, so only these two are reserved
// while sizing to keep the frame from growing past the area the screen gave it.
const footerRows = 2

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
	// err is the last failed operation, shown as a status line in the popup.
	err error
}

// NewCollectionsPopup builds the browser and loads its first page of rows.
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

	c.err = c.reload()

	return c
}

// reload re-reads the rows from the store and redraws. Every flow that changes
// the data ends here. Load failures are returned so the caller can show them.
func (c *CollectionsPopup) reload() error {
	rows, err := store.GetCollections(c.db)
	c.list.setRows(rows)
	c.list.syncView()
	return err
}

func (c *CollectionsPopup) Init() tea.Cmd {
	return nil
}

// Update is a dispatch table. The handling for each state belongs to the file
// named after that flow, so this is the whole of the routing.
func (c *CollectionsPopup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Errors arrive as commands, after the flow that caused them has already
	// returned to the list. Handle them before state dispatch so no state can
	// swallow the status line.
	if ev, ok := msg.(OperationError); ok {
		c.err = ev.Err
		return c, nil
	}

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
	case key.Matches(keyMsg, c.keybind.Open):
		if coll, ok := c.list.current(); ok {
			return func() tea.Msg { return OpenCollectionMsg{Collection: coll} }
		}
		return nil
	}

	return c.list.update(msg)
}

// SetSize lays out the popup around the area inside its own frame. The list gets
// what the search field, footer and status line do not need, and nested popups
// get what is left once their own frame is accounted for.
func (c *CollectionsPopup) SetSize(w, h int) {
	c.width, c.height = max(w, 1), max(h, 1)

	// Reserve every state's rows at once: opening a search or showing an error
	// must not change the frame's height.
	c.list.SetSize(c.width, max(c.height-searchFieldRows-footerRows, 1))
	c.input.Width = fieldWidth(c.width, searchPrompt)

	c.pm.SetSize(
		max(c.width-ui.PopupFrameWidth, 1),
		max(c.height-ui.PopupFrameHeight, 1),
	)
}

// statusView is the error line, or one blank row when there is no error. It
// always takes the same space so the popup does not jump when a failure shows.
func (c *CollectionsPopup) statusView() string {
	msg := ""
	if c.err != nil {
		msg = "Error: " + c.err.Error()
	}

	return lipgloss.NewStyle().
		MaxWidth(c.width).
		Height(1).
		Foreground(lipgloss.Color(theme.Error)).
		Render(msg)
}

// fit crops content to the area inside the frame. Wrapping an overlong line
// would tear the rounded border apart instead of just clipping it.
func (c *CollectionsPopup) fit(s string) string {
	return lipgloss.NewStyle().MaxWidth(c.width).MaxHeight(c.height).Render(s)
}

func (c *CollectionsPopup) View() string {
	help := c.keybind.ShortHelp()
	status := c.statusView()

	if _, m := c.pm.Active(); m != nil {
		sub := lipgloss.NewStyle().
			MaxWidth(c.width).
			MaxHeight(c.height).
			Render(m.View())

		content := c.fit(lipgloss.JoinVertical(lipgloss.Top, sub, status, help))

		return lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Width(c.width).
			Height(c.height - 2).
			Render(content)
	}

	if c.state == stateSearch {
		return ui.PopupBox(c.fit(lipgloss.JoinVertical(lipgloss.Top, c.searchView(), status, help)))
	}

	// The status line sits above the help so the padding PopupBox draws below
	// it stays blank whether or not an error is showing.
	return ui.PopupBox(c.fit(lipgloss.JoinVertical(lipgloss.Top, c.list.view(), status, help)))
}
