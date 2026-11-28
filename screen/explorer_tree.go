package screen

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/mhbidhan/lazycurl/theme"
	"github.com/mhbidhan/lazycurl/ui"
)

// treeRowLines is how many viewport lines one visible node occupies. Unlike the
// collections list, a tree row needs no blank line under it: indentation and
// markers already separate rows visually.
const treeRowLines = 1

// treeRow is one node in the flattened, currently visible tree.
type treeRow struct {
	Node        store.Node
	Depth       int
	Expanded    bool
	HasChildren bool
}

// explorerTree owns the tree's nodes, expansion state, filter and cursor. It is
// refreshed from Update; View only renders the viewport content already there.
type explorerTree struct {
	vp    viewport.Model
	title string

	nodes    []store.Node
	query    string
	cursor   int
	rows     []treeRow
	expanded map[uint]bool
}

func newExplorerTree(w, h int) *explorerTree {
	vp := viewport.New(max(w, 1), max(h, 1))
	vp.MouseWheelEnabled = true
	return &explorerTree{
		vp:       vp,
		expanded: make(map[uint]bool),
	}
}

// SetSize gives the viewport the area left for rows after the screen reserves
// its title, filter and footer.
func (t *explorerTree) SetSize(w, h int) {
	t.vp.Width = max(w, 1)
	t.vp.Height = max(h, 1)
}

func (t *explorerTree) setTitle(title string) {
	t.title = title
}

// load replaces the nodes while keeping expansion state where node IDs still
// exist. New directories start expanded so opening a collection shows its tree.
func (t *explorerTree) load(nodes []store.Node) {
	t.nodes = nodes
	for _, n := range nodes {
		if n.Type == store.NodeTypeDirectory {
			if _, ok := t.expanded[n.ID]; !ok {
				t.expanded[n.ID] = true
			}
		}
	}
	t.rebuild()
}

func (t *explorerTree) setQuery(q string) {
	t.query = q
	t.cursor = 0
	t.rebuild()
}

// rebuild flattens the tree again and keeps the cursor on screen. Call it from
// Update after any change to nodes, expansion or the filter.
func (t *explorerTree) rebuild() {
	t.rows = t.flatten()
	t.cursor = min(max(t.cursor, 0), max(len(t.rows)-1, 0))
	t.refresh()
	t.scrollToCursor()
}

func (t *explorerTree) refresh() {
	t.vp.SetContent(t.renderRows())
}

// flatten turns the parent/position model into the visible rows. When a filter
// is active, matching nodes bring their ancestors along and the matching
// branches are shown expanded.
func (t *explorerTree) flatten() []treeRow {
	// Collect all IDs before resolving parents: nodes are sorted globally by
	// position, so a child can appear before its parent in t.nodes. Treating an
	// unseen parent as root here renders correctly-parented nodes at the
	// collection root and makes them look like they moved.
	ids := make(map[uint]bool, len(t.nodes))
	for _, n := range t.nodes {
		ids[n.ID] = true
	}

	parents := make(map[uint]uint, len(t.nodes))
	children := make(map[uint][]store.Node)
	for _, n := range t.nodes {
		key := uint(0)
		if n.ParentID != nil && ids[*n.ParentID] {
			key = *n.ParentID
			parents[n.ID] = key
		}
		children[key] = append(children[key], n)
	}

	visible := make(map[uint]bool)
	if t.query != "" {
		for _, n := range t.nodes {
			if !chrome.FuzzyMatch(n.Name, t.query) {
				continue
			}
			current := n.ID
			for current != 0 {
				visible[current] = true
				current = parents[current]
			}
		}
	}

	var rows []treeRow
	var walk func(parent uint, depth int)
	walk = func(parent uint, depth int) {
		for _, n := range children[parent] {
			if t.query != "" && !visible[n.ID] {
				continue
			}
			hasChildren := len(children[n.ID]) > 0
			expanded := n.Type == store.NodeTypeDirectory && (t.query != "" || t.expanded[n.ID])
			rows = append(rows, treeRow{
				Node:        n,
				Depth:       depth,
				Expanded:    expanded,
				HasChildren: hasChildren,
			})
			if expanded {
				walk(n.ID, depth+1)
			}
		}
	}
	walk(0, 0)
	return rows
}

