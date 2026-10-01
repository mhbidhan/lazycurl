package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type RootKeys struct {
	help help.Model

	Keybind key.Binding
	Quit    key.Binding
}

func NewRootKeys() *RootKeys {
	return &RootKeys{
		help: help.New(),
		Keybind: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "keybinds"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
	}
}

func (k RootKeys) ShortHelp() string {
	return k.help.ShortHelpView([]key.Binding{k.Keybind, k.Quit})
}

func (k RootKeys) FullHelp() string {
	return k.help.FullHelpView([][]key.Binding{{k.Keybind, k.Quit}})
}
