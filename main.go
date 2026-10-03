package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/mhbidhan/lazycurl/screen"
)

func main() {
	root := screen.NewRoot()

	// INFO: Added for auto build and view in dev mode
	// to be removed latter
	w, h, _ := term.GetSize(os.Stdout.Fd())
	d := screen.NewHomeScreen(w, h)
	fmt.Println(d.View())
	// END

	p := tea.NewProgram(root, tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Printf("failed to run program: %v\n", err)
		os.Exit(1)
	}
}
