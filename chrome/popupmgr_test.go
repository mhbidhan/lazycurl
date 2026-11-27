package chrome

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type probePopup struct {
	w, h    int
	opens   int
	updates int
}

func (p *probePopup) Init() tea.Cmd { return nil }

func (p *probePopup) Update(tea.Msg) (tea.Model, tea.Cmd) {
	p.updates++
	return p, func() tea.Msg { return DismissPopup{} }
}

func (p *probePopup) View() string { return "probe" }

func (p *probePopup) SetSize(w, h int) {
	p.w, p.h = w, h
}

func (p *probePopup) OnOpen() { p.opens++ }

func TestPopupManager(t *testing.T) {
	pm := NewPopupManager()
	pm.SetSize(30, 12)

	first := &probePopup{}
	pm.Register("first", first)
	if first.w != 30 || first.h != 12 {
		t.Fatalf("registered popup size = %dx%d", first.w, first.h)
	}

	pm.Open("missing")
	if pm.IsOpen() {
		t.Fatal("unknown id became active")
	}

	pm.Open("first")
	if first.opens != 1 {
		t.Fatalf("OnOpen called %d times", first.opens)
	}
	id, active := pm.Active()
	if id != "first" || active != tea.Model(first) {
		t.Fatalf("active = (%q, %T)", id, active)
	}

	cmd := pm.Update(keyMsg("x"))
	if cmd == nil {
		t.Fatal("Update returned no command")
	}
	if _, ok := cmd().(DismissPopup); !ok {
		t.Fatalf("command returned %T", cmd())
	}
	if first.updates != 1 {
		t.Fatalf("active popup updated %d times", first.updates)
	}

	pm.Close("missing")
	if !pm.IsOpen() {
		t.Fatal("closing an inactive popup changed the active one")
	}
	pm.Close("first")
	if pm.IsOpen() {
		t.Fatal("popup still open after Close")
	}
	if _, active := pm.Active(); active != nil {
		t.Fatal("Active returned a closed popup")
	}
	if cmd := pm.Update(keyMsg("x")); cmd != nil {
		t.Fatal("Update with no active popup returned a command")
	}

	pm.SetSize(10, 8)
	if first.w != 10 || first.h != 8 {
		t.Fatalf("resize did not reach registered popup: %dx%d", first.w, first.h)
	}
	second := &probePopup{}
	pm.Register("second", second)
	if second.w != 10 || second.h != 8 {
		t.Fatalf("late registration used stale size: %dx%d", second.w, second.h)
	}
}

func TestFieldWidth(t *testing.T) {
	if got := fieldWidth(20, "/ "); got != 17 {
		t.Fatalf("fieldWidth = %d, want 17", got)
	}
	if got := fieldWidth(2, "/ "); got != 1 {
		t.Fatalf("fieldWidth floor = %d, want 1", got)
	}
}
