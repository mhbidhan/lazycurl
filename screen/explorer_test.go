package screen

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/muesli/termenv"
)

func enterKey() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

func newExplorer(t *testing.T) (*ExplorerScreen, store.Collection) {
	t.Helper()
	db := testDB(t)
	coll := store.NewCollection("work")
	if err := store.SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	return NewExplorerScreen(db, *coll, 80, 24), *coll
}

func addNode(t *testing.T, s *ExplorerScreen, coll store.Collection, parent *uint, name string, typ store.NodeType) store.Node {
	t.Helper()
	node, err := store.CreateNode(s.db, coll.ID, parent, 0, name, typ)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.reload(); err != nil {
		t.Fatal(err)
	}
	return node
}

func namedNodes(nodes []store.Node) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	return names
}

func TestExplorerTreeNavigation(t *testing.T) {
	// Color is stripped in tests, so force an ANSI profile to observe the
	// selected-row highlight moving with j/k.
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI)
	defer lipgloss.SetColorProfile(profile)

	s, coll := newExplorer(t)
	api := addNode(t, s, coll, nil, "api", store.NodeTypeDirectory)
	addNode(t, s, coll, &api.ID, "list", store.NodeTypeRequest)
	addNode(t, s, coll, nil, "health", store.NodeTypeRequest)

	view := s.View()
	for _, want := range []string{"Collection: work", "api", "list", "health", "▾", "GET"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if lipgloss.Height(view) != 24 || lipgloss.Width(view) != 80 {
		t.Fatalf("view = %dx%d, want 80x24", lipgloss.Width(view), lipgloss.Height(view))
	}

	// j/k must move the highlight from a directory to a request and redraw it.
	focusBefore := s.tree.vp.View()
	dirFocus := s.tree.rows[s.tree.cursor].Node.Type
	s.Update(keyMsg("j"))
	if got := s.tree.currentID(); got != s.tree.rows[1].Node.ID {
		t.Fatalf("cursor did not move down: %d", got)
	}
	if reqFocus := s.tree.rows[s.tree.cursor].Node.Type; dirFocus == reqFocus {
		t.Fatalf("j did not change node type: %s", reqFocus)
	}
	focusAfterDown := s.tree.vp.View()
	if focusAfterDown == focusBefore {
		t.Fatal("j moved the cursor but did not redraw the focus")
	}
	s.Update(keyMsg("k"))
	if s.tree.cursor != 0 {
		t.Fatalf("cursor did not move up: %d", s.tree.cursor)
	}
	if got := s.tree.vp.View(); got == focusAfterDown || got != focusBefore {
		t.Fatal("k moved back but did not redraw the focus")
	}
	s.Update(keyMsg("G"))
	if s.tree.currentID() == 0 || s.tree.cursor != len(s.tree.rows)-1 {
		t.Fatalf("bottom left cursor at %d", s.tree.cursor)
	}
	s.Update(keyMsg("g"))
	if s.tree.cursor != 0 {
		t.Fatalf("top left cursor at %d", s.tree.cursor)
	}

	s.Update(keyMsg("enter"))
	if strings.Contains(s.View(), "list") {
		t.Fatalf("collapsed directory still shows its child:\n%s", s.View())
	}
	if !strings.Contains(s.View(), "▸") {
		t.Fatal("collapsed marker missing")
	}
	s.Update(keyMsg("enter"))
	if !strings.Contains(s.View(), "list") {
		t.Fatal("expanding did not restore the child")
	}

	s.Update(keyMsg("/"))
	if s.state != stateSearch {
		t.Fatalf("search state = %d", s.state)
	}
	s.Update(keyMsg("hea"))
	if len(s.tree.rows) != 1 || s.tree.rows[0].Node.Name != "health" {
		t.Fatalf("filtered rows = %+v", s.tree.rows)
	}
	s.Update(enterKey())
	if s.state != stateTree || s.tree.query != "hea" {
		t.Fatalf("submit left state = %d, query = %q", s.state, s.tree.query)
	}
	s.Update(keyMsg("/"))
	s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s.tree.query != "" {
		t.Fatalf("cancel left query = %q", s.tree.query)
	}
}

