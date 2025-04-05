package pkg

import (
	"fmt"
	"os/exec"
	"strings"
)

func CheckDependency(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("required dependency '%s' not found in PATH", name)
	}
	return nil
}

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

func CheckDependencies() error {
	dependencies := []string{"wmctrl", "fzf", "st", "xkill"}

	var missing []string
	for _, dep := range dependencies {
		if err := CheckDependency(dep); err != nil {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		return &DependencyError{Missing: missing}
	}
	return nil
}
