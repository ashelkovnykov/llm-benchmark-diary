package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"sql-tui/internal/db"
	"sql-tui/internal/tui"
)

func main() {
	// ensure db directory exists
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatalf("could not create db directory: %v", err)
	}

	entries, err := db.ListDBFiles("db")
	if err != nil {
		log.Fatalf("could not list db files: %v", err)
	}

	model := tui.NewModel("db", entries)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	// clean up: remove empty path? not needed
}
