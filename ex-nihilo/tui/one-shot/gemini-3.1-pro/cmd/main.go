package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/db-tui/internal/tui"
)

func main() {
	if err := os.MkdirAll("./db", 0755); err != nil {
		fmt.Printf("Error creating db directory: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
