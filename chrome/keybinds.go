package chrome

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mhbidhan/lazycurl/theme"
)

// Keybind represents a keybinding to display.
type Keybind struct {
	Key  string
	Help string
}

// RenderKeybinds renders keybindings as a vertical list with the key in the
// accent color and the description in the default color.
func RenderKeybinds(bindings []Keybind) string {
	if len(bindings) == 0 {
		return ""
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent))
	descStyle := lipgloss.NewStyle()

	lines := make([]string, len(bindings))
	for i, kb := range bindings {
		lines[i] = keyStyle.Render(kb.Key) + ": " + descStyle.Render(kb.Help)
	}

	return strings.Join(lines, "\n")
}
