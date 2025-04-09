package shared

import (
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"

	"gofi/pkg/log"
)

// Process represents a running process
type Process struct {
	PID     int
	Name    string
	CmdLine []string
}

// FindGofiDaemons finds all gofi daemons
//
// Returns:
//
//	[]Process: List of matching processes
//	error: Any error that occurred
func FindGofiDaemons() ([]Process, error) {
	procs, err := process.Processes() // Get all processes
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	var processes []Process

	// Get list of PIDs
	for _, proc := range procs {
		pid := proc.Pid
		name, _ := proc.Name()
		if !strings.Contains(name, "gofi") {
			continue
		}

		cmdline, _ := proc.Cmdline()
		log.Debug("Process: %d, %s, %s", pid, name, cmdline)

		if !strings.Contains(cmdline, "-daemon") {
			continue
		}

		processes = append(processes, Process{
			PID:     int(pid),
			Name:    name,
			CmdLine: strings.Fields(cmdline),
		})
	}

	return processes, nil
}

// KillProcess kills a process by PID
// Args:
//
//	pid: Process ID to kill
//	wait: Time to wait before checking if process is gone
//
// Returns:
//
//	error: Any error that occurred
func KillProcess(pid int, wait time.Duration) error {
	log.Debug("Killing process with PID %d", pid)

	// Send SIGTERM first
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		log.Error("Failed to send SIGTERM to PID %d: %s", pid, err)
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait a bit to see if process terminates
	time.Sleep(wait)

	// Check if process is still alive
	if err := syscall.Kill(pid, 0); err == nil {
		// Process is still alive, send SIGKILL
		log.Debug("Killing process with PID %d with SIGKILL", pid)
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
			log.Error("Failed to send SIGKILL to PID %d: %s", pid, err)
			return fmt.Errorf("failed to send SIGKILL: %w", err)
		}
	}

	// Wait for process to be completely gone
	time.Sleep(100 * time.Millisecond)
	if err := syscall.Kill(pid, 0); err == nil {
		return fmt.Errorf("process %d still running after SIGKILL", pid)
	}

	log.Debug("Successfully killed process %d", pid)
	return nil
}
