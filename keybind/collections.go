package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// CollectionKeys are the bindings for browsing collections and their flows.
type CollectionKeys struct {
	helpModel  help.Model
	CursorDown key.Binding
	CursorUp   key.Binding
	New        key.Binding
	Search     key.Binding
	Delete     key.Binding
	Edit       key.Binding
	Escape     key.Binding
	Submit     key.Binding
}

// NewCollectionKeys builds the collections bindings with their help labels.
func NewCollectionKeys() *CollectionKeys {
	return &CollectionKeys{
		helpModel: help.New(),
		CursorDown: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "cursor down"),
		),
		CursorUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "cursor up"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "cancel"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new collection"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search collection"),
		),
		Delete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "delete collection"),
		),
		Edit: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "rename collection"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply filter"),
		),
	}
}

// ShortHelp renders the hint line under the collections popup.
func (k *CollectionKeys) ShortHelp() string {
	return k.helpModel.ShortHelpView([]key.Binding{k.CursorDown, k.CursorUp, k.Escape})
}

// GetKeybinds lists every collections binding for the keybinds popup.
func (k *CollectionKeys) GetKeybinds() []KeyInfo {
	return []KeyInfo{
		{Key: "j/down", Help: "cursor down"},
		{Key: "k/up", Help: "cursor up"},
		{Key: "esc/ctrl+c", Help: "cancel"},
		{Key: "n", Help: "new collection"},
		{Key: "/", Help: "search collection"},
		{Key: "ctrl+d", Help: "delete collection"},
		{Key: "ctrl+e", Help: "rename collection"},
		{Key: "enter", Help: "apply filter"},
	}
}
