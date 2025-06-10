package gofi2

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gofi/pkg/log"
	"gofi/pkg/shared"
)

func KillInstance() {
	tempDir := os.TempDir()
	lockPath := filepath.Join(tempDir, lockFileName)
	socketPath := filepath.Join(tempDir, socketFileName)

	if data, err := os.ReadFile(lockPath); err == nil {
		if pid, err := strconv.Atoi(string(data)); err == nil {
			log.Debug("Found gofi2 process with PID %d", pid)

			if _, err := os.FindProcess(pid); err == nil {
				if err := shared.KillProcess(pid, 100*time.Millisecond); err != nil {
					log.Error("Failed to kill gofi2 process %d: %s", pid, err)
				} else {
					log.Info("Killed gofi2 process %d", pid)
				}
			}
		}
	}

	os.Remove(lockPath)
	os.Remove(socketPath)

	log.Info("gofi2 instance cleanup completed")
}
