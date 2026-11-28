package chrome

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/mhbidhan/lazycurl/theme"
	"github.com/mhbidhan/lazycurl/ui"
)

const (
	// listHeaderLines is the room the list keeps above its rows for the title
	// and the gap below it. The title is drawn outside the viewport rather than
	// inside its content: as content it scrolls away the moment the list moves,
	// which loses the one line that says what the list is.
	listHeaderLines = 2

	// listRowLines is how many lines each collection takes: the row, plus the
	// blank line under it. renderRows joins rows with this and scrollToCursor
	// counts with it, so the highlight and the scroll position cannot drift
	// apart.
	listRowLines = 2

	// listVisibleRows is how many collections the viewport shows at once. The
	// viewport works in lines, so it needs one listRowLines block per row.
	listVisibleRows  = 4
	listContentLines = listVisibleRows * listRowLines
	listMinHeight    = listHeaderLines + listContentLines
)

// collectionList is the scrollable list behind the collections popup. It owns
// the rows, the filter and the cursor, and is the only place that knows how the
// list is drawn and how far it has scrolled.
type collectionList struct {
	vp          viewport.Model
	collections []store.Collection
	filtered    []store.Collection
	query       string
	cursor      int
	// gap is whether there is room for the blank line under the title. It is
	// decided by SetSize, not by view, so the height a list reports does not
	// depend on what is being drawn.
	gap bool
}

func newCollectionList(w, h int) *collectionList {
	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true

	l := &collectionList{
		vp:       vp,
		filtered: []store.Collection{},
	}
	// Through SetSize, so the reserved title rows and the gap are decided the
	// same way here as on every later resize.
	l.SetSize(max(w, 40), max(h, listMinHeight))

	return l
}

// setRows replaces the contents, re-applies the filter and resets the cursor.
func (l *collectionList) setRows(rows []store.Collection) {
	l.collections = rows
	l.applyFilter()
	l.cursor = 0
}

// setQuery filters the rows and pulls the cursor back into range.
func (l *collectionList) setQuery(q string) {
	l.query = q
	l.applyFilter()

	if l.cursor >= len(l.filtered) {
		l.cursor = max(len(l.filtered)-1, 0)
	}
}

// clearQuery drops the filter and the cursor back to the top.
func (l *collectionList) clearQuery() {
	l.query = ""
	l.applyFilter()
	l.cursor = 0
}

func (l *collectionList) applyFilter() {
	if l.query == "" {
		l.filtered = make([]store.Collection, len(l.collections))
		copy(l.filtered, l.collections)
		return
	}

	res := make([]store.Collection, 0, len(l.collections))
	for _, coll := range l.collections {
		if FuzzyMatch(coll.Name, l.query) {
			res = append(res, coll)
		}
	}
	l.filtered = res
}

// FuzzyMatch reports whether pattern appears in s as a subsequence,
// case-insensitively. Other screens reuse it for their name filters.
func FuzzyMatch(s, pattern string) bool {
	if pattern == "" {
		return true
	}
	sRunes := []rune(strings.ToLower(s))
	pRunes := []rune(strings.ToLower(pattern))
	j := 0
	for _, r := range sRunes {
		if j < len(pRunes) && r == pRunes[j] {
			j++
		}
	}
	return j == len(pRunes)
}

func (l *collectionList) empty() bool {
	return len(l.filtered) == 0
}

// current is the row under the cursor.
func (l *collectionList) current() (store.Collection, bool) {
	if l.empty() {
		return store.Collection{}, false
	}
	return l.filtered[l.cursor], true
}

func (l *collectionList) moveDown() {
	if l.empty() {
		return
	}
	l.cursor = (l.cursor + 1) % len(l.filtered)
	l.syncView()
}

func (l *collectionList) moveUp() {
	if l.empty() {
		return
	}
	l.cursor = (l.cursor - 1 + len(l.filtered)) % len(l.filtered)
	l.syncView()
}

// rowY is the line row i starts on within the viewport content. The rows start
// at line zero because the title lives outside the viewport now.
func (l *collectionList) rowY(i int) int {
	return i * listRowLines
}

// scrollToCursor keeps the selected row visible. Without it the highlight walks
// off screen in lists longer than the viewport.
func (l *collectionList) scrollToCursor() {
	y := l.rowY(l.cursor)
	if y < l.vp.YOffset {
		l.vp.SetYOffset(y)
		return
	}
	if y >= l.vp.YOffset+l.vp.Height {
		l.vp.SetYOffset(y - l.vp.Height + listRowLines)
	}
}

// refresh pushes the current rows into the viewport. Only call it from Update:
// View runs on every render tick and must not mutate state.
func (l *collectionList) refresh() {
	l.vp.SetContent(l.renderRows())
}

// syncView is what almost every state change ends with: redraw, then make sure
// the cursor is still on screen.
func (l *collectionList) syncView() {
	l.refresh()
	l.scrollToCursor()
}

// header is the title line. It is drawn by view rather than fed to the viewport
// so that scrolling the rows cannot scroll it away.
func (l *collectionList) header() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Heading)).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Collections")
}

// renderRows is the viewport content: the rows and nothing else.
func (l *collectionList) renderRows() string {
	if len(l.filtered) == 0 {
		return lipgloss.NewStyle().Align(lipgloss.Center).Render("No collections found")
	}

	lines := make([]string, len(l.filtered))

	normalBg := lipgloss.
		NewStyle().
		Padding(0, 1)
	selectedBg := lipgloss.
		NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color(theme.Primary)).
		Foreground(lipgloss.Color(theme.OnPrimary))
	for i, coll := range l.filtered {
		if i == l.cursor {
			lines[i] = selectedBg.Render(coll.Name)
		} else {
			lines[i] = normalBg.Render(coll.Name)
		}

		if i == len(l.filtered)-1 {
			lines[i] += "\n"
		}
	}

	// The gap between rows comes from the same constant scrollToCursor counts
	// lines with.
	return lipgloss.NewStyle().Align(lipgloss.Left).
		Render(strings.Join(lines, strings.Repeat("\n", listRowLines)))
}

// update passes anything the list itself does not handle to the viewport, which
// is what makes the mouse wheel work.
func (l *collectionList) update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	l.vp, cmd = l.vp.Update(msg)
	return cmd
}

// view is the whole list: the title, which never moves, then the scrolled rows.
// The gap under the title is skipped when the list has no room for it, so the
// list never grows taller than the height it was given.
func (l *collectionList) view() string {
	parts := []string{l.header()}
	if l.gap {
		parts = append(parts, ui.GapVertical(1))
	}

	return lipgloss.JoinVertical(lipgloss.Top, append(parts, l.vp.View())...)
}

// SetSize takes the height of the whole list, title included, and gives the
// viewport what is left. The title sits outside the viewport, so this is where
// its lines get reserved.
func (l *collectionList) SetSize(w, h int) {
	l.vp.Width = w
	// One row is the floor: the viewport renders a row even at height zero, so
	// there is no smaller a list can be.
	l.vp.Height = max(h-listHeaderLines, 1)
	l.gap = h >= l.vp.Height+listHeaderLines
}
