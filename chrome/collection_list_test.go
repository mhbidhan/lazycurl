package chrome

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/store"
)

func listRows(n int) []store.Collection {
	rows := make([]store.Collection, n)
	for i := range rows {
		rows[i] = store.Collection{Name: string(rune('a' + i))}
	}
	return rows
}

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		s, pattern string
		want       bool
	}{
		{"Collections", "cl", true},
		{"Collections", "COL", true},
		{"Collections", "xyz", false},
		{"Collections", "", true},
		{"work", "wok", true},
		{"work", "krow", false},
	}
	for _, tc := range cases {
		if got := FuzzyMatch(tc.s, tc.pattern); got != tc.want {
			t.Errorf("FuzzyMatch(%q, %q) = %t, want %t", tc.s, tc.pattern, got, tc.want)
		}
	}
}

func TestCollectionListFilterAndCursor(t *testing.T) {
	l := newCollectionList(30, 8)
	l.setRows([]store.Collection{
		{Name: "alpha"}, {Name: "bravo"}, {Name: "charlie"},
	})
	if len(l.filtered) != 3 || l.cursor != 0 {
		t.Fatalf("rows = %d, cursor = %d", len(l.filtered), l.cursor)
	}

	l.setQuery("alp")
	if len(l.filtered) != 1 || l.filtered[0].Name != "alpha" {
		t.Fatalf("filtered = %+v", l.filtered)
	}
	if coll, ok := l.current(); !ok || coll.Name != "alpha" {
		t.Fatalf("current = %+v, %t", coll, ok)
	}

	l.setQuery("zzz")
	if !l.empty() || l.cursor != 0 {
		t.Fatalf("empty filter cursor = %d", l.cursor)
	}
	if _, ok := l.current(); ok {
		t.Fatal("empty list returned a current row")
	}

	l.clearQuery()
	if len(l.filtered) != 3 || l.cursor != 0 {
		t.Fatalf("cleared filter = %+v, cursor = %d", l.filtered, l.cursor)
	}

	l.moveUp()
	if l.cursor != 2 {
		t.Fatalf("moveUp wrapped to %d, want 2", l.cursor)
	}
	l.moveDown()
	if l.cursor != 0 {
		t.Fatalf("moveDown wrapped to %d, want 0", l.cursor)
	}

	l.moveDown()
	l.moveDown()
	l.setQuery("charlie")
	if l.cursor != 0 {
		t.Fatalf("cursor not clamped to filtered list: %d", l.cursor)
	}

	l.setRows(nil)
	l.moveDown()
	l.moveUp()
	if l.cursor != 0 {
		t.Fatalf("cursor moved on an empty list: %d", l.cursor)
	}
}

func TestCollectionListScroll(t *testing.T) {
	l := newCollectionList(30, 6)
	l.SetSize(30, 6)
	l.setRows(listRows(12))
	l.syncView()

	for range 11 {
		l.moveDown()
		y := l.rowY(l.cursor)
		if y < l.vp.YOffset || y >= l.vp.YOffset+l.vp.Height {
			t.Fatalf("cursor %d outside viewport [%d,%d)", l.cursor, l.vp.YOffset, l.vp.YOffset+l.vp.Height)
		}
		if !strings.Contains(l.view(), l.filtered[l.cursor].Name) {
			t.Fatalf("cursor %d not visible", l.cursor)
		}
	}
	if l.vp.YOffset == 0 {
		t.Fatal("scrolling down did not move the viewport")
	}

	for range 11 {
		l.moveUp()
		if !strings.Contains(l.view(), l.filtered[l.cursor].Name) {
			t.Fatalf("cursor %d not visible after scrolling up", l.cursor)
		}
	}
	if l.vp.YOffset != 0 {
		t.Fatalf("offset after scrolling up = %d, want 0", l.vp.YOffset)
	}
}

func TestCollectionListShowsFourRows(t *testing.T) {
	l := newCollectionList(30, listMinHeight)
	l.SetSize(30, listMinHeight)
	l.setRows(listRows(8))
	l.syncView()

	if l.vp.Height != listContentLines {
		t.Fatalf("viewport height = %d, want %d", l.vp.Height, listContentLines)
	}
	row := func(name string) string { return " " + name + " " }
	for _, want := range []string{"a", "b", "c", "d"} {
		if !strings.Contains(l.view(), row(want)) {
			t.Fatalf("top of list is missing %q:\n%s", want, l.view())
		}
	}
	if strings.Contains(l.view(), row("e")) {
		t.Fatalf("fifth row visible at the top:\n%s", l.view())
	}

	for range listVisibleRows {
		l.moveDown()
	}
	if l.vp.YOffset != listRowLines {
		t.Fatalf("offset after four moves = %d, want %d", l.vp.YOffset, listRowLines)
	}
	for _, want := range []string{"b", "c", "d", "e"} {
		if !strings.Contains(l.view(), row(want)) {
			t.Fatalf("scrolled list is missing %q:\n%s", want, l.view())
		}
	}
	if strings.Contains(l.view(), row("a")) {
		t.Fatalf("first row still visible after scrolling:\n%s", l.view())
	}
}

func TestCollectionListRenderingAndSizing(t *testing.T) {
	l := newCollectionList(30, 6)
	l.setRows([]store.Collection{{Name: "work"}})
	l.syncView()

	lines := strings.Split(l.view(), "\n")
	if strings.TrimSpace(lines[0]) != "Collections" {
		t.Fatalf("first line = %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "" {
		t.Fatalf("second line = %q, want the title gap", lines[1])
	}
	if !strings.Contains(l.renderRows(), "work") {
		t.Fatal("selected row missing from viewport content")
	}

	l.setRows(nil)
	l.syncView()
	if !strings.Contains(l.view(), "No collections found") {
		t.Fatal("empty list message missing")
	}
	if !strings.Contains(l.view(), "Collections") {
		t.Fatal("empty list lost its title")
	}

	for _, h := range []int{1, 2, 3, 6, 12, 20} {
		l.SetSize(30, h)
		if got := lipgloss.Height(l.view()); got != max(h, 2) {
			t.Errorf("h=%d: list rendered %d rows", h, got)
		}
		if want := max(h-listHeaderLines, 1); l.vp.Height != want {
			t.Errorf("h=%d: viewport height = %d, want %d", h, l.vp.Height, want)
		}
	}

	// Unhandled messages reach the viewport (mouse wheel, resize, etc.).
	l.update(tea.MouseMsg{Type: tea.MouseWheelDown, X: 1, Y: 1})
}
