package client

import (
	"fmt"
	"os/exec"

	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// KillWindowWmctrl kills a window using wmctrl
// Args:
//
//	window: Window to kill
//
// Returns:
//
//	bool: True if window was killed successfully
func KillWindowWmctrl(window shared.Window) bool {
	log.Debug(fmt.Sprintf("Attempting to kill window %d", window.ID))

	cmd := exec.Command("wmctrl", "-ic", fmt.Sprintf("%d", window.ID))
	if err := cmd.Run(); err != nil {
		log.Error(fmt.Sprintf("Failed to kill window %d: %s", window.ID, err))
		return false
	}

	log.Debug(fmt.Sprintf("Successfully killed window %d", window.ID))
	return true
}

// KillWindowXkill kills a window using xkill
// Args:
//
//	window: Window to kill
//
// Returns:
//
//	bool: True if window was killed successfully
func KillWindowXkill(window shared.Window) bool {
	log.Debug(fmt.Sprintf("Attempting to kill window %d", window.ID))

	cmd := exec.Command("xkill", "-id", fmt.Sprintf("%d", window.ID))
	if err := cmd.Run(); err != nil {
		log.Error(fmt.Sprintf("Failed to kill window %d: %s", window.ID, err))
		return false
	}

	log.Debug(fmt.Sprintf("Successfully killed window %d", window.ID))
	return true
}
