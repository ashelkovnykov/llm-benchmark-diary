// cmd/main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sqltui/internal/tui"
)

func main() {
	if err := os.MkdirAll("./db", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create db directory: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
