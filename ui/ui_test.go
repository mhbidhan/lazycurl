package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGapVertical(t *testing.T) {
	if got := lipgloss.Height(GapVertical(2)); got != 2 {
		t.Fatalf("gap height = %d, want 2", got)
	}
}

func TestPopupBox(t *testing.T) {
	box := PopupBox("x")
	if got := lipgloss.Width(box); got != 1+PopupFrameWidth {
		t.Errorf("box width = %d, want %d", got, 1+PopupFrameWidth)
	}
	if got := lipgloss.Height(box); got != 1+PopupFrameHeight {
		t.Errorf("box height = %d, want %d", got, 1+PopupFrameHeight)
	}
}

func TestPopupContentSize(t *testing.T) {
	cases := []struct {
		w, h, wantW, wantH, gotW, gotH int
	}{
		{120, 40, 50, 10, 50, 10},
		{40, 20, 50, 10, 34, 10},
		{10, 8, 50, 10, 4, 4},
		{2, 2, 50, 10, 1, 1},
	}
	for _, tc := range cases {
		w, h := PopupContentSize(tc.w, tc.h, tc.wantW, tc.wantH)
		if w != tc.gotW || h != tc.gotH {
			t.Errorf("PopupContentSize(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				tc.w, tc.h, tc.wantW, tc.wantH, w, h, tc.gotW, tc.gotH)
		}
	}
}
