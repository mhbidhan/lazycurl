package screen

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glebarez/sqlite"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/store"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.Collection{}, &store.Node{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestHomeScreenRouting(t *testing.T) {
	s := NewHomeScreen(testDB(t), 80, 24)

	s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	s.Update(keyMsg("?"))
	if !strings.Contains(s.View(), "?: keybinds") {
		t.Fatalf("help popup not shown:\n%s", s.View())
	}

	// Resizing must reach an open popup without closing it.
	s.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	if lipgloss.Height(s.View()) != 20 {
		t.Fatalf("resized view height = %d", lipgloss.Height(s.View()))
	}
	if !strings.Contains(s.View(), "?: keybinds") {
		t.Fatal("resize closed the help popup")
	}

	s.Update(chrome.DismissPopup{})
	if !strings.Contains(s.View(), "? keybinds") {
		t.Fatalf("home screen not restored:\n%s", s.View())
	}

	s.Update(keyMsg("c"))
	if !strings.Contains(s.View(), "Collections") {
		t.Fatalf("collections popup not shown:\n%s", s.View())
	}
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc did not ask to close collections")
	}
	s.Update(cmd())
	if !strings.Contains(s.View(), "? keybinds") {
		t.Fatal("esc did not close collections")
	}

	_, cmd = s.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("q did not return a command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("q command returned %T, want tea.QuitMsg", cmd())
	}
}

func TestHomeScreenIgnoresGlobalKeysUnderPopup(t *testing.T) {
	s := NewHomeScreen(testDB(t), 80, 24)
	s.Update(keyMsg("c"))

	// q belongs to the home screen; while a popup is open it must not quit.
	_, cmd := s.Update(keyMsg("q"))
	if cmd != nil {
		t.Fatalf("global key leaked through popup: %T", cmd())
	}
	if !strings.Contains(s.View(), "Collections") {
		t.Fatal("q closed the collections popup")
	}
}

func TestOverlay(t *testing.T) {
	got := overlay(20, 6, "x")
	if lipgloss.Width(got) != 20 || lipgloss.Height(got) != 6 {
		t.Fatalf("overlay = %dx%d, want 20x6", lipgloss.Width(got), lipgloss.Height(got))
	}

	cropped := overlay(4, 2, "abcdefghij")
	if lipgloss.Width(cropped) != 4 || lipgloss.Height(cropped) != 2 {
		t.Fatalf("cropped overlay = %dx%d, want 4x2", lipgloss.Width(cropped), lipgloss.Height(cropped))
	}
}
