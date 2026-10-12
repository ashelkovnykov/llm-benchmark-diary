package main

import (
	"dbviewer/internal/db"
	"dbviewer/internal/tui"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
)

var selectedDBName string

func main() {
	databases, err := db.ListDatabases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing databases: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(tui.NewDBList(databases, handleSelect, handleCreate))
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if selectedDBName == "" {
		fmt.Println("No database selected. Exiting.")
		os.Exit(0)
	}

	viewer, err := tui.NewDBViewer(selectedDBName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading database: %v\n", err)
		os.Exit(1)
	}

	program = tea.NewProgram(viewer)
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running viewer: %v\n", err)
		os.Exit(1)
	}
}

func handleSelect(name string) {
	selectedDBName = name
}

func handleCreate(name string) error {
	newDB, err := db.CreateDatabase(name, db.GetSchema())
	if err != nil {
		return err
	}
	defer newDB.Close()

	if err := db.InsertDummyRecords(newDB, 15); err != nil {
		return err
	}

	selectedDBName = name
	return nil
}
