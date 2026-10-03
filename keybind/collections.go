package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type CollectionKeys struct {
	helpModel help.Model
	Keybind   key.Binding
	ScrollUp  key.Binding
	Quit      key.Binding
}

func NewCollectionKeys() *CollectionKeys {
	return &CollectionKeys{
		helpModel: help.New(),
		Keybind: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "keybinds"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k", "scroll up"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
	}
}

func (k *CollectionKeys) ShortHelp() string {
	return k.helpModel.ShortHelpView([]key.Binding{k.Keybind, k.Quit})
}

func (k *CollectionKeys) GetKeybinds() []KeyInfo {
	return []KeyInfo{
		{Key: "?", Help: "keybinds"},
		{Key: "c", Help: "collections"},
		{Key: "q", Help: "quit"},
	}
}
