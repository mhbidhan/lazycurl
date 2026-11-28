package chrome

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glebarez/sqlite"
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

func testPopup(t *testing.T, w, h int) *CollectionsPopup {
	t.Helper()
	p := NewCollectionsPopup(testDB(t), w, h)
	p.SetSize(w, h)
	return p
}
