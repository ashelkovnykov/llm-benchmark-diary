package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"sqlnav/internal/ui"
)

func main() {
	if _, err := os.Stat("./db"); os.IsNotExist(err) {
		if err := os.MkdirAll("./db", 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating db directory: %v\n", err)
			os.Exit(1)
		}
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
