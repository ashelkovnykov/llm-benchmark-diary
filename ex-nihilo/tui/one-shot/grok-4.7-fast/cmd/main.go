package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dbtui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Ensure db dir exists
	dbDir := "./db"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create db dir: %v\n", err)
		os.Exit(1)
	}

	schemaPath := filepath.Join("internal", "db", "schema.sql")
	// If running from bin/, adjust path
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = filepath.Join("..", "internal", "db", "schema.sql")
	}

	m := tui.New(dbDir, schemaPath)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
