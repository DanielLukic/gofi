package main

import (
	"flag"

	"gofi/pkg/log"
	"gofi/pkg/shared"
)

func main() {
	// Define command-line flags
	daemonFlag := flag.Bool("daemon", false, "Run as daemon")
	tuiFlag := flag.Bool("tui", false, "Run in TUI mode")
	kill := flag.Bool("kill", false, "Kill daemon")
	logLevel := flag.String("log", "info", "Set logging level (off, error, warning, info, debug)")

	flag.Parse()

	// Set up logger
	log.SetupLogger(*logLevel, *daemonFlag)

	log.Debug("Starting gofi")
	if *kill {
		log.Debug("Killing daemon")
		shared.KillDaemon()
	} else if *daemonFlag {
		log.Debug("Starting daemon")
		DaemonMain()
	} else {
		log.Debug("Starting client")
		ClientMain(*logLevel, *tuiFlag)
	}
}
