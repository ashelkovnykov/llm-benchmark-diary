// ./cmd/main.go
package main

import (
	"fmt"
	"os"

	dbpkg "dbnav/internal/db"
	"dbnav/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := dbpkg.EnsureDBDir(); err != nil {
		fmt.Fprintf(os.Stderr, "ensure db dir: %v\n", err)
		os.Exit(1)
	}
	dbs, err := dbpkg.ListDBs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list dbs: %v\n", err)
		os.Exit(1)
	}

	m := tui.New(dbs)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
