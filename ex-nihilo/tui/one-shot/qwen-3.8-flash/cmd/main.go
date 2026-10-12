package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/example/sql-tui/internal/config"
	"github.com/example/sql-tui/internal/db"
	"github.com/example/sql-tui/internal/tui"
)

func main() {
	cfg := config.Load()

	if err := db.EnsureDir(cfg.DBDir); err != nil {
		fmt.Fprintln(os.Stderr, "failed to prepare database directory:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(tui.New(cfg), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "TUI error:", err)
		os.Exit(1)
	}
}
