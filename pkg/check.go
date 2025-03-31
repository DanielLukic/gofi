package pkg

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckDependency checks if a command is available in PATH
func CheckDependency(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("required dependency '%s' not found in PATH", name)
	}
	return nil
}

// DependencyError represents a missing dependency error
type DependencyError struct {
	Missing []string
}

func (e *DependencyError) Error() string {
	if len(e.Missing) == 0 {
		return "no missing dependencies"
	}

	if len(e.Missing) == 1 {
		return fmt.Sprintf("missing required dependency: %s", e.Missing[0])
	}

	return fmt.Sprintf("missing required dependencies: %s", strings.Join(e.Missing, ", "))
}

// CheckDependencies checks if all required dependencies are available
func CheckDependencies(terminal string) error {
	var missing []string
	// Always required
	if err := CheckDependency("wmctrl"); err != nil {
		missing = append(missing, "wmctrl")
	}

	if err := CheckDependency("fzf"); err != nil {
		missing = append(missing, "fzf")
	}

	// Check terminal based on mode
	if err := CheckDependency(terminal); err != nil {
		missing = append(missing, terminal)
	}

	if len(missing) > 0 {
		return &DependencyError{Missing: missing}
	}

	return nil
}