func TestExplorerHelpAndBack(t *testing.T) {
	s, _ := newExplorer(t)

	s.Update(keyMsg("?"))
	if !s.pm.IsOpen() || !strings.Contains(s.View(), "cursor down") {
		t.Fatalf("help popup not shown:\n%s", s.View())
	}
	s.Update(chrome.DismissPopup{})
	if s.pm.IsOpen() {
		t.Fatal("dismiss did not close help")
	}

	_, cmd := s.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("q did not return a command")
	}
	if _, ok := cmd().(BackToHomeMsg); !ok {
		t.Fatalf("q command returned %T, want BackToHomeMsg", cmd())
	}
}

func TestExplorerCreateRenameDelete(t *testing.T) {
	s, coll := newExplorer(t)

	// a creates a request from a plain name, including the command/event hop
	// the real Bubble Tea loop performs.
	s.Update(keyMsg("a"))
	if s.state != stateName || !s.pm.IsOpen() {
		t.Fatalf("add request state = %d, open = %t", s.state, s.pm.IsOpen())
	}
	s.Update(keyMsg("call"))
	_, cmd := s.Update(enterKey())
	if cmd == nil {
		t.Fatal("name submit produced no command")
	}
	s.Update(cmd().(chrome.NamePromptEvent))

	nodes, _ := store.GetNodes(s.db, coll.ID)
	if len(nodes) != 1 || nodes[0].Name != "call" || nodes[0].Type != store.NodeTypeRequest {
		t.Fatalf("created nodes = %+v", nodes)
	}
	if s.state != stateTree || s.pm.IsOpen() {
		t.Fatalf("create left state = %d, open = %t", s.state, s.pm.IsOpen())
	}
	if !strings.Contains(s.View(), "GET call") {
		t.Fatalf("request marker missing:\n%s", s.View())
	}

	// a with a trailing slash creates a directory and drops the slash.
	s.Update(keyMsg("a"))
	s.Update(chrome.NamePromptEvent{Value: "folder/"})
	nodes, _ = store.GetNodes(s.db, coll.ID)
	if len(nodes) != 2 || nodes[1].Name != "folder" || nodes[1].Type != store.NodeTypeDirectory {
		t.Fatalf("slash nodes = %+v", nodes)
	}

	// A always creates a directory, even without a slash.
	s.Update(keyMsg("A"))
	s.Update(chrome.NamePromptEvent{Value: "sibling"})
	nodes, _ = store.GetNodes(s.db, coll.ID)
	if len(nodes) != 3 || nodes[2].Name != "sibling" || nodes[2].Type != store.NodeTypeDirectory {
		t.Fatalf("sibling nodes = %+v", nodes)
	}

	// A bare slash leaves no usable name.
	s.Update(keyMsg("a"))
	s.Update(chrome.NamePromptEvent{Value: "/"})
	if s.message != "Name cannot be empty" {
		t.Fatalf("empty name message = %q", s.message)
	}
	nodes, _ = store.GetNodes(s.db, coll.ID)
	if len(nodes) != 3 {
		t.Fatalf("empty name created %d nodes", len(nodes))
	}

	selected := s.tree.currentID()
	s.Update(keyMsg("r"))
	s.Update(chrome.NamePromptEvent{Value: "renamed"})
	nodes, _ = store.GetNodes(s.db, coll.ID)
	found := false
	for _, n := range nodes {
		if n.ID == selected {
			found = n.Name == "renamed"
		}
	}
	if !found {
		t.Fatalf("rename did not stick: %+v", nodes)
	}

	s.Update(keyMsg("d"))
	if s.state != stateDelete || !s.pm.IsOpen() {
		t.Fatalf("delete state = %d, open = %t", s.state, s.pm.IsOpen())
	}
	s.Update(chrome.ConfirmEvent{Confirmed: true})
	nodes, _ = store.GetNodes(s.db, coll.ID)
	if len(nodes) != 2 {
		t.Fatalf("nodes after delete = %+v", nodes)
	}
}

