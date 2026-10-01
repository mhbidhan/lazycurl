package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mhbidhan/lazycurl/screen"
)

func main() {
	root := screen.NewRoot()
	p := tea.NewProgram(root, tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Printf("failed to run program: %v\n", err)
		os.Exit(1)
	}
}
