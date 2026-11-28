package store

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// NodeType tells a Node apart from a directory or a request. Requests are leaf
// nodes; only directories may own children.
type NodeType string

const (
	NodeTypeDirectory NodeType = "directory"
	NodeTypeRequest   NodeType = "request"
)

// Node is one entry in a collection's tree. Request-only fields stay empty on
// directories, but live on the same row so a node never loses its type-specific
// data when it moves around the tree.
type Node struct {
	gorm.Model

	Type NodeType
	Name string

	CollectionID uint
	ParentID     *uint
	Position     int

	// Request-only fields. Empty/ignored for directories.
	Method     string
	URL        string
	Headers    string // JSON-encoded map[string]string
	Query      string // JSON-encoded map[string]string
	Body       string
	BodyType   string
	AuthType   string
	AuthConfig string // JSON-encoded auth configuration

	Children []Node `gorm:"foreignKey:ParentID"`
}

// NewNode builds a node ready to save. New requests start with GET so a method
// label is always available to the tree view.
func NewNode(name string, typ NodeType, collectionID uint) *Node {
	n := &Node{
		Type:         typ,
		Name:         name,
		CollectionID: collectionID,
	}
	if typ == NodeTypeRequest {
		n.Method = "GET"
	}
	return n
}

// TableName names the table GORM maps nodes to.
func (*Node) TableName() string {
	return "nodes"
}

// GetNode returns one live node by id.
func GetNode(db *gorm.DB, id uint) (Node, error) {
	var n Node
	err := db.First(&n, id).Error
	return n, err
}

// GetNodes returns the live nodes of a collection in sibling order. Tree
// construction groups them by ParentID; the flat result keeps loading simple.
func GetNodes(db *gorm.DB, collectionID uint) ([]Node, error) {
	var nodes []Node
	err := db.Where("collection_id = ?", collectionID).Order("position asc, id asc").Find(&nodes).Error
	return nodes, err
}

// CreateNode inserts a node as a child of parentID, after afterID (or at the
// end when afterID is zero). Both IDs must belong to the same collection.
func CreateNode(db *gorm.DB, collectionID uint, parentID *uint, afterID uint, name string, typ NodeType) (Node, error) {
	if typ != NodeTypeDirectory && typ != NodeTypeRequest {
		return Node{}, fmt.Errorf("unknown node type %q", typ)
	}
	if name == "" {
		return Node{}, errors.New("node name cannot be empty")
	}

	if err := validateParentDB(db, collectionID, parentID); err != nil {
		return Node{}, err
	}
	nodes, err := GetNodes(db, collectionID)
	if err != nil {
		return Node{}, err
	}
	if err := validateParent(nodes, collectionID, parentID); err != nil {
		return Node{}, err
	}
	if afterID != 0 && !isSibling(nodes, afterID, parentID) {
		return Node{}, errors.New("after node is not a sibling of the target parent")
	}

	pos := nextPosition(nodes, parentID)
	if afterID != 0 {
		for _, n := range nodes {
			if n.ID == afterID {
				pos = n.Position + 1
				break
			}
		}
	}

	node := NewNode(name, typ, collectionID)
	node.ParentID = parentID
	node.Position = pos

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := shiftPositions(tx, collectionID, parentID, pos); err != nil {
			return err
		}
		return tx.Create(node).Error
	})
	return *node, err
}

// RenameNode changes a node's name.
func RenameNode(db *gorm.DB, id uint, name string) error {
	if name == "" {
		return errors.New("node name cannot be empty")
	}
	return db.Model(&Node{}).Where("id = ?", id).Update("name", name).Error
}

// DeleteNode removes a node and every descendant, then closes the gap left in
// the deleted node's sibling list.
func DeleteNode(db *gorm.DB, id uint) error {
	node, err := GetNode(db, id)
	if err != nil {
		return err
	}

	nodes, err := GetNodes(db, node.CollectionID)
	if err != nil {
		return err
	}
	ids := append([]uint{id}, descendantIDs(nodes, id)...)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ?", ids).Delete(&Node{}).Error; err != nil {
			return err
		}
		return normalizePositions(tx, node.CollectionID, node.ParentID, 0)
	})
}

