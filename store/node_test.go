package store

import (
	"strings"
	"testing"
)

func TestNewNodeDefaults(t *testing.T) {
	dir := NewNode("folder", NodeTypeDirectory, 1)
	if dir.Type != NodeTypeDirectory || dir.Method != "" || dir.ParentID != nil {
		t.Fatalf("directory defaults = %+v", dir)
	}
	req := NewNode("call", NodeTypeRequest, 1)
	if req.Type != NodeTypeRequest || req.Method != "GET" {
		t.Fatalf("request defaults = %+v", req)
	}
	if req.TableName() != "nodes" {
		t.Fatalf("table = %q", req.TableName())
	}
}

func TestNodeCreateAndSiblingOrder(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}

	one, err := CreateNode(db, coll.ID, nil, 0, "one", NodeTypeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	two, err := CreateNode(db, coll.ID, nil, 0, "two", NodeTypeRequest)
	if err != nil {
		t.Fatal(err)
	}
	middle, err := CreateNode(db, coll.ID, nil, one.ID, "middle", NodeTypeRequest)
	if err != nil {
		t.Fatal(err)
	}
	ordered, _ := GetNodes(db, coll.ID)
	if ordered[1].ID != middle.ID || ordered[1].Position != 1 || ordered[2].ID != two.ID || ordered[2].Position != 2 {
		t.Fatalf("insert after did not shift siblings: %+v", ordered)
	}

	child, err := CreateNode(db, coll.ID, &one.ID, 0, "child", NodeTypeRequest)
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentID == nil || *child.ParentID != one.ID {
		t.Fatalf("child parent = %v, want %d", child.ParentID, one.ID)
	}

	nodes, err := GetNodes(db, coll.ID)
	if err != nil {
		t.Fatal(err)
	}
	var roots, children []string
	for _, n := range nodes {
		if n.ParentID == nil {
			roots = append(roots, n.Name)
		} else {
			children = append(children, n.Name)
		}
	}
	if strings.Join(roots, ",") != "one,middle,two" || strings.Join(children, ",") != "child" {
		t.Fatalf("roots = %v, children = %v", roots, children)
	}
}

