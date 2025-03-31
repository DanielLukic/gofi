package pkg

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
)

// TUISelector implements window selection using terminal UI
type TUISelector struct {
	manager Manager
	filter  Filter
}

// NewTUISelector creates a new TUISelector instance
func NewTUISelector(manager Manager, filter Filter) *TUISelector {
	return &TUISelector{
		manager: manager,
		filter:  filter,
	}
}

// SelectWindow opens a terminal UI for window selection
func (s *TUISelector) SelectWindow() (Window, error) {
	// Get windows and filter them
	windows := s.manager.GetWindows()
	var filteredWindows []Window
	for _, w := range windows {
		if s.filter.ShouldInclude(w) {
			filteredWindows = append(filteredWindows, w)
		}
	}
	if len(filteredWindows) == 0 {
		return Window{}, fmt.Errorf("no windows found")
	}

	// Sort windows by desktop
	currentDesktop, err := s.manager.GetCurrentDesktop()
	if err != nil {
		return Window{}, err
	}
	filteredWindows = SortWindows(filteredWindows, currentDesktop)

	// Prepare options for fuzzy finder
	maxTitleLen := 40
	maxClassLen := 40

	var options []string
	for _, w := range filteredWindows {
		title := truncateString(w.Title, maxTitleLen)
		class := truncateString(w.Name, maxClassLen) // Use full class name

		// Format TUI list (no command column)
		option := fmt.Sprintf("%d: %-*s %-*s",
			w.Desktop,
			maxTitleLen, title,
			maxClassLen, class)
		options = append(options, option)
	}

	// Run fuzzy finder
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
			window := filteredWindows[i]
			// Format desktop number in preview too
			return fmt.Sprintf("Window Details:\n\nDesktop: %d\nCommand: %s\nTitle: %s\nClass: %s",
				window.Desktop,
				window.Command,
				window.Title,
				window.Name)
		}),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return Window{}, nil
		}
		return Window{}, fmt.Errorf("error selecting window: %v", err)
	}

	return filteredWindows[idx], nil
}

func truncateString(str string, maxLen int) string {
	if len(str) > maxLen {
		return str[:maxLen]
	}
	return str
}
