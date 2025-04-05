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
	daemonMode := flag.Bool("daemon", false, "Run as a daemon")
	restartMode := flag.Bool("restart", false, "Force restart daemon")
	killMode := flag.Bool("kill", false, "Kill daemon")
	flag.Parse()

	if *daemonMode {
		pkg.Log("Starting daemon...")
		pkg.RunDaemon()
		return
	}

	if *killMode {
		pkg.KillDaemon()
		return
	}

	if err := pkg.EnsureDaemon(*restartMode); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := pkg.HelloDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

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
