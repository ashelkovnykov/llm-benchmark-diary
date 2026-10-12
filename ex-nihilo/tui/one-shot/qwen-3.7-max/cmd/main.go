package main

import (
	"db-explorer/internal/db"
	"db-explorer/internal/tui"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := db.EnsureDBDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating db directory: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
