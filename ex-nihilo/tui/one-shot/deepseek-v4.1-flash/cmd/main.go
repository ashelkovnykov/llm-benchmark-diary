package main

import (
	"fmt"
	"os"

	"dbnav/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model, err := tui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "dbnav:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "dbnav:", err)
		os.Exit(1)
	}
}
