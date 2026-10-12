package main

import (
	"fmt"
	"os"

	"sqlnav/internal/db"
	"sqlnav/internal/tui"
)

func main() {
	if err := db.EnsureDir(); err != nil {
		fmt.Fprintln(os.Stderr, "sqlnav:", err)
		os.Exit(1)
	}

	if err := tui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "sqlnav:", err)
		os.Exit(1)
	}
}
