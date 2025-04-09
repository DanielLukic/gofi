package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RequiredDependencies lists required dependencies
var RequiredDependencies = []string{
	"xkill",
	"st",
	"wmctrl",
	"fzf",
}

// CheckDependencies verifies that all required dependencies are installed
// Returns:
//
//	error: nil if all dependencies are present, error otherwise
func CheckDependencies() error {
	for _, cmd := range RequiredDependencies {
		if !isCommandAvailable(cmd) {
			return fmt.Errorf("%s is not installed", cmd)
		}
	}

	// Check for X11 libraries
	if !isX11Available() {
		return fmt.Errorf("python-xlib is not installed")
	}

	return nil
}

// isCommandAvailable checks if a command is available in PATH
// Args:
//
//	cmd: Command to check
//
// Returns:
//
//	bool: True if command is available
func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// isX11Available checks if X11 libraries are available
// Returns:
//
//	bool: True if X11 libraries are available
func isX11Available() bool {
	// Check if X11 includes are available
	_, err := os.Stat(filepath.Join("/usr", "include", "X11", "Xlib.h"))
	if err != nil {
		return false
	}

	// Check if X11 libraries are available
	_, err = os.Stat(filepath.Join("/usr", "lib", "x86_64-linux-gnu", "libX11.so"))
	if err != nil {
		return false
	}

	return true
}
