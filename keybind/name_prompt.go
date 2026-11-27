package keybind

import (
	"github.com/charmbracelet/bubbles/key"
)

// NamePromptKeys drives the single-field popup used for both creating and
// renaming a collection.
type NamePromptKeys struct {
	Submit key.Binding
	Cancel key.Binding
}

// NewNamePromptKeys builds the submit and cancel bindings of a name prompt.
func NewNamePromptKeys() *NamePromptKeys {
	return &NamePromptKeys{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "cancel"),
		),
	}
}