// MoveNode reparents a node under parentID, placing it after afterID (or at the
// end when afterID is zero). A node cannot be moved into itself or one of its
// descendants.
func MoveNode(db *gorm.DB, id uint, parentID *uint, afterID uint) error {
	node, err := GetNode(db, id)
	if err != nil {
		return err
	}
	if parentID != nil && *parentID == id {
		return errors.New("a node cannot be moved into itself")
	}
	if afterID == id {
		return errors.New("a node cannot be moved after itself")
	}

	if err := validateParentDB(db, node.CollectionID, parentID); err != nil {
		return err
	}
	nodes, err := GetNodes(db, node.CollectionID)
	if err != nil {
		return err
	}
	if err := validateParent(nodes, node.CollectionID, parentID); err != nil {
		return err
	}
	if afterID != 0 && !isSibling(nodes, afterID, parentID) {
		return errors.New("after node is not a sibling of the target parent")
	}
	if wouldCycle(nodes, id, parentID) {
		return errors.New("a node cannot be moved into its own descendant")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := normalizePositions(tx, node.CollectionID, node.ParentID, id); err != nil {
			return err
		}

		// Re-read after normalization: positions just changed, and deciding
		// which siblings to update from the old snapshot leaves gaps.
		current, err := GetNodes(tx, node.CollectionID)
		if err != nil {
			return err
		}
		siblings := siblingsAfter(current, parentID, id)
		index := len(siblings)
		for i, sibling := range siblings {
			if sibling.ID == afterID {
				index = i + 1
				break
			}
		}

		if err := placeInOrder(tx, siblings, index); err != nil {
			return err
		}
		return tx.Model(&Node{}).Where("id = ?", id).Updates(map[string]any{
			"parent_id": parentID,
			"position":  index,
		}).Error
	})
}

// CopyNode duplicates a node under parentID, placing the copy after afterID (or
// at the end when afterID is zero). Directories copy their whole subtree; the
// original is untouched.
func CopyNode(db *gorm.DB, id uint, parentID *uint, afterID uint) (Node, error) {
	source, err := GetNode(db, id)
	if err != nil {
		return Node{}, err
	}
	if err := validateParentDB(db, source.CollectionID, parentID); err != nil {
		return Node{}, err
	}
	nodes, err := GetNodes(db, source.CollectionID)
	if err != nil {
		return Node{}, err
	}
	if err := validateParent(nodes, source.CollectionID, parentID); err != nil {
		return Node{}, err
	}
	if afterID != 0 && !isSibling(nodes, afterID, parentID) {
		return Node{}, errors.New("after node is not a sibling of the target parent")
	}

	children := make(map[uint][]Node)
	for _, n := range nodes {
		if n.ParentID != nil {
			children[*n.ParentID] = append(children[*n.ParentID], n)
		}
	}

	var copy Node
	err = db.Transaction(func(tx *gorm.DB) error {
		current, err := GetNodes(tx, source.CollectionID)
		if err != nil {
			return err
		}
		pos := nextPosition(current, parentID)
		if afterID != 0 {
			for _, n := range current {
				if n.ID == afterID {
					pos = n.Position + 1
					break
				}
			}
		}
		if err := shiftPositions(tx, source.CollectionID, parentID, pos); err != nil {
			return err
		}
		copy, err = copySubtree(tx, source, parentID, pos, children)
		return err
	})
	return copy, err
}

// copySubtree clones one node and its descendants. Every clone gets a fresh
// primary key so a copy can live inside its own original subtree safely.
func copySubtree(db *gorm.DB, source Node, parentID *uint, pos int, children map[uint][]Node) (Node, error) {
	clone := source
	clone.Model = gorm.Model{}
	clone.ParentID = parentID
	clone.Position = pos
	clone.Children = nil

	if err := db.Create(&clone).Error; err != nil {
		return Node{}, err
	}
	for i, child := range children[source.ID] {
		if _, err := copySubtree(db, child, &clone.ID, i, children); err != nil {
			return Node{}, err
		}
	}
	return clone, nil
}

// ReorderNode moves a node one slot up (delta -1) or down (delta +1) among its
// siblings. It does nothing at either end of the list.
func ReorderNode(db *gorm.DB, id uint, delta int) error {
	if delta != -1 && delta != 1 {
		return errors.New("reorder delta must be -1 or 1")
	}
	node, err := GetNode(db, id)
	if err != nil {
		return err
	}

	nodes, err := GetNodes(db, node.CollectionID)
	if err != nil {
		return err
	}
	siblings := siblingsAfter(nodes, node.ParentID, 0)
	index := -1
	for i, sibling := range siblings {
		if sibling.ID == id {
			index = i
			break
		}
	}
	if index < 0 || index+delta < 0 || index+delta >= len(siblings) {
		return nil
	}
	other := siblings[index+delta]

	return db.Transaction(func(tx *gorm.DB) error {
		if err := setPosition(tx, id, other.Position); err != nil {
			return err
		}
		return setPosition(tx, other.ID, node.Position)
	})
}

