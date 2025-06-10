package client

import (
	"os"
	"os/exec"

	"gofi/pkg/desktop"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// KillExistingGofiWindows finds and kills any existing gofi windows
// Args:
//
//	ours: List of window titles to kill
func KillExistingGofiWindows(ours []string) {
	if ours == nil {
		ours = []string{"gofi", "pofi", "rofi"}
	}

	wm := desktop.Instance()
	windows := wm.StackingList()

	// Find gofi windows
	var gofis []*shared.Window
	for _, window := range windows {
		if containsStr(ours, window.Title) && isStWindow(window.ClassName) {
			gofis = append(gofis, window)
		}
	}

	// Kill each gofi window
	for _, window := range gofis {
		log.Debug("Killing gofi window: %s", window.Title)

		if KillWindowWmctrl(*window) {
			continue
		}
		if KillWindowXkill(*window) {
			continue
		}
		log.Error("Failed to kill window: %s", window.Title)
	}
}

// KillWindow kills a window by its ID
// Args:
//
//	windowID: Window ID to kill
//
// Returns:
//
//	error: Error if any
func KillWindow(windowID string) error {
	cmd := exec.Command("xkill", "-id", windowID)
	return cmd.Run()
}

// KillStWindows kills all st terminal windows except the current one
// Returns:
//
//	error: Error if any
func KillStWindows() error {
	wm := desktop.Instance()
	windows := wm.StackingList()

	ours := []string{
		os.Getenv("WINDOWID"),
		os.Getenv("WINDOWID_OLD"),
	}

	for _, window := range windows {
		if containsStr(ours, window.Title) && isStWindow(window.ClassName) {
			if err := KillWindow(window.HexID()); err != nil {
				log.Error("Failed to kill window %s: %s", window.HexID(), err)
			}
		}
	}

	return nil
}

// isStWindow checks if a window is an st terminal window
// Args:
//
//	className: Window class name
//
// Returns:
//
//	bool: True if window is an st terminal window
func isStWindow(className string) bool {
	return className == "st-256color" || className == "st"
}
