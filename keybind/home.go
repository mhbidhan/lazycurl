package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// HomeKeys are the bindings available when no popup is open.
type HomeKeys struct {
	helpModel   help.Model
	Keybind     key.Binding
	Collections key.Binding
	Quit        key.Binding
}

// NewHomeKeys builds the home screen bindings with their help labels.
func NewHomeKeys() *HomeKeys {
	return &HomeKeys{
		helpModel: help.New(),
		Keybind: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "keybinds"),
		),
		Collections: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "collections"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
	}
}

// ShortHelp renders the bindings shown on the home screen itself.
func (k *HomeKeys) ShortHelp() string {
	return k.helpModel.ShortHelpView([]key.Binding{k.Keybind, k.Quit})
}

// GetKeybinds lists every home binding for the keybinds popup.
func (k *HomeKeys) GetKeybinds() []KeyInfo {
	return []KeyInfo{
		{Key: "?", Help: "keybinds"},
		{Key: "c", Help: "collections"},
		{Key: "q/ctrl+c", Help: "quit"},
	}
}