func TestExplorerMovePasteAndReorder(t *testing.T) {
	s, coll := newExplorer(t)
	alpha := addNode(t, s, coll, nil, "alpha", store.NodeTypeDirectory)
	child := addNode(t, s, coll, &alpha.ID, "child", store.NodeTypeRequest)
	beta := addNode(t, s, coll, nil, "beta", store.NodeTypeDirectory)

	// Rows are alpha, child, beta. The removed m binding must be inert.
	s.Update(keyMsg("j"))
	s.Update(keyMsg("m"))
	if s.marked != 0 || s.markMode != markNone {
		t.Fatalf("m marked a node: marked = %d, mode = %d", s.marked, s.markMode)
	}

	// x marks for a move; p puts child into beta.
	s.Update(keyMsg("x"))
	if s.marked == 0 || s.markMode != markMove || !strings.Contains(s.View(), "Marked") {
		t.Fatalf("mark did not stick: marked = %d, mode = %d", s.marked, s.markMode)
	}
	s.Update(keyMsg("j"))
	s.Update(keyMsg("p"))

	childNode, err := store.GetNode(s.db, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if childNode.Name != "child" || childNode.ParentID == nil || *childNode.ParentID != beta.ID {
		t.Fatalf("pasted node = %+v", childNode)
	}
	if s.marked != 0 || s.markMode != markNone {
		t.Fatalf("paste did not clear the mark: marked = %d, mode = %d", s.marked, s.markMode)
	}

	third := addNode(t, s, coll, nil, "third", store.NodeTypeRequest)
	s.Update(keyMsg("G"))
	s.Update(keyMsg("K"))
	nodes, _ := store.GetNodes(s.db, coll.ID)
	var roots []string
	for _, n := range nodes {
		if n.ParentID == nil {
			roots = append(roots, n.Name)
		}
	}
	if strings.Join(roots, ",") != "alpha,third,beta" {
		t.Fatalf("roots after reorder = %v", roots)
	}
	if s.tree.currentID() != third.ID {
		t.Fatal("cursor did not follow the reordered node")
	}
}

func TestExplorerPasteRevealsNodeInCollapsedDir(t *testing.T) {
	s, coll := newExplorer(t)
	dir := addNode(t, s, coll, nil, "dir", store.NodeTypeDirectory)
	addNode(t, s, coll, &dir.ID, "existing", store.NodeTypeRequest)
	req := addNode(t, s, coll, nil, "req", store.NodeTypeRequest)

	s.Update(keyMsg("g"))
	s.Update(keyMsg("h"))
	if len(s.tree.rows) != 2 {
		t.Fatalf("collapsed rows = %+v", s.tree.rows)
	}
	s.Update(keyMsg("j"))
	s.Update(keyMsg("x"))
	s.Update(keyMsg("k"))
	s.Update(keyMsg("p"))

	moved, err := store.GetNode(s.db, req.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.ParentID == nil || *moved.ParentID != dir.ID {
		t.Fatalf("pasted parent = %v, want %d", moved.ParentID, dir.ID)
	}
	if !s.tree.expanded[dir.ID] {
		t.Fatal("paste left the target directory collapsed")
	}
	visible := false
	for _, row := range s.tree.rows {
		if row.Node.ID == req.ID && row.Depth == 1 {
			visible = true
		}
	}
	if !visible || s.tree.currentID() != req.ID {
		t.Fatalf("moved node not revealed: rows = %+v, cursor = %d", s.tree.rows, s.tree.cursor)
	}
}

func TestExplorerPasteKeepsMarkOnFailure(t *testing.T) {
	s, coll := newExplorer(t)
	dir := addNode(t, s, coll, nil, "dir", store.NodeTypeDirectory)
	req := addNode(t, s, coll, nil, "req", store.NodeTypeRequest)

	s.Update(keyMsg("j")) // req
	s.Update(keyMsg("x"))
	_, cmd := s.Update(keyMsg("p")) // pasting onto itself is rejected
	if cmd == nil {
		t.Fatal("rejected paste returned no error command")
	}
	if s.marked != req.ID || s.markMode != markMove {
		t.Fatalf("failed paste cleared the mark: marked = %d, mode = %d", s.marked, s.markMode)
	}
	s.Update(cmd().(chrome.OperationError))

	s.Update(keyMsg("k")) // dir
	s.Update(keyMsg("p"))
	moved, err := store.GetNode(s.db, req.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.ParentID == nil || *moved.ParentID != dir.ID {
		t.Fatalf("retry parent = %v, want %d", moved.ParentID, dir.ID)
	}
	if s.marked != 0 || s.markMode != markNone {
		t.Fatalf("successful paste kept the mark: marked = %d, mode = %d", s.marked, s.markMode)
	}
}

func TestExplorerCreateRevealsNodeInCollapsedDir(t *testing.T) {
	s, coll := newExplorer(t)
	dir := addNode(t, s, coll, nil, "dir", store.NodeTypeDirectory)

	s.Update(keyMsg("g"))
	s.Update(keyMsg("h"))
	s.Update(keyMsg("a"))
	s.Update(chrome.NamePromptEvent{Value: "new"})

	nodes, _ := store.GetNodes(s.db, coll.ID)
	if len(nodes) != 2 || nodes[1].Name != "new" || nodes[1].ParentID == nil || *nodes[1].ParentID != dir.ID {
		t.Fatalf("created nodes = %+v", nodes)
	}
	if !s.tree.expanded[dir.ID] {
		t.Fatal("create left the target directory collapsed")
	}
	if s.tree.currentID() != nodes[1].ID {
		t.Fatalf("cursor = %d, want created node %d", s.tree.currentID(), nodes[1].ID)
	}
}

func TestExplorerCopyPaste(t *testing.T) {
	s, coll := newExplorer(t)
	source := addNode(t, s, coll, nil, "source", store.NodeTypeDirectory)
	req := addNode(t, s, coll, &source.ID, "call", store.NodeTypeRequest)
	if err := s.db.Model(&store.Node{}).Where("id = ?", req.ID).Updates(map[string]any{
		"method": "POST", "url": "https://example.com",
	}).Error; err != nil {
		t.Fatal(err)
	}
	target := addNode(t, s, coll, nil, "target", store.NodeTypeDirectory)
	addNode(t, s, coll, &target.ID, "existing", store.NodeTypeRequest)

	s.Update(keyMsg("g")) // source
	s.Update(keyMsg("j")) // call
	s.Update(keyMsg("c"))
	if s.marked != req.ID || s.markMode != markCopy || !strings.Contains(s.View(), "copy") {
		t.Fatalf("copy mark did not stick: marked = %d, mode = %d", s.marked, s.markMode)
	}
	s.Update(keyMsg("j")) // target
	s.Update(keyMsg("p"))

	nodes, _ := store.GetNodes(s.db, coll.ID)
	var copies []store.Node
	for _, n := range nodes {
		if n.Name == "call" {
			copies = append(copies, n)
		}
	}
	if len(copies) != 2 {
		t.Fatalf("nodes named call = %+v", copies)
	}
	var clone store.Node
	for _, n := range copies {
		if n.ID != req.ID {
			clone = n
		}
	}
	if clone.ID == 0 || clone.ParentID == nil || *clone.ParentID != target.ID {
		t.Fatalf("copy = %+v", clone)
	}
	if clone.Method != "POST" || clone.URL != "https://example.com" {
		t.Fatalf("copy lost request fields: %+v", clone)
	}
	original, _ := store.GetNode(s.db, req.ID)
	if original.ParentID == nil || *original.ParentID != source.ID {
		t.Fatalf("original moved: parent = %v", original.ParentID)
	}
	if s.marked != 0 || s.markMode != markNone {
		t.Fatalf("paste did not clear the mark: marked = %d, mode = %d", s.marked, s.markMode)
	}
	if s.tree.currentID() != clone.ID {
		t.Fatalf("cursor = %d, want copy %d", s.tree.currentID(), clone.ID)
	}
}

func TestExplorerCopyDirectorySubtree(t *testing.T) {
	s, coll := newExplorer(t)
	source := addNode(t, s, coll, nil, "folder", store.NodeTypeDirectory)
	child := addNode(t, s, coll, &source.ID, "nested", store.NodeTypeDirectory)
	req := addNode(t, s, coll, &child.ID, "deep", store.NodeTypeRequest)
	target := addNode(t, s, coll, nil, "target", store.NodeTypeDirectory)

	s.Update(keyMsg("g")) // folder
	s.Update(keyMsg("c"))
	s.Update(keyMsg("j")) // nested
	s.Update(keyMsg("j")) // deep
	s.Update(keyMsg("j")) // target
	s.Update(keyMsg("p"))

	nodes, _ := store.GetNodes(s.db, coll.ID)
	if len(nodes) != 7 {
		t.Fatalf("nodes after copy = %d, want 7", len(nodes))
	}
	var clone store.Node
	for _, n := range nodes {
		if n.Name == "folder" && n.ID != source.ID {
			clone = n
		}
	}
	if clone.ID == 0 || clone.ParentID == nil || *clone.ParentID != target.ID {
		t.Fatalf("directory copy = %+v", clone)
	}
	var copiedChild store.Node
	for _, n := range nodes {
		if n.Name == "nested" && n.ParentID != nil && *n.ParentID == clone.ID {
			copiedChild = n
		}
	}
	if copiedChild.ID == 0 || copiedChild.ID == child.ID {
		t.Fatalf("nested copy = %+v", copiedChild)
	}
	var copiedReq store.Node
	for _, n := range nodes {
		if n.Name == "deep" && n.ParentID != nil && *n.ParentID == copiedChild.ID {
			copiedReq = n
		}
	}
	if copiedReq.ID == 0 || copiedReq.ID == req.ID {
		t.Fatalf("deep copy = %+v", copiedReq)
	}
	if s.tree.currentID() != clone.ID {
		t.Fatalf("cursor = %d, want directory copy %d", s.tree.currentID(), clone.ID)
	}
}

func TestExplorerTreeKeepsChildUnderLateParent(t *testing.T) {
	s, coll := newExplorer(t)
	dir := addNode(t, s, coll, nil, "dir", store.NodeTypeDirectory)
	child := addNode(t, s, coll, &dir.ID, "child", store.NodeTypeRequest)
	rootReq := addNode(t, s, coll, nil, "root", store.NodeTypeRequest)

	// Force the global position order to child, root, dir. The child is sorted
	// before its parent, which must not make it a root-level row.
	for id, pos := range map[uint]int{dir.ID: 2, rootReq.ID: 1} {
		if err := s.db.Model(&store.Node{}).Where("id = ?", id).Update("position", pos).Error; err != nil {
			t.Fatal(err)
		}
	}
	s.reload()

	if len(s.tree.rows) != 3 {
		t.Fatalf("rows = %+v", s.tree.rows)
	}
	if s.tree.rows[0].Node.ID != rootReq.ID || s.tree.rows[0].Depth != 0 {
		t.Fatalf("first row = %+v", s.tree.rows[0])
	}
	if s.tree.rows[1].Node.ID != dir.ID || s.tree.rows[1].Depth != 0 {
		t.Fatalf("second row = %+v", s.tree.rows[1])
	}
	if s.tree.rows[2].Node.ID != child.ID || s.tree.rows[2].Depth != 1 {
		t.Fatalf("child row = %+v", s.tree.rows[2])
	}
}

func TestExplorerOperationErrorStatus(t *testing.T) {
	s, _ := newExplorer(t)
	sqlDB, err := s.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}

	s.Update(keyMsg("a"))
	_, cmd := s.Update(chrome.NamePromptEvent{Value: "folder"})
	if cmd == nil {
		t.Fatal("failed create returned no error command")
	}
	msg, ok := cmd().(chrome.OperationError)
	if !ok {
		t.Fatalf("error command returned %T", cmd())
	}
	s.Update(msg)
	if s.err == nil || !strings.Contains(s.View(), "Error:") {
		t.Fatalf("status line missing:\n%s", s.View())
	}
}

func TestRootSwitchesScreens(t *testing.T) {
	db := testDB(t)
	coll := store.NewCollection("work")
	if err := store.SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	root := NewRoot(db)
	root.SetSize(80, 24)

	root.Update(chrome.OpenCollectionMsg{Collection: *coll})
	if root.explorer == nil || root.current != root.explorer {
		t.Fatal("open collection did not switch to the explorer")
	}
	if !strings.Contains(root.View(), "Collection: work") {
		t.Fatalf("explorer view = %q", root.View())
	}

	// Resizing while in the explorer must also reach the inactive home screen.
	root.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
	if root.home.width != 70 || root.home.height != 20 {
		t.Fatalf("home size = %dx%d", root.home.width, root.home.height)
	}
	if lipgloss.Height(root.View()) != 20 {
		t.Fatalf("resized explorer height = %d", lipgloss.Height(root.View()))
	}

	root.Update(BackToHomeMsg{})
	if root.current != root.home {
		t.Fatal("back did not restore the home screen")
	}
	if !strings.Contains(root.View(), "? keybinds") {
		t.Fatalf("home view = %q", root.View())
	}
}
