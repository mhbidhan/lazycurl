package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestScrollPreservedOnResize(t *testing.T) {
	v := NewViewport(20, 20, nil, nil)
	v.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v.SetContent(strings.Repeat("line\n", 200))

	for range 10 {
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	before := v.viewport.YOffset
	v.Update(tea.WindowSizeMsg{Width: 90, Height: 30})

	if after := v.viewport.YOffset; after != before {
		t.Errorf("scroll position lost on resize: %d -> %d", before, after)
	}
	if v.viewport.Width != 90 || v.viewport.Height != 30 {
		t.Errorf("size not applied: got %dx%d", v.viewport.Width, v.viewport.Height)
	}
	if v.viewport.TotalLineCount() == 0 {
		t.Error("content dropped on resize")
	}
}

func TestHeightClampedWhenChromeExceedsTerminal(t *testing.T) {
	v := NewViewport(20, 20, &fakeComponent{h: 40}, &fakeComponent{h: 40})
	v.Update(tea.WindowSizeMsg{Width: 80, Height: 10})

	if v.viewport.Height < 1 {
		t.Errorf("height went negative: %d", v.viewport.Height)
	}
	if got := v.ContentHeight(); got < 1 {
		t.Errorf("content height went negative: %d", got)
	}
}

func TestChromeReservedFromHeight(t *testing.T) {
	v := NewViewport(20, 20, &fakeComponent{h: 3}, nil)
	v.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	if v.viewport.Height != 37 {
		t.Errorf("expected viewport height 37, got %d", v.viewport.Height)
	}
	if got := len(strings.Split(v.View(), "\n")); got != 40 {
		t.Errorf("expected 40 rendered lines, got %d", got)
	}
}

func TestContentWidthExcludesStyleFrame(t *testing.T) {
	v := NewViewport(20, 20, nil, nil)
	v.SetStyle(lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()))
	v.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	fx, fy := v.viewport.Style.GetFrameSize()
	if want := 80 - fx; v.ContentWidth() != want {
		t.Errorf("ContentWidth = %d, want %d", v.ContentWidth(), want)
	}
	if want := 24 - fy; v.ContentHeight() != want {
		t.Errorf("ContentHeight = %d, want %d", v.ContentHeight(), want)
	}
}

func TestViewRespectsOuterBounds(t *testing.T) {
	v := NewViewport(20, 20, nil, nil)
	v.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	v.SetContent(strings.Repeat("a very long line of content\n", 200))

	for _, l := range strings.Split(v.View(), "\n") {
		if w := lipgloss.Width(l); w > 60 {
			t.Errorf("line width %d exceeds viewport width 60: %q", w, l)
		}
	}
}

type fakeComponent struct{ h int }

func (f *fakeComponent) Height() int { return f.h }

func (f *fakeComponent) View() string {
	return strings.Repeat("x\n", f.h-1) + "x"
}