func TestNodeCreateValidation(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	req, err := CreateNode(db, coll.ID, nil, 0, "request", NodeTypeRequest)
	if err != nil {
		t.Fatal(err)
	}
	otherColl := NewCollection("other")
	if err := SaveCollection(db, otherColl); err != nil {
		t.Fatal(err)
	}
	other, err := CreateNode(db, otherColl.ID, nil, 0, "other", NodeTypeDirectory)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		run  func() error
		want string
	}{
		{"unknown type", func() error { _, err := CreateNode(db, coll.ID, nil, 0, "x", "bad"); return err }, "unknown node type"},
		{"empty name", func() error { _, err := CreateNode(db, coll.ID, nil, 0, "", NodeTypeRequest); return err }, "cannot be empty"},
		{"request parent", func() error { _, err := CreateNode(db, coll.ID, &req.ID, 0, "x", NodeTypeRequest); return err }, "only directories"},
		{"foreign parent", func() error { _, err := CreateNode(db, coll.ID, &other.ID, 0, "x", NodeTypeRequest); return err }, "different collection"},
		{"missing parent", func() error {
			missing := uint(999)
			_, err := CreateNode(db, coll.ID, &missing, 0, "x", NodeTypeRequest)
			return err
		}, "not found"},
		{"foreign sibling", func() error { _, err := CreateNode(db, coll.ID, nil, other.ID, "x", NodeTypeRequest); return err }, "not a sibling"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestNodeRenameAndReorder(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	one, _ := CreateNode(db, coll.ID, nil, 0, "one", NodeTypeDirectory)
	two, _ := CreateNode(db, coll.ID, nil, 0, "two", NodeTypeRequest)
	three, _ := CreateNode(db, coll.ID, nil, 0, "three", NodeTypeRequest)

	if err := RenameNode(db, two.ID, "renamed"); err != nil {
		t.Fatal(err)
	}
	if err := RenameNode(db, two.ID, ""); err == nil {
		t.Fatal("empty rename succeeded")
	}
	if err := ReorderNode(db, three.ID, -1); err != nil {
		t.Fatal(err)
	}
	nodes, _ := GetNodes(db, coll.ID)
	if got := strings.Join(nodeNames(nodes), ","); got != "one,three,renamed" {
		t.Fatalf("after reorder = %q", got)
	}
	// Moving past either end is a no-op, not an error.
	if err := ReorderNode(db, one.ID, -1); err != nil {
		t.Fatal(err)
	}
	if err := ReorderNode(db, one.ID, 0); err == nil {
		t.Fatal("invalid reorder delta succeeded")
	}
	nodes, _ = GetNodes(db, coll.ID)
	if nodes[0].ID != one.ID {
		t.Fatal("top node moved when reordering past the top")
	}

	got, err := GetNode(db, one.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "one" {
		t.Fatalf("name = %q", got.Name)
	}
	if _, err := GetNode(db, 999); err == nil {
		t.Fatal("missing node did not fail")
	}
}

func TestNodeMoveAndCyclePrevention(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	alpha, _ := CreateNode(db, coll.ID, nil, 0, "alpha", NodeTypeDirectory)
	beta, _ := CreateNode(db, coll.ID, nil, 0, "beta", NodeTypeDirectory)
	alphaChild, _ := CreateNode(db, coll.ID, &alpha.ID, 0, "child", NodeTypeDirectory)
	alphaKeep, _ := CreateNode(db, coll.ID, &alpha.ID, 0, "keep", NodeTypeDirectory)
	betaChild, _ := CreateNode(db, coll.ID, &beta.ID, 0, "other", NodeTypeRequest)

	if err := MoveNode(db, alphaChild.ID, &beta.ID, betaChild.ID); err != nil {
		t.Fatal(err)
	}
	nodes, _ := GetNodes(db, coll.ID)
	positions := map[uint]int{}
	for _, n := range nodes {
		positions[n.ID] = n.Position
	}
	if positions[betaChild.ID] != 0 || positions[alphaChild.ID] != 1 {
		t.Fatalf("move positions = %+v", positions)
	}

	if err := MoveNode(db, alpha.ID, &alphaKeep.ID, 0); err == nil || !strings.Contains(err.Error(), "descendant") {
		t.Fatalf("cycle was not blocked: %v", err)
	}
	self := alpha.ID
	if err := MoveNode(db, alpha.ID, &self, 0); err == nil || !strings.Contains(err.Error(), "itself") {
		t.Fatalf("self move was not blocked: %v", err)
	}
	if err := MoveNode(db, alpha.ID, nil, alphaChild.ID); err == nil || !strings.Contains(err.Error(), "sibling") {
		t.Fatalf("bad sibling move was not blocked: %v", err)
	}
	if err := MoveNode(db, 999, nil, 0); err == nil {
		t.Fatal("missing node move succeeded")
	}
}

func TestNodeDeleteCascades(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	dir, _ := CreateNode(db, coll.ID, nil, 0, "dir", NodeTypeDirectory)
	child, _ := CreateNode(db, coll.ID, &dir.ID, 0, "child", NodeTypeDirectory)
	grandchild, _ := CreateNode(db, coll.ID, &child.ID, 0, "grandchild", NodeTypeRequest)
	root, _ := CreateNode(db, coll.ID, nil, 0, "root", NodeTypeDirectory)

	if err := DeleteNode(db, dir.ID); err != nil {
		t.Fatal(err)
	}
	nodes, err := GetNodes(db, coll.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].ID != root.ID || nodes[0].Position != 0 {
		t.Fatalf("nodes after cascade = %+v", nodes)
	}
	for _, id := range []uint{dir.ID, child.ID, grandchild.ID} {
		if _, err := GetNode(db, id); err == nil {
			t.Fatalf("node %d survived cascade", id)
		}
	}
	if err := DeleteNode(db, 999); err == nil {
		t.Fatal("deleting missing node succeeded")
	}
}

func TestNodeMoveKeepsExistingChildren(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	source, _ := CreateNode(db, coll.ID, nil, 0, "source", NodeTypeDirectory)
	target, _ := CreateNode(db, coll.ID, nil, 0, "target", NodeTypeDirectory)
	existing, _ := CreateNode(db, coll.ID, &target.ID, 0, "existing", NodeTypeRequest)
	second, _ := CreateNode(db, coll.ID, &target.ID, 0, "second", NodeTypeRequest)
	moving, _ := CreateNode(db, coll.ID, &source.ID, 0, "moving", NodeTypeRequest)

	if err := MoveNode(db, moving.ID, &target.ID, existing.ID); err != nil {
		t.Fatal(err)
	}

	nodes, _ := GetNodes(db, coll.ID)
	children := childrenOf(nodes, target.ID)
	if got := strings.Join(nodeNames(children), ","); got != "existing,moving,second" {
		t.Fatalf("target children = %q", got)
	}
	for i, child := range children {
		if child.Position != i {
			t.Fatalf("child %s position = %d, want %d", child.Name, child.Position, i)
		}
	}
	kept, _ := GetNode(db, existing.ID)
	if kept.ParentID == nil || *kept.ParentID != target.ID {
		t.Fatalf("existing child moved out: parent = %v", kept.ParentID)
	}
	alsoKept, _ := GetNode(db, second.ID)
	if alsoKept.ParentID == nil || *alsoKept.ParentID != target.ID {
		t.Fatalf("second child moved out: parent = %v", alsoKept.ParentID)
	}
}

func TestNodeMoveAfterSiblingKeepsDensePositions(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	dir, _ := CreateNode(db, coll.ID, nil, 0, "dir", NodeTypeDirectory)
	first, _ := CreateNode(db, coll.ID, &dir.ID, 0, "a", NodeTypeRequest)
	second, _ := CreateNode(db, coll.ID, &dir.ID, 0, "b", NodeTypeRequest)
	CreateNode(db, coll.ID, &dir.ID, 0, "c", NodeTypeRequest)
	CreateNode(db, coll.ID, &dir.ID, 0, "d", NodeTypeRequest)

	if err := MoveNode(db, first.ID, &dir.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	nodes, _ := GetNodes(db, coll.ID)
	children := childrenOf(nodes, dir.ID)
	if got := strings.Join(nodeNames(children), ","); got != "b,a,c,d" {
		t.Fatalf("children = %q", got)
	}
	for i, child := range children {
		if child.Position != i {
			t.Fatalf("child %s position = %d, want %d", child.Name, child.Position, i)
		}
	}
}

func TestCopyNode(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	sourceDir, _ := CreateNode(db, coll.ID, nil, 0, "source", NodeTypeDirectory)
	targetDir, _ := CreateNode(db, coll.ID, nil, 0, "target", NodeTypeDirectory)
	request, _ := CreateNode(db, coll.ID, &sourceDir.ID, 0, "call", NodeTypeRequest)
	if err := db.Model(&Node{}).Where("id = ?", request.ID).Updates(map[string]any{
		"method": "POST", "url": "https://example.com", "headers": `{"X":"y"}`,
	}).Error; err != nil {
		t.Fatal(err)
	}
	existing, _ := CreateNode(db, coll.ID, &targetDir.ID, 0, "existing", NodeTypeRequest)

	copied, err := CopyNode(db, request.ID, &targetDir.ID, existing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if copied.ID == request.ID || copied.ParentID == nil || *copied.ParentID != targetDir.ID {
		t.Fatalf("copy = %+v", copied)
	}
	if copied.Method != "POST" || copied.URL != "https://example.com" || copied.Headers != `{"X":"y"}` {
		t.Fatalf("copy lost request fields: %+v", copied)
	}
	original, _ := GetNode(db, request.ID)
	if original.ParentID == nil || *original.ParentID != sourceDir.ID {
		t.Fatalf("original moved: parent = %v", original.ParentID)
	}

	nodes, _ := GetNodes(db, coll.ID)
	children := childrenOf(nodes, targetDir.ID)
	if got := strings.Join(nodeNames(children), ","); got != "existing,call" {
		t.Fatalf("target children = %q", got)
	}
}

func TestCopyNodeSubtree(t *testing.T) {
	db := testDB(t)
	coll := NewCollection("work")
	if err := SaveCollection(db, coll); err != nil {
		t.Fatal(err)
	}
	source, _ := CreateNode(db, coll.ID, nil, 0, "folder", NodeTypeDirectory)
	child, _ := CreateNode(db, coll.ID, &source.ID, 0, "nested", NodeTypeDirectory)
	CreateNode(db, coll.ID, &child.ID, 0, "deep", NodeTypeRequest)
	target, _ := CreateNode(db, coll.ID, nil, 0, "target", NodeTypeDirectory)

	copied, err := CopyNode(db, source.ID, &target.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if copied.ID == source.ID {
		t.Fatal("copy reused the source ID")
	}

	nodes, _ := GetNodes(db, coll.ID)
	if len(nodes) != 7 {
		t.Fatalf("nodes after copy = %d, want 7", len(nodes))
	}
	copiedChildren := childrenOf(nodes, copied.ID)
	if len(copiedChildren) != 1 || copiedChildren[0].Name != "nested" || copiedChildren[0].ID == child.ID {
		t.Fatalf("copied children = %+v", copiedChildren)
	}
	deep := childrenOf(nodes, copiedChildren[0].ID)
	if len(deep) != 1 || deep[0].Name != "deep" {
		t.Fatalf("copied grandchildren = %+v", deep)
	}
	original, _ := GetNode(db, source.ID)
	if original.ParentID != nil {
		t.Fatalf("original parent = %v", original.ParentID)
	}
}

func childrenOf(nodes []Node, parentID uint) []Node {
	var children []Node
	for _, n := range nodes {
		if n.ParentID != nil && *n.ParentID == parentID {
			children = append(children, n)
		}
	}
	return children
}

func nodeNames(nodes []Node) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	return names
}
