package pkg

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Window represents a window in the window manager
type Window struct {
	ID      string
	Desktop int
	PID     string
	Command string
	Class   string // Short window class (e.g. thunderbird)
	Name    string // Full window class with instance (e.g. Mail.thunderbird)
	Title   string
}

// Manager interface defines methods for window management
type Manager interface {
	GetWindows() []Window
	FocusWindow(windowID string) error
	GetCurrentDesktop() (int, error)
}

// Filter interface defines methods for window filtering
type Filter interface {
	ShouldInclude(window Window) bool
}

// DefaultFilter implements basic window filtering
type DefaultFilter struct{}

// NewDefaultFilter creates a new DefaultFilter instance
func NewDefaultFilter() *DefaultFilter {
	return &DefaultFilter{}
}

// ShouldInclude implements window filtering logic
func (f *DefaultFilter) ShouldInclude(window Window) bool {
	// Filter out desktop and panel windows
	if (window.Command == "caja" && window.Name == "Desktop") || window.Command == "mate-panel" {
		return false
	}

	// Filter out our own gofi windows
	if window.Command == "st" && strings.Contains(window.Title, "gofi") {
		return false
	}

	// Filter out desktop -1 windows
	if window.Desktop == -1 {
		return false
	}
	return true
}

// WmctrlManager implements the Manager interface using wmctrl
type WmctrlManager struct{}

// NewWmctrlManager creates a new WmctrlManager instance
func NewWmctrlManager() *WmctrlManager {
	return &WmctrlManager{}
}

func (m *WmctrlManager) GetWindows() []Window {
	cmd := exec.Command("wmctrl", "-l", "-p", "-x")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var windows []Window
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		desktop, _ := strconv.Atoi(fields[1])
		title := strings.Join(fields[5:], " ")
		// Split window class into short and full versions
		classParts := strings.Split(fields[3], ".")
		shortClass := classParts[0]
		window := Window{
			ID:      fields[0],
			Desktop: desktop,
			PID:     fields[2],
			Class:   shortClass, // Short window class (e.g. thunderbird)
			Name:    fields[3],  // Full window class with instance (e.g. Mail.thunderbird)
			Title:   title,      // Window title (e.g. Tsunami IMAP - Mozilla Thunderbird)
		}

		// Get command name
		cmd := exec.Command("ps", "-p", window.PID, "-o", "comm=")
		if cmdOutput, err := cmd.Output(); err == nil {
			window.Command = strings.TrimSpace(string(cmdOutput))
		}

		windows = append(windows, window)
	}
	return windows
}

func (m *WmctrlManager) FocusWindow(windowID string) error {
	cmd := exec.Command("wmctrl", "-i", "-a", windowID)
	return cmd.Run()
}

func (m *WmctrlManager) GetCurrentDesktop() (int, error) {
	cmd := exec.Command("wmctrl", "-d")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "*" {
			desktop, err := strconv.Atoi(fields[0])
			if err != nil {
				return 0, err
			}
			return desktop, nil
		}
	}

	return 0, nil
}

// SortWindows sorts windows by desktop, prioritizing the current desktop
// and preserving wmctrl order within each desktop group
func SortWindows(windows []Window, currentDesktop int) []Window {
	// Create a map to store the original indices
	indexMap := make(map[Window]int)
	for i, w := range windows {
		indexMap[w] = i
	}

	// Sort windows
	sort.SliceStable(windows, func(i, j int) bool {
		// First, prioritize current desktop
		if windows[i].Desktop == currentDesktop && windows[j].Desktop != currentDesktop {
			return true
		}
		if windows[i].Desktop != currentDesktop && windows[j].Desktop == currentDesktop {
			return false
		}

		// Then sort by desktop number
		if windows[i].Desktop != windows[j].Desktop {
			return windows[i].Desktop < windows[j].Desktop
		}

		// Within same desktop, preserve original order
		return indexMap[windows[i]] < indexMap[windows[j]]
	})

	return windows
}
