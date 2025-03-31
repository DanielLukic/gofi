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
	// Parse command line flags
	tuiMode := flag.Bool("tui", false, "Use terminal UI mode")
	terminal := flag.String("terminal", "st", "Terminal emulator to use (st or xterm)")
	flag.Parse()

	// Validate terminal option
	if *terminal != "st" && *terminal != "xterm" {
		fmt.Fprintf(os.Stderr, "Error: --terminal must be either 'st' or 'xterm'\n")
		os.Exit(1)
	}

	// Check dependencies
	if err := pkg.CheckDependencies(*terminal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create window manager and filter
	manager := pkg.NewWmctrlManager()
	filter := pkg.NewDefaultFilter()

	// Select window based on mode
	var selectedWindow pkg.Window
	var err error

	if *tuiMode {
		// Use terminal UI mode
		selector := pkg.NewTUISelector(manager, filter)
		selectedWindow, err = selector.SelectWindow()
	} else {
		// Use FZF mode (default)
		selector := pkg.NewFzfSelector(manager, filter)
		selector.SetTerminal(*terminal)
		if err := selector.SelectWindow(); err != nil {
			fmt.Fprintf(os.Stderr, "Error launching window selector: %v\n", err)
			os.Exit(1)
		}
		// Exit successfully - the FZF script will handle window focusing
		os.Exit(0)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting window: %v\n", err)
		os.Exit(1)
	}

	// Focus the selected window
	if err := manager.FocusWindow(selectedWindow.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error focusing window: %v\n", err)
		os.Exit(1)
	}
}
