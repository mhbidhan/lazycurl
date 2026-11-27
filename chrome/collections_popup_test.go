package chrome

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/store"
	"github.com/mhbidhan/lazycurl/ui"
)

func enterKey() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

func TestCreateFlow(t *testing.T) {
	p := testPopup(t, 40, 12)
	p.Update(keyMsg("n"))
	if p.state != stateCreate || !p.pm.IsOpen() {
		t.Fatalf("create state = %d, popup open = %t", p.state, p.pm.IsOpen())
	}

	p.Update(keyMsg("work"))
	_, cmd := p.Update(enterKey())
	if cmd == nil {
		t.Fatal("submit did not return an event")
	}
	ev, ok := cmd().(NamePromptEvent)
	if !ok {
		t.Fatalf("submit event = %T", cmd())
	}

	p.Update(ev)
	if p.state != stateList || p.pm.IsOpen() {
		t.Fatalf("create did not return to the list: state = %d, open = %t", p.state, p.pm.IsOpen())
	}
	rows, err := store.GetCollections(p.db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "work" {
		t.Fatalf("rows after create = %+v", rows)
	}
}

func TestFailedCreateShowsStatusLine(t *testing.T) {
	p := testPopup(t, 40, 12)
	sqlDB, err := p.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}

	p.Update(keyMsg("n"))
	p.Update(keyMsg("work"))
	_, cmd := p.Update(enterKey())
	if cmd == nil {
		t.Fatal("submit did not return an event")
	}
	_, cmd = p.Update(cmd().(NamePromptEvent))
	if cmd == nil {
		t.Fatal("failed save did not return an error command")
	}
	msg, ok := cmd().(OperationError)
	if !ok {
		t.Fatalf("error command returned %T", cmd())
	}

	p.Update(msg)
	if p.err == nil {
		t.Fatal("operation error was not kept")
	}
	if !strings.Contains(p.View(), "Error:") {
		t.Fatalf("status line missing from view:\n%s", p.View())
	}

	p.report(nil)
	if p.err != nil {
		t.Fatal("success did not clear the status line")
	}
}

func TestSearchFlow(t *testing.T) {
	p := testPopup(t, 40, 12)
	if err := store.SaveCollection(p.db, store.NewCollection("work")); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveCollection(p.db, store.NewCollection("home")); err != nil {
		t.Fatal(err)
	}
	if err := p.reload(); err != nil {
		t.Fatal(err)
	}

	p.Update(keyMsg("/"))
	if p.state != stateSearch {
		t.Fatalf("search state = %d", p.state)
	}
	p.Update(keyMsg("wo"))
	if p.input.Value() != "wo" {
		t.Fatalf("filter = %q", p.input.Value())
	}
	if len(p.list.filtered) != 1 {
		t.Fatalf("filtered rows = %d", len(p.list.filtered))
	}

	p.Update(enterKey())
	if p.state != stateList || p.list.query != "wo" {
		t.Fatalf("submit left state = %d, query = %q", p.state, p.list.query)
	}

	p.Update(keyMsg("/"))
	p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if p.state != stateList || p.list.query != "" {
		t.Fatalf("cancel left state = %d, query = %q", p.state, p.list.query)
	}
}

func TestDeleteFlow(t *testing.T) {
	p := testPopup(t, 40, 12)
	if err := store.SaveCollection(p.db, store.NewCollection("work")); err != nil {
		t.Fatal(err)
	}
	if err := p.reload(); err != nil {
		t.Fatal(err)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if p.state != stateConfirmDelete || !p.pm.IsOpen() {
		t.Fatalf("delete state = %d, open = %t", p.state, p.pm.IsOpen())
	}

	p.Update(ConfirmEvent{Confirmed: true})
	if p.state != stateList || p.pm.IsOpen() {
		t.Fatalf("delete did not return to the list: state = %d", p.state)
	}
	rows, err := store.GetCollections(p.db)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows after delete = %+v", rows)
	}
}

func TestRenameFlow(t *testing.T) {
	p := testPopup(t, 40, 12)
	if err := store.SaveCollection(p.db, store.NewCollection("work")); err != nil {
		t.Fatal(err)
	}
	if err := p.reload(); err != nil {
		t.Fatal(err)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if p.state != stateRename {
		t.Fatalf("rename state = %d", p.state)
	}

	p.Update(NamePromptEvent{Cancel: true})
	if p.state != stateList {
		t.Fatal("cancelled rename did not return to the list")
	}

	p.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	p.Update(NamePromptEvent{Value: "home"})
	if p.state != stateRename || !p.pm.IsOpen() {
		t.Fatalf("rename confirmation state = %d, open = %t", p.state, p.pm.IsOpen())
	}
	p.Update(NamePromptEvent{Value: "home", Cancel: true})
	if p.state != stateList {
		t.Fatal("cancelled rename confirmation did not return to the list")
	}

	p.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	p.Update(NamePromptEvent{Value: "home"})
	p.Update(ConfirmEvent{Confirmed: false})
	if p.state != stateList {
		t.Fatal("rejected rename did not return to the list")
	}
	rows, _ := store.GetCollections(p.db)
	if rows[0].Name != "work" {
		t.Fatalf("rejected rename changed the name to %q", rows[0].Name)
	}

	p.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	p.Update(NamePromptEvent{Value: "home"})
	p.Update(ConfirmEvent{Confirmed: true})
	rows, _ = store.GetCollections(p.db)
	if rows[0].Name != "home" {
		t.Fatalf("rename saved %q, want home", rows[0].Name)
	}
}

func TestPopupViewFitsItsArea(t *testing.T) {
	for _, tc := range []struct{ w, h int }{{40, 12}, {30, 10}, {20, 8}, {14, 6}} {
		p := testPopup(t, tc.w, tc.h)
		if err := store.SaveCollection(p.db, store.NewCollection("work")); err != nil {
			t.Fatal(err)
		}
		if err := p.reload(); err != nil {
			t.Fatal(err)
		}

		fits := func(what string, framed bool) {
			t.Helper()
			view := p.View()
			wantW, wantH := tc.w, tc.h
			if framed {
				wantW += ui.PopupFrameWidth
				wantH += ui.PopupFrameHeight
			}
			if w, h := lipgloss.Width(view), lipgloss.Height(view); w > wantW || h > wantH {
				t.Errorf("%dx%d: %s = %dx%d, want at most %dx%d", tc.w, tc.h, what, w, h, wantW, wantH)
			}
		}

		fits("list", true)
		before := lipgloss.Height(p.View())

		p.Update(keyMsg("/"))
		fits("search", true)
		p.Update(tea.KeyMsg{Type: tea.KeyEsc})

		p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
		fits("confirm", false)
		p.Update(ConfirmEvent{Confirmed: false})

		p.err = errors.New("a failure with a message much longer than the popup is wide")
		fits("error", true)
		if got := lipgloss.Height(p.View()); got != before {
			t.Errorf("%dx%d: error changed height from %d to %d", tc.w, tc.h, before, got)
		}
	}
}

func TestPopupIgnoresUnknownMessages(t *testing.T) {
	p := testPopup(t, 40, 12)
	if _, cmd := p.Update(struct{}{}); cmd != nil {
		t.Fatalf("unknown message returned %T", cmd())
	}
}
