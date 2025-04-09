package shared

import (
	"os"
	"time"

	"gofi/pkg/ipc"
	"gofi/pkg/log"
)

// KillDaemon kills the daemon process
// This function:
// 1. Kills all remaining daemon processes
// 2. Cleans up socket
// 3. Removes PID file
func KillDaemon() error {
	if err := killAllDaemons(); err != nil {
		log.Error("Failed to kill all daemons: %s", err)
	}
	if err := ipc.CleanupSocket(); err != nil {
		log.Error("Failed to cleanup socket: %s", err)
	}
	return nil
}

// killAllDaemons kills all daemon processes
// Returns:
//
//	error: Any error that occurred
func killAllDaemons() error {
	log.Debug("Looking for daemon processes")

	// Find all processes matching 'gofi -daemon'
	processes, err := FindGofiDaemons()
	if err != nil {
		log.Error("Failed to find daemon processes: %s", err)
		return err
	}

	for _, proc := range processes {
		if proc.PID == os.Getpid() {
			continue
		}

		log.Debug("Killing daemon process %d", proc.PID)
		if err := KillProcess(proc.PID, 100*time.Millisecond); err != nil {
			log.Error("Failed to kill daemon process %d: %s", proc.PID, err)
		}
	}

	return nil
}
