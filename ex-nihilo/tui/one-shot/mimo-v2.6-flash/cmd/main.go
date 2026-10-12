package main

import (
	"fmt"
	"os"

	"sqlnav/internal/ui"
)

func main() {
	if err := ui.New().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "sqlnav: %v\n", err)
		os.Exit(1)
	}
}
