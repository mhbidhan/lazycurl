package keybind

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func runeKey(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestHomeKeys(t *testing.T) {
	k := NewHomeKeys()
	if !key.Matches(runeKey("?"), k.Keybind) {
		t.Error("? should open keybinds")
	}
	if !key.Matches(runeKey("c"), k.Collections) {
		t.Error("c should open collections")
	}
	if !key.Matches(runeKey("q"), k.Quit) {
		t.Error("q should quit")
	}

	help := k.ShortHelp()
	if !strings.Contains(help, "keybinds") || !strings.Contains(help, "quit") {
		t.Fatalf("short help = %q", help)
	}
	if len(k.GetKeybinds()) != 3 {
		t.Fatalf("keybinds = %+v", k.GetKeybinds())
	}
}

func TestCollectionKeys(t *testing.T) {
	k := NewCollectionKeys()
	cases := []struct {
		msg string
		kb  key.Binding
	}{
		{"j", k.CursorDown},
		{"k", k.CursorUp},
		{"n", k.New},
		{"/", k.Search},
		{"esc", k.Escape},
	}
	for _, tc := range cases {
		if !key.Matches(runeKey(tc.msg), tc.kb) {
			t.Errorf("%q does not match binding %+v", tc.msg, tc.kb)
		}
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyEnter}, k.Submit) {
		t.Error("enter should apply the filter")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyCtrlD}, k.Delete) {
		t.Error("ctrl+d should delete")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyCtrlE}, k.Edit) {
		t.Error("ctrl+e should rename")
	}

	help := k.ShortHelp()
	for _, want := range []string{"cursor down", "cursor up", "cancel"} {
		if !strings.Contains(help, want) {
			t.Errorf("short help %q missing %q", help, want)
		}
	}

	bindings := k.GetKeybinds()
	if len(bindings) != 9 {
		t.Fatalf("keybinds = %+v", bindings)
	}
	if bindings[7].Key != "enter" || bindings[7].Help != "apply filter" {
		t.Fatalf("submit entry = %+v", bindings[7])
	}
	if bindings[8].Key != "enter" || bindings[8].Help != "open collection" {
		t.Fatalf("open entry = %+v", bindings[8])
	}
}

func TestExplorerKeys(t *testing.T) {
	k := NewExplorerKeys()

	for _, tc := range []struct {
		msg  tea.KeyMsg
		bind key.Binding
	}{
		{runeKey("j"), k.CursorDown},
		{runeKey("k"), k.CursorUp},
		{runeKey("h"), k.Collapse},
		{runeKey("l"), k.Expand},
		{runeKey("enter"), k.Toggle},
		{runeKey("g"), k.Top},
		{runeKey("G"), k.Bottom},
		{runeKey("/"), k.Search},
		{runeKey("a"), k.AddChild},
		{runeKey("A"), k.AddSibling},
		{runeKey("d"), k.Delete},
		{runeKey("r"), k.Rename},
		{runeKey("x"), k.Cut},
		{runeKey("c"), k.Copy},
		{runeKey("p"), k.Paste},
		{runeKey("J"), k.MoveDown},
		{runeKey("K"), k.MoveUp},
		{runeKey("?"), k.Help},
		{runeKey("q"), k.Back},
		{tea.KeyMsg{Type: tea.KeyDown}, k.CursorDown},
		{tea.KeyMsg{Type: tea.KeyRight}, k.Expand},
	} {
		if !key.Matches(tc.msg, tc.bind) {
			t.Errorf("%v should match %q", tc.msg, tc.bind.Help().Key)
		}
	}
	if key.Matches(runeKey("m"), k.Cut) || key.Matches(runeKey("m"), k.Copy) {
		t.Error("m should no longer mark a node")
	}

	help := k.ShortHelp()
	for _, want := range []string{"cursor down", "toggle", "add request", "mark to move", "mark to copy", "back to home"} {
		if !strings.Contains(help, want) {
			t.Errorf("short help %q missing %q", help, want)
		}
	}
	if got := len(k.GetKeybinds()); got != 19 {
		t.Fatalf("keybinds = %d, want 19", got)
	}
}

func TestPromptKeys(t *testing.T) {
	confirm := NewConfirmKeys()
	if !key.Matches(runeKey("y"), confirm.Confirm) {
		t.Error("y should confirm")
	}
	if !key.Matches(runeKey("n"), confirm.Cancel) || !key.Matches(runeKey("esc"), confirm.Cancel) {
		t.Error("n/esc should cancel")
	}
	if help := confirm.ShortHelp(); !strings.Contains(help, "confirm") {
		t.Fatalf("confirm help = %q", help)
	}

	prompt := NewNamePromptKeys()
	if !key.Matches(tea.KeyMsg{Type: tea.KeyEnter}, prompt.Submit) {
		t.Error("enter should submit a name")
	}
	if !key.Matches(runeKey("esc"), prompt.Cancel) {
		t.Error("esc should cancel a name prompt")
	}

	popup := NewHelpPopupKeys()
	if !key.Matches(runeKey("esc"), popup.Escape) {
		t.Error("esc should close the help popup")
	}
}
