package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type HomeKeys struct {
	helpModel   help.Model
	Keybind     key.Binding
	Collections key.Binding
	Quit        key.Binding
}

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

func (k *HomeKeys) ShortHelp() string {
	return k.helpModel.ShortHelpView([]key.Binding{k.Keybind, k.Quit})
}

func (k *HomeKeys) GetKeybinds() []KeyInfo {
	return []KeyInfo{
		{Key: "?", Help: "keybinds"},
		{Key: "c", Help: "collections"},
		{Key: "q", Help: "quit"},
	}
}
