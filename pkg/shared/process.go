package shared

import (
	"fmt"
	"syscall"
	"time"

	"gofi/pkg/log"
)

func KillProcess(pid int, wait time.Duration) error {
	log.Debug("Killing process with PID %d", pid)

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		log.Error("Failed to send SIGTERM to PID %d: %s", pid, err)
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	time.Sleep(wait)

	if err := syscall.Kill(pid, 0); err == nil {
		log.Debug("Killing process with PID %d with SIGKILL", pid)
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
			log.Error("Failed to send SIGKILL to PID %d: %s", pid, err)
			return fmt.Errorf("failed to send SIGKILL: %w", err)
		}
	}

	time.Sleep(100 * time.Millisecond)
	if err := syscall.Kill(pid, 0); err == nil {
		return fmt.Errorf("process %d still running after SIGKILL", pid)
	}

	log.Debug("Successfully killed process %d", pid)
	return nil
}
