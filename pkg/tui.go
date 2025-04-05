package pkg

import (
	"fmt"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
)

func TuiSelectWindow() error {
	x, err := NewXUtil()
	if err != nil {
		return fmt.Errorf("failed to connect to X server: %v", err)
	}
	defer x.Close()

	// Get windows using X11 direct implementation
	windows, _ := ListWindows(x)

	options := _prepare_list(windows)
	idx, err := _select_window(options, windows)
	if err == fuzzyfinder.ErrAbort {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error selecting window: %v", err)
	}
	err = _activate(windows[idx].ID)
	if err != nil {
		return fmt.Errorf("error focusing window: %v", err)
	}

	return nil
}

func _activate(windowID string) error {
	cmd := exec.Command("wmctrl", "-i", "-a", windowID)
	return cmd.Run()
}

func _prepare_list(windows []Window) []string {
	maxTitleLen := 40
	maxClassLen := 40

	var options []string
	for _, w := range windows {
		title := _truncate(w.Title, maxTitleLen)
		class := _truncate(w.Name, maxClassLen) // Use full class name

		// Format TUI list (no command column)
		option := fmt.Sprintf("%d: %-*s %-*s",
			w.Desktop,
			maxTitleLen, title,
			maxClassLen, class)
		options = append(options, option)
	}
	return options
}

func _select_window(options []string, windows []Window) (int, error) {
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i]
		},
		fuzzyfinder.WithPromptString("> "),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			window := windows[i]
			// Format desktop number in preview too
			return fmt.Sprintf("Window Details:\n\nDesktop: %d\nCommand: %s\nTitle: %s\nClass: %s",
				window.Desktop,
				window.Command,
				window.Title,
				window.Name)
		}),
	)
	return idx, err
}
