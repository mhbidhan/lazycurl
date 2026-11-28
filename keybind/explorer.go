package keybind

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// ExplorerKeys are the bindings for navigating and editing a collection tree.
type ExplorerKeys struct {
	helpModel help.Model

	CursorDown key.Binding
	CursorUp   key.Binding
	Collapse   key.Binding
	Expand     key.Binding
	Toggle     key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Search     key.Binding
	Submit     key.Binding
	Cancel     key.Binding
	AddChild   key.Binding
	AddSibling key.Binding
	Delete     key.Binding
	Rename     key.Binding
	Cut        key.Binding
	Copy       key.Binding
	Paste      key.Binding
	MoveDown   key.Binding
	MoveUp     key.Binding
	Help       key.Binding
	Back       key.Binding
}

// NewExplorerKeys builds the explorer bindings with their help labels.
func NewExplorerKeys() *ExplorerKeys {
	return &ExplorerKeys{
		helpModel: help.New(),

		CursorDown: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "cursor down"),
		),
		CursorUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "cursor up"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/left", "collapse"),
		),
		Expand: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/right", "expand"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle directory"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter nodes"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply filter"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
		AddChild: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add request"),
		),
		AddSibling: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "add directory"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete node"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename node"),
		),
		Cut: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "mark to move"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "mark to copy"),
		),
		Paste: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "paste marked"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "move down"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "move up"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "keybinds"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q", "ctrl+c"),
			key.WithHelp("esc/q", "back to home"),
		),
	}
}

// ShortHelp renders the hint line on the explorer screen.
func (k *ExplorerKeys) ShortHelp() string {
	return k.helpModel.ShortHelpView([]key.Binding{
		k.CursorDown, k.CursorUp, k.Toggle, k.AddChild, k.Cut, k.Copy, k.Paste, k.Help, k.Back,
	})
}

// GetKeybinds lists every explorer binding for the keybinds popup.
func (k *ExplorerKeys) GetKeybinds() []KeyInfo {
	return []KeyInfo{
		{Key: "j/down", Help: "cursor down"},
		{Key: "k/up", Help: "cursor up"},
		{Key: "h/left", Help: "collapse directory"},
		{Key: "l/right", Help: "expand directory"},
		{Key: "enter", Help: "toggle directory"},
		{Key: "g", Help: "jump to top"},
		{Key: "G", Help: "jump to bottom"},
		{Key: "/", Help: "filter nodes"},
		{Key: "a", Help: "add request (end name with / for directory)"},
		{Key: "A", Help: "add directory"},
		{Key: "d", Help: "delete node"},
		{Key: "r", Help: "rename node"},
		{Key: "x", Help: "mark node to move"},
		{Key: "c", Help: "mark node to copy"},
		{Key: "p", Help: "paste marked node (move or copy)"},
		{Key: "J", Help: "move node down"},
		{Key: "K", Help: "move node up"},
		{Key: "?", Help: "keybinds"},
		{Key: "esc/q", Help: "back to home"},
	}
}
