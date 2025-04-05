//go:build !gui
// +build !gui

package main

import (
	"flag"
	"fmt"
	"os"

	"gofi/pkg"
)

func main() {
	tuiMode := flag.Bool("tui", false, "Use terminal UI mode")
	flag.Parse()

	if err := pkg.CheckDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var err error

	if *tuiMode {
		err = pkg.TuiSelectWindow()
	} else {
		err = pkg.GuiSelectWindow()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting window: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
