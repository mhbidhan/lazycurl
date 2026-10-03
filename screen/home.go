package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/chrome"
	"github.com/mhbidhan/lazycurl/keybind"
	"github.com/mhbidhan/lazycurl/ui"
)

type PopupID string

const (
	PopupHelp        PopupID = "help"
	PopupCollections PopupID = "collections"
)

type HomeScreen struct {
	width  int
	height int

	keybind *keybind.HomeKeys
	pm      *chrome.PopupManager
}

func NewHomeScreen(w, h int) *HomeScreen {
	kb := keybind.NewHomeKeys()
	pm := chrome.NewPopupManager()

	bindings := make([]chrome.Keybind, len(kb.GetKeybinds()))
	for i, b := range kb.GetKeybinds() {
		bindings[i] = chrome.Keybind{Key: b.Key, Help: b.Help}
	}

	pm.Register(string(PopupHelp), chrome.NewHelpPopup(bindings, 30, 8))
	pm.Register(string(PopupCollections), chrome.NewCollectionsPopup(30, 8))

	return &HomeScreen{
		width:   w,
		height:  h,
		keybind: kb,
		pm:      pm,
	}
}

func (s *HomeScreen) SetSize(w, h int) {
	s.width = w
	s.height = h
	popupWidth := min(40, w-4)
	popupHeight := min(10, h-6)
	if popupWidth < 20 {
		popupWidth = 20
	}
	if popupHeight < 5 {
		popupHeight = 5
	}
	s.pm.SetSize(popupWidth, popupHeight)
}

func (s *HomeScreen) Init() tea.Cmd {
	return nil
}

func (s *HomeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, s.keybind.Quit):
			if s.pm.IsOpen() {
				s.pm.CloseActive()
				return s, nil
			}
			return s, tea.Quit
		case key.Matches(keyMsg, s.keybind.Keybind):
			s.pm.Open(string(PopupHelp))
			return s, nil
		case key.Matches(keyMsg, s.keybind.Collections):
			s.pm.Open(string(PopupCollections))
			return s, nil
		}
	}

	if s.pm.IsOpen() {
		return s, s.pm.Update(msg)
	}
	return s, nil
}

func (s *HomeScreen) View() string {
	if s.pm.IsOpen() {
		_, m := s.pm.Active()
		if m != nil {
			return lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center).
				Width(s.width).
				Height(s.height).
				Render(m.View())
		}
	}

	body := s.keybind.ShortHelp()
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		ui.Logo,
		ui.GapVertical(1),
		body,
	)
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(s.width).
		Height(s.height).
		Render(content)
}
