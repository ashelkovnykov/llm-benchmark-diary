package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/db"
	"sqlnav/internal/ui"
)

func main() {
	store, err := db.NewStore("./db", "./internal/db/schema.sql")
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	p := tea.NewProgram(ui.New(store))
	if _, err := p.Run(); err != nil {
		log.Fatalf("tui: %v", err)
	}
}
