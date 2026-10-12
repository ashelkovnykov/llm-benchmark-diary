package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sqltui/internal/db"
	"sqltui/internal/ui"
)

func main() {
	if err := db.EnsureDir(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	m, err := ui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
