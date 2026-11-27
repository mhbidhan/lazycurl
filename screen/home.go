package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
	"gorm.io/gorm"
)

// HomeScreen renders the landing view and owns the top-level popups.
type HomeScreen struct {
	width  int
	height int

	keybind *keybind.HomeKeys
	pm      *chrome.PopupManager
}

// How big a popup is allowed to get. Both are caps rather than sizes: a popup
// shrinks below them whenever the terminal is too small to hold them.
const (
	popupMaxWidth  = 40
	popupMaxHeight = 12
)

// NewHomeScreen builds the home screen and registers its popups.
func NewHomeScreen(db *gorm.DB, w, h int) *HomeScreen {
	kb := keybind.NewHomeKeys()
	pm := chrome.NewPopupManager()

	pm.Register(chrome.PopupHelp, chrome.NewHelpPopup(kb.GetKeybinds(), 30, 8))
	pm.Register(chrome.PopupCollections, chrome.NewCollectionsPopup(db, 30, 8))

	s := &HomeScreen{
		keybind: kb,
		pm:      pm,
	}
	s.SetSize(w, h)

	return s
}

// SetSize records the container and hands every popup the largest content area
// that still fits inside it. This is the only place popups are sized: the
// window-size handler below goes through here, so a resize and a direct call
// cannot disagree.
func (s *HomeScreen) SetSize(w, h int) {
	s.width = w
	s.height = h
	contentW, contentH := ui.PopupContentSize(w, h, popupMaxWidth, popupMaxHeight)
	s.pm.SetSize(contentW, contentH)
}

func (s *HomeScreen) Init() tea.Cmd {
	return nil
}

func (s *HomeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Screen-owned messages first: they must be handled even while a popup is
	// open, otherwise the popup can never ask to be closed.
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		s.SetSize(size.Width, size.Height)
	}

	if _, ok := msg.(chrome.DismissPopup); ok {
		if id, m := s.pm.Active(); m != nil {
			s.pm.Close(id)
		}
		return s, nil
	}

	// A popup owns every other key while it is open, so the bindings below
	// never fire underneath it.
	if s.pm.IsOpen() {
		return s, s.pm.Update(msg)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, s.keybind.Keybind):
			s.pm.Open(chrome.PopupHelp)
			return s, nil
		case key.Matches(keyMsg, s.keybind.Collections):
			s.pm.Open(chrome.PopupCollections)
			return s, nil
		case key.Matches(keyMsg, s.keybind.Quit):
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s *HomeScreen) View() string {
	if _, m := s.pm.Active(); m != nil {
		return overlay(s.width, s.height, m.View())
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		ui.Logo,
		ui.GapVertical(1),
		s.keybind.ShortHelp(),
	)
	return overlay(s.width, s.height, content)
}

// overlay centers content in the full screen area. Popups return their own
// bordered box, so the screen is what aligns them.
//
// The content is cropped before it is centered. Width alone would wrap anything
// larger than the container instead, and wrapping tears a box frame apart rather
// than clipping it, so a popup one column too wide would shred instead of just
// being cut off.
func overlay(w, h int, content string) string {
	cropped := lipgloss.NewStyle().MaxWidth(w).MaxHeight(h).Render(content)

	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(w).
		Height(h).
		Render(cropped)
}