func validateParentDB(db *gorm.DB, collectionID uint, parentID *uint) error {
	if parentID == nil {
		return nil
	}
	parent, err := GetNode(db, *parentID)
	if err != nil {
		return fmt.Errorf("parent node %d not found", *parentID)
	}
	if parent.CollectionID != collectionID {
		return errors.New("parent belongs to a different collection")
	}
	return nil
}

func validateParent(nodes []Node, collectionID uint, parentID *uint) error {
	if parentID == nil {
		return nil
	}
	for _, n := range nodes {
		if n.ID == *parentID {
			if n.CollectionID != collectionID {
				return errors.New("parent belongs to a different collection")
			}
			if n.Type != NodeTypeDirectory {
				return errors.New("only directories can have children")
			}
			return nil
		}
	}
	return fmt.Errorf("parent node %d not found", *parentID)
}

func isSibling(nodes []Node, id uint, parentID *uint) bool {
	for _, n := range nodes {
		if n.ID != id {
			continue
		}
		if parentID == nil {
			return n.ParentID == nil
		}
		return n.ParentID != nil && *n.ParentID == *parentID
	}
	return false
}

func nextPosition(nodes []Node, parentID *uint) int {
	maxPos := -1
	for _, n := range nodes {
		if sameParent(n, parentID) && n.Position > maxPos {
			maxPos = n.Position
		}
	}
	return maxPos + 1
}

func sameParent(n Node, parentID *uint) bool {
	if parentID == nil {
		return n.ParentID == nil
	}
	return n.ParentID != nil && *n.ParentID == *parentID
}

func parentClause(parentID *uint) (string, []any) {
	if parentID == nil {
		return "parent_id IS NULL", nil
	}
	return "parent_id = ?", []any{*parentID}
}

func shiftPositions(db *gorm.DB, collectionID uint, parentID *uint, from int) error {
	clause, args := parentClause(parentID)
	where := append([]any{collectionID}, args...)
	where = append(where, from)
	return db.Model(&Node{}).
		Where("collection_id = ? AND "+clause+" AND position >= ?", where...).
		UpdateColumn("position", gorm.Expr("position + 1")).Error
}

func normalizePositions(db *gorm.DB, collectionID uint, parentID *uint, exclude uint) error {
	nodes, err := GetNodes(db, collectionID)
	if err != nil {
		return err
	}
	siblings := siblingsAfter(nodes, parentID, exclude)
	for i, n := range siblings {
		if n.Position != i {
			if err := setPosition(db, n.ID, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func siblingsAfter(nodes []Node, parentID *uint, exclude uint) []Node {
	var out []Node
	for _, n := range nodes {
		if n.ID != exclude && sameParent(n, parentID) {
			out = append(out, n)
		}
	}
	return out
}

// placeInOrder writes a dense position to every sibling. Skipping rows whose
// stale Position already matches the target is unsafe: a prior normalization
// may have changed the stored value even when the old snapshot says otherwise.
func placeInOrder(db *gorm.DB, siblings []Node, index int) error {
	for i, sibling := range siblings {
		pos := i
		if i >= index {
			pos++
		}
		if err := setPosition(db, sibling.ID, pos); err != nil {
			return err
		}
	}
	return nil
}

func setPosition(db *gorm.DB, id uint, pos int) error {
	return db.Model(&Node{}).Where("id = ?", id).Update("position", pos).Error
}

func descendantIDs(nodes []Node, id uint) []uint {
	children := make(map[uint][]uint)
	for _, n := range nodes {
		if n.ParentID != nil {
			children[*n.ParentID] = append(children[*n.ParentID], n.ID)
		}
	}

	var ids []uint
	stack := []uint{id}
	for len(stack) > 0 {
		last := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, child := range children[last] {
			ids = append(ids, child)
			stack = append(stack, child)
		}
	}
	return ids
}

func wouldCycle(nodes []Node, id uint, parentID *uint) bool {
	if parentID == nil {
		return false
	}
	byID := make(map[uint]Node, len(nodes))
	for _, n := range nodes {
		byID[n.ID] = n
	}

	seen := make(map[uint]bool)
	current := *parentID
	for current != 0 {
		if current == id || seen[current] {
			return true
		}
		seen[current] = true
		n, ok := byID[current]
		if !ok || n.ParentID == nil {
			return false
		}
		current = *n.ParentID
	}
	return false
}
