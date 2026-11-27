package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// ConfirmKeys are the yes/no bindings of a confirmation popup.
type ConfirmKeys struct {
	helpModel help.Model
	Confirm   key.Binding
	Cancel    key.Binding
}

// NewConfirmKeys builds the confirmation bindings with their help labels.
func NewConfirmKeys() *ConfirmKeys {
	return &ConfirmKeys{
		helpModel: help.New(),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "esc", "ctrl+c"),
			key.WithHelp("n/esc", "cancel"),
		),
	}
}

// ShortHelp renders the hint line under a confirmation popup.
func (k *ConfirmKeys) ShortHelp() string {
	return k.helpModel.ShortHelpView([]key.Binding{k.Confirm, k.Cancel})
}
