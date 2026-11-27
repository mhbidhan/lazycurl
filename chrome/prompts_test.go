package chrome

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/keybind"
)

func TestConfirmPopup(t *testing.T) {
	p := NewConfirmPopup("Delete?")
	p.SetSize(30, 8)
	if p.Init() != nil {
		t.Fatal("Init returned a command")
	}
	if view := p.View(); !strings.Contains(view, "Delete?") || !strings.Contains(view, "confirm") {
		t.Fatalf("confirm view:\n%s", view)
	}

	if _, cmd := p.Update(keyMsg("x")); cmd != nil {
		t.Fatalf("unmatched key returned %T", cmd())
	}
	_, cmd := p.Update(keyMsg("y"))
	ev, ok := cmd().(ConfirmEvent)
	if !ok || !ev.Confirmed {
		t.Fatalf("y event = %+v", cmd())
	}
	_, cmd = p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if ev, ok := cmd().(ConfirmEvent); !ok || ev.Confirmed {
		t.Fatalf("esc event = %+v", cmd())
	}
}

func TestNamePromptPopup(t *testing.T) {
	p := NewNamePromptPopup("Name: ", "old")
	p.SetSize(30, 8)
	p.OnOpen()
	if !p.input.Focused() {
		t.Fatal("prompt did not focus on open")
	}

	p.Update(keyMsg("x"))
	if p.input.Value() != "oldx" {
		t.Fatalf("input value = %q", p.input.Value())
	}

	_, cmd := p.Update(enterKey())
	ev, ok := cmd().(NamePromptEvent)
	if !ok || ev.Cancel || ev.Value != "oldx" {
		t.Fatalf("submit event = %+v", cmd())
	}

	_, cmd = p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if ev, ok := cmd().(NamePromptEvent); !ok || !ev.Cancel {
		t.Fatalf("cancel event = %+v", cmd())
	}

	p.OnOpen()
	if p.input.Value() != "old" {
		t.Fatalf("OnOpen kept stale value %q", p.input.Value())
	}
	if !strings.Contains(p.View(), "Name:") {
		t.Fatal("prompt label missing from view")
	}

	p.SetSize(5, 8)
	if p.input.Prompt != "" {
		t.Fatal("overlong prompt was not dropped on a narrow popup")
	}
	p.SetSize(30, 8)
	if p.input.Prompt != "Name: " {
		t.Fatalf("prompt not restored: %q", p.input.Prompt)
	}
}

func TestHelpPopup(t *testing.T) {
	p := NewHelpPopup([]keybind.KeyInfo{{Key: "?", Help: "keybinds"}}, 30, 8)
	p.SetSize(20, 6)
	if p.Init() != nil {
		t.Fatal("Init returned a command")
	}
	if view := p.View(); !strings.Contains(view, "?: keybinds") {
		t.Fatalf("help view:\n%s", view)
	}
	if _, cmd := p.Update(struct{}{}); cmd != nil {
		t.Fatalf("non-key message returned %T", cmd())
	}
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if _, ok := cmd().(DismissPopup); !ok {
		t.Fatalf("esc command returned %T", cmd())
	}
}

func TestRenderKeybinds(t *testing.T) {
	if got := RenderKeybinds(nil); got != "" {
		t.Fatalf("empty keybinds = %q", got)
	}
	got := RenderKeybinds([]keybind.KeyInfo{{Key: "j", Help: "cursor down"}})
	if !strings.Contains(got, "j") || !strings.Contains(got, "cursor down") {
		t.Fatalf("rendered keybinds = %q", got)
	}
	if strings.Count(got, "\n") != 0 {
		t.Fatalf("one keybind rendered on multiple lines: %q", got)
	}
}
