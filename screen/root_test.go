package screen

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestViewFitsTerminal(t *testing.T) {
	sizes := [][2]int{{120, 40}, {80, 24}, {60, 20}, {20, 8}, {10, 4}, {4, 2}}

	for _, s := range sizes {
		r := NewRoot()
		r.Init()
		r.Update(tea.WindowSizeMsg{Width: s[0], Height: s[1]})

		lines := strings.Split(r.View(), "\n")
		widest := 0
		for _, l := range lines {
			if w := lipgloss.Width(l); w > widest {
				widest = w
			}
		}

		if widest > s[0] {
			t.Errorf("terminal %dx%d: rendered width %d overflows", s[0], s[1], widest)
		}
		if len(lines) > s[1] {
			t.Errorf("terminal %dx%d: rendered height %d overflows", s[0], s[1], len(lines))
		}
	}
}

func TestHelpToggleRerenders(t *testing.T) {
	r := NewRoot()
	r.Init()
	r.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	short := r.View()

	r.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if full := r.View(); full == short {
		t.Error("pressing ? did not change the view")
	}

	r.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if back := r.View(); back != short {
		t.Error("pressing ? again did not restore the short help")
	}
}
