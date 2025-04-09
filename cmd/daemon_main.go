package main

import (
	"os"
	"os/signal"
	"syscall"

	"gofi/pkg/daemon"
	"gofi/pkg/ipc"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// CommandHandlerFunc wraps a simple function to implement ipc.CommandHandler
type CommandHandlerFunc func() string

// Handle implements the ipc.CommandHandler interface
func (f CommandHandlerFunc) Handle(command string) (string, error) {
	return f(), nil
}

// DaemonMain is the main entry point for the daemon
func DaemonMain() {
	// Clean up any existing daemon
	shared.KillDaemon()

	// Set up IPC socket
	listener, err := ipc.SetupSocket()
	if err != nil {
		log.Error("Failed to set up IPC socket")
		os.Exit(1)
	}
	defer listener.Close()

	// Initialize API and watcher
	api := daemon.NewAPI()
	watcher := daemon.NewWindowWatcher(nil, api)

	exitCode := 0
	commandHandlers := ipc.CommandHandlers{
		"HELLO": CommandHandlerFunc(daemon.HandleHello),
		"QUIT":  CommandHandlerFunc(daemon.HandleQuit),
	}

	// Handle ACTIVE_WINDOW_LIST separately since it needs the API
	commandHandlers["ACTIVE_WINDOW_LIST"] = CommandHandlerFunc(func() string {
		windows := api.ClientList()
		// Convert []*shared.Window to []shared.Window
		windowList := make([]shared.Window, len(windows))
		for i, w := range windows {
			windowList[i] = *w
		}
		return daemon.HandleActiveWindowList(windowList)
	})

	// Set up signal handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start watcher and process commands
	if !watcher.Start() {
		log.Error("Failed to start window watcher")
		os.Exit(1)
	}

	// Process commands until interrupted
	go ipc.ProcessCommands(commandHandlers)

	// Wait for interrupt or error
	<-interrupt
	log.Debug("Daemon stopped by interrupt")

	// Cleanup
	if !watcher.Stop() {
		log.Error("Failed to stop window watcher")
		exitCode = 1
	}

	if err := ipc.CleanupSocket(); err != nil {
		log.Error("Failed to clean up IPC socket")
		exitCode = 1
	}

	log.Info("Daemon stopped with exit code %d", exitCode)
	os.Exit(exitCode)
}
