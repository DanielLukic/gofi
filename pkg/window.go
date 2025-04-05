package pkg

import (
	"sync"
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

// Global variables to store the active window list and mutex for thread safety
var (
	_active_windows  []Window
	_windows_mutex   sync.RWMutex
	_watcher_running bool
)

// start_window_watcher initializes and starts listening for active window change events
// and keeps the window list updated with the active window first
func start_window_watcher() error {
	// Don't start multiple watchers
	if _watcher_running {
		return nil
	}

	// Initialize X connection
	x, err := NewXUtil()
	if err != nil {
		return err
	}

	// Initialize the window list
	windows, err := ListWindows(x)
	if err != nil {
		x.Close()
		return err
	}

	// Get the initial active window
	active_window, err := get_active_window(x)
	if err == nil && active_window != "" {
		// Reorder windows with active window first
		windows = reorder_windows(windows, active_window)
	}

	// Set initial window list
	_windows_mutex.Lock()
	_active_windows = windows
	_windows_mutex.Unlock()

	// Mark as running before setting up events
	_watcher_running = true

	// Set up event-based monitoring for active window changes
	err = setup_active_window_events(x, func(active_window string) {
		// This callback is called whenever the active window changes
		if !_watcher_running {
			return
		}

		// Get the full window list
		new_windows, err := ListWindows(x)
		if err != nil || len(new_windows) == 0 {
			return
		}
		Log("Got window list with %d windows", len(new_windows))
		// log all windows by title:
		for _, w := range new_windows {
			Log("Window: %s", w.Title)
		}
		Log("Active window: %s", active_window)
		Log("Reordering windows...")

		// Merge the new windows with the existing list, preserving order
		_windows_mutex.Lock()
		_active_windows = merge_window_lists(_active_windows, new_windows, active_window)
		_windows_mutex.Unlock()
	})

	if err != nil {
		_watcher_running = false
		x.Close()
		return err
	}

	return nil
}

// reorder_windows takes a list of windows and moves the active window to the front
func reorder_windows(windows []Window, active_window_id string) []Window {
	if len(windows) <= 1 {
		return windows
	}

	// Find the active window
	activeIdx := -1
	for i, win := range windows {
		if win.ID == active_window_id {
			activeIdx = i
			break
		}
	}

	// If active window found, move it to the front
	if activeIdx > 0 {
		Log("Moving active window to front")
		// Create a new slice with active window first
		reordered := make([]Window, len(windows))
		reordered[0] = windows[activeIdx]

		// Copy the windows before the active window
		copy(reordered[1:activeIdx+1], windows[:activeIdx])

		// Copy the windows after the active window
		if activeIdx < len(windows)-1 {
			copy(reordered[activeIdx+1:], windows[activeIdx+1:])
		}

		return reordered
	}

	Log("Active window not found or already at front")

	// If active window not found or already at front, return original list
	return windows
}

// ActiveWindowList returns the current list of windows with the active window first
func ActiveWindowList() []Window {
	// Start the watcher if it's not already running
	if !_watcher_running {
		err := start_window_watcher()
		if err != nil {
			// If we can't start the watcher, fall back to regular window list
			x, err := NewXUtil()
			if err != nil {
				return []Window{}
			}
			defer x.Close()

			windows, err := ListWindows(x)
			if err != nil {
				return []Window{}
			}
			return windows
		}
	}

	// Return a copy of the current window list
	_windows_mutex.RLock()
	defer _windows_mutex.RUnlock()

	// Create a copy to avoid race conditions
	result := make([]Window, len(_active_windows))
	copy(result, _active_windows)

	return result
}
