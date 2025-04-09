package main

import (
	"os"

	"gofi/pkg/client"
	"gofi/pkg/log"
)

// ClientMain is the main entry point for the client
// Args:
//
//	args: Command line arguments
func ClientMain(log_level string, tuiFlag bool) {
	// Kill existing gofi windows
	client.KillExistingGofiWindows(nil)

	// Get window list
	data := client.ActiveWindowList()
	if len(data) == 0 {
		ensureDaemonRunning(log_level)
		data = client.ActiveWindowList()
	}

	// Create windows and select
	client.SelectWindow(data, tuiFlag)
}

func ensureDaemonRunning(log_level string) {
	if client.IsDaemonRunning() {
		return
	}
	if !client.StartDaemon(log_level) {
		log.Error("Failed to start daemon")
		os.Exit(1)
	}
	log.Info("Daemon started")
}
