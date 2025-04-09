package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"gofi/pkg/ipc"
	"gofi/pkg/log"
)

// IsDaemonRunning checks if the daemon is running
// Returns:
//
//	bool: True if daemon is running
func IsDaemonRunning() bool {
	if !ipc.CheckSocketExists() {
		return false
	}
	return SendHello()
}

// StartDaemon starts the daemon process
// Args:
//
//	logLevel: Optional log level
//
// Returns:
//
//	bool: True if daemon started successfully
func StartDaemon(logLevel string) bool {
	if IsDaemonRunning() {
		err := fmt.Errorf("daemon is already running")
		log.Error(err.Error())
		return false
	}

	// Get the executable path reliably
	exePath, err := os.Executable()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get executable path: %s", err))
		return false
	}

	// Resolve the absolute path
	exeAbsPath, err := filepath.Abs(exePath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get absolute executable path: %s", err))
		return false
	}

	// Construct arguments for the daemon process
	args := []string{exeAbsPath} // Use absolute path
	args = append(args, "-daemon")

	// Add log level if not default
	if logLevel != "info" {
		args = append(args, "-log", logLevel)
	}

	log.Info(fmt.Sprintf("Daemon arguments: %v", args))

	// Open the output log file
	logFile, err := os.OpenFile("/tmp/gofi_spawn.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to open daemon log file: %s", err))
		return false
	}
	defer logFile.Close() // Ensure the file is closed in the parent process

	// Start daemon process
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logFile // Redirect stdout to file
	cmd.Stderr = logFile // Redirect stderr to file
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		// Setsid:  true, // Temporarily removed for testing EPERM
	}

	if err := cmd.Start(); err != nil {
		log.Error(fmt.Sprintf("Failed to start daemon process: %s", err))
		return false
	}

	return awaitDaemon()
}

// awaitDaemon waits for daemon to start
// Returns:
//
//	bool: True if daemon started successfully
func awaitDaemon() bool {
	maxAttempts := 5
	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(200 * time.Millisecond)

		if IsDaemonRunning() {
			log.Debug("Daemon started successfully")
			return true
		}

		progress := fmt.Sprintf("%d/%d", attempt+1, maxAttempts)
		log.Debug(fmt.Sprintf("Waiting for daemon to start (%s)", progress))
	}

	return false
}
