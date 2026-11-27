package ui

import "github.com/charmbracelet/lipgloss"

// GapVertical returns a blank run of n rows.
func GapVertical(n int) string {
	return lipgloss.NewStyle().Height(n).Render("")
}
