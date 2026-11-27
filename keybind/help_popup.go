package keybind

import (
	"github.com/charmbracelet/bubbles/key"
)

// HelpPopupKeys are the bindings of the scrollable keybinds popup.
type HelpPopupKeys struct {
	Escape key.Binding
}

// NewHelpPopupKeys builds the bindings for closing the keybinds popup.
func NewHelpPopupKeys() *HelpPopupKeys {
	return &HelpPopupKeys{
		Escape: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "cancel"),
		),
	}
}
