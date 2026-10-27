package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"sql-navigator/internal/ui"
)

func main() {
	m, err := ui.NewModel("./db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
