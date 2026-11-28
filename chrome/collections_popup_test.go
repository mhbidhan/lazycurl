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

func TestOpenCollectionMessage(t *testing.T) {
	p := testPopup(t, 40, 12)
	coll := store.NewCollection("work")
	if err := store.SaveCollection(p.db, coll); err != nil {
		t.Fatal(err)
	}
	if err := p.reload(); err != nil {
		t.Fatal(err)
	}

	_, cmd := p.Update(enterKey())
	if cmd == nil {
		t.Fatal("enter did not ask to open the collection")
	}
	msg, ok := cmd().(OpenCollectionMsg)
	if !ok {
		t.Fatalf("enter returned %T, want OpenCollectionMsg", cmd())
	}
	if msg.Collection.ID != coll.ID || msg.Collection.Name != "work" {
		t.Fatalf("open message = %+v", msg)
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

func TestPopupShowsFourCollectionsAtMaxSize(t *testing.T) {
	p := testPopup(t, 40, 14)
	for _, name := range []string{"one", "two", "three", "four", "five", "six"} {
		if err := store.SaveCollection(p.db, store.NewCollection(name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.reload(); err != nil {
		t.Fatal(err)
	}

	if p.list.vp.Height != listContentLines {
		t.Fatalf("viewport height = %d, want %d", p.list.vp.Height, listContentLines)
	}
	view := p.View()
	for _, want := range []string{"one", "two", "three", "four"} {
		if !strings.Contains(view, want) {
			t.Fatalf("popup is missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "five") {
		t.Fatalf("fifth collection visible without scrolling:\n%s", view)
	}
}

func TestHelpLineHasOnlyFramePaddingBelow(t *testing.T) {
	p := testPopup(t, 40, 14)
	if err := store.SaveCollection(p.db, store.NewCollection("work")); err != nil {
		t.Fatal(err)
	}
	if err := p.reload(); err != nil {
		t.Fatal(err)
	}

	blankRow := func(line string) bool { return strings.Trim(line, " │\t") == "" }
	helpIndex := func(lines []string) int {
		t.Helper()
		for i, line := range lines {
			if strings.Contains(line, "cursor down") {
				return i
			}
		}
		t.Fatalf("help line missing:\n%s", strings.Join(lines, "\n"))
		return -1
	}
	onlyFramePadding := func(lines []string, i int) bool {
		return i+1 < len(lines) && blankRow(lines[i+1]) &&
			i+2 < len(lines) && strings.HasPrefix(lines[i+2], "╰")
	}

	lines := strings.Split(p.View(), "\n")
	i := helpIndex(lines)
	if !blankRow(lines[i-1]) || !onlyFramePadding(lines, i) {
		t.Fatalf("idle layout around help is wrong:\n%s", strings.Join(lines, "\n"))
	}

	p.err = errors.New("boom")
	lines = strings.Split(p.View(), "\n")
	i = helpIndex(lines)
	if !strings.Contains(lines[i-1], "Error: boom") {
		t.Fatalf("error is not above the help line:\n%s", strings.Join(lines, "\n"))
	}
	if !onlyFramePadding(lines, i) {
		t.Fatalf("error added a row below help:\n%s", strings.Join(lines, "\n"))
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