func (t *explorerTree) renderRows() string {
	selected := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Primary)).
		Foreground(lipgloss.Color(theme.OnPrimary))

	lines := make([]string, 0, len(t.rows))
	for i, row := range t.rows {
		marker := " "
		if row.Node.Type == store.NodeTypeDirectory {
			if row.Expanded {
				marker = "▾"
			} else {
				marker = "▸"
			}
		} else if row.Node.Method != "" {
			marker = row.Node.Method
		} else {
			marker = "REQ"
		}

		line := fmt.Sprintf("%s%s %s", strings.Repeat("  ", row.Depth), marker, row.Node.Name)
		if i == t.cursor {
			line = selected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (t *explorerTree) current() (treeRow, bool) {
	if len(t.rows) == 0 {
		return treeRow{}, false
	}
	return t.rows[t.cursor], true
}

func (t *explorerTree) currentID() uint {
	row, ok := t.current()
	if !ok {
		return 0
	}
	return row.Node.ID
}

func (t *explorerTree) selectID(id uint) {
	for i, row := range t.rows {
		if row.Node.ID == id {
			t.cursor = i
			t.refresh()
			t.scrollToCursor()
			return
		}
	}
}

// revealID expands every ancestor of id before selecting it. Structural edits
// use this instead of selectID so a node created or pasted into a collapsed
// directory does not silently disappear from the view.
func (t *explorerTree) revealID(id uint) {
	parents := make(map[uint]*uint, len(t.nodes))
	for _, n := range t.nodes {
		parents[n.ID] = n.ParentID
	}

	seen := make(map[uint]bool)
	for parent := parents[id]; parent != nil && !seen[*parent]; parent = parents[*parent] {
		seen[*parent] = true
		t.expanded[*parent] = true
	}

	t.rebuild()
	t.selectID(id)
}

func (t *explorerTree) move(delta int) {
	if len(t.rows) == 0 {
		return
	}
	t.cursor = (t.cursor + delta + len(t.rows)) % len(t.rows)
	t.refresh()
	t.scrollToCursor()
}

func (t *explorerTree) top() {
	t.cursor = 0
	t.refresh()
	t.scrollToCursor()
}

func (t *explorerTree) bottom() {
	t.cursor = max(len(t.rows)-1, 0)
	t.refresh()
	t.scrollToCursor()
}

func (t *explorerTree) toggle() {
	row, ok := t.current()
	if !ok || row.Node.Type != store.NodeTypeDirectory {
		return
	}
	if t.expanded[row.Node.ID] {
		delete(t.expanded, row.Node.ID)
	} else {
		t.expanded[row.Node.ID] = true
	}
	t.rebuild()
}

func (t *explorerTree) setExpanded(expanded bool) {
	row, ok := t.current()
	if !ok || row.Node.Type != store.NodeTypeDirectory {
		return
	}
	t.expanded[row.Node.ID] = expanded
	t.rebuild()
}

func (t *explorerTree) scrollToCursor() {
	y := t.cursor * treeRowLines
	if y < t.vp.YOffset {
		t.vp.SetYOffset(y)
		return
	}
	if y >= t.vp.YOffset+t.vp.Height {
		t.vp.SetYOffset(y - t.vp.Height + treeRowLines)
	}
}

// view draws the title outside the viewport so scrolling cannot hide it.
func (t *explorerTree) view() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Heading)).
		Align(lipgloss.Center, lipgloss.Center).
		Render(t.title)
	return lipgloss.JoinVertical(lipgloss.Top, title, ui.GapVertical(explorerHeaderRows-1), t.vp.View())
}
