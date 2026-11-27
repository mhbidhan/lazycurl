package keybind

import (
	"github.com/charmbracelet/bubbles/key"
)

type HelpPopupKeys struct {
	Escape key.Binding
}

func NewHelpPopupKeys() *HelpPopupKeys {
	return &HelpPopupKeys{
		Escape: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "cancel"),
		),
	}
}
