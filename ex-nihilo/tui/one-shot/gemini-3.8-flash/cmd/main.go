package main

import (
	"fmt"
	"os"

	"sqlnav/internal/db"
	"sqlnav/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := db.EnsureDBDir(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize db directory: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}
