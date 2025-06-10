package daemon

import (
	"gofi/pkg/desktop"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

const (
	// Constants for logging format widths
	logTitleWidth = 40
)

// WindowList manages the current list and history of windows.
type WindowList struct {
	wm      desktop.WindowManager
	history *History // Maintains the ordered history and active window state
}

// NewWindowList creates a new WindowList instance.
// It requires a WindowManager and a History instance.
// If wm or history are nil, it uses default implementations.
func NewWindowList(wm desktop.WindowManager, history *History) *WindowList {
	if wm == nil {
		// Rely on desktop.Instance() to handle its own potential errors/nil return
		wm = desktop.Instance()
	}
	if history == nil {
		history = NewHistory()
	}

	return &WindowList{
		wm:      wm,
		history: history,
	}
}

// Initialize fetches the current window state and populates the history.
// It should be called once when the service starts.
func (wl *WindowList) Initialize() {
	initialWindows := wl.wm.StackingList()
	// Assume StackingList returns a valid slice (even if empty) or history handles nil
	wl.history.Initialize(initialWindows)

	activeID := wl.wm.ActiveWindowID()
	wl.history.UpdateActiveWindow(activeID)
	log.Debug("WindowList initialized.") // Simplified log message
}

// UpdateWindowList refreshes the window list based on the current state from the WindowManager.
// It updates the internal history and logs the list if changes occurred.
func (wl *WindowList) UpdateWindowList() {
	currentWindows := wl.wm.StackingList()
	activeID := wl.wm.ActiveWindowID()

	// Update history state
	listChanged := wl.history.KeepOnly(currentWindows)
	listChanged = wl.history.AddNew(currentWindows) || listChanged
	activeChanged := wl.history.UpdateActiveWindow(activeID)

	// Log if the list or active window changed
	if listChanged || activeChanged {
		wl.logWindowList(wl.history.windows)
	}
}

// ClientList prepares and returns the window list formatted for client consumption (e.g., Alt-Tab).
// It partitions windows by type ("Normal" vs. others) and swaps the first two for quick toggling.
// Returns nil if the history is empty.
func (wl *WindowList) ClientList() []*shared.Window {
	orderedWindows := wl.history.windows
	if len(orderedWindows) == 0 {
		return nil
	}

	// Partition and reorder for presentation
	presentedList := wl.partitionAndReorder(orderedWindows)

	// Perform Alt-Tab swap
	wl.applyAltTabSwap(presentedList)

	// We have to update all titles now
	for _, w := range presentedList {
		w.Title = wl.wm.WindowTitle(w.ID)
	}

	return presentedList
}

// partitionAndReorder separates windows into "Normal" and "Special" types,
// returning a new slice with "Normal" windows first.
func (wl *WindowList) partitionAndReorder(windows []*shared.Window) []*shared.Window {
	normalWindows := make([]*shared.Window, 0, len(windows))
	specialWindows := make([]*shared.Window, 0, len(windows))

	for _, w := range windows {
		// Assume w is not nil and w.Type is valid, as per original code's implicit assumption
		if w.Type == "Normal" {
			normalWindows = append(normalWindows, w)
		} else { // Treat non-Normal as Special
			specialWindows = append(specialWindows, w)
		}
	}
	return append(normalWindows, specialWindows...)
}

// applyAltTabSwap swaps the first two elements of the slice if it has at least two elements.
func (wl *WindowList) applyAltTabSwap(windows []*shared.Window) {
	if len(windows) >= 2 {
		windows[0], windows[1] = windows[1], windows[0]
	}
}

// logWindowList formats and logs the provided window list for debugging.
func (wl *WindowList) logWindowList(windows []*shared.Window) {
	log.Debug("Window state changed. Current list (%d windows):", len(windows))
	if len(windows) == 0 {
		log.Debug("No windows found")
		return
	}
	
	// Log each window's details
	for i, w := range windows {
		title := w.Title
		if len(title) > logTitleWidth {
			title = title[:logTitleWidth] // Truncate title
		}
		log.Debug("%2d: %10d %s %-7s %s (%d)",
			i,
			w.ID,
			w.DesktopStr(),
			w.Type,
			title,
			w.PID,
		)
	}
	log.Debug("----------------------------------------")
}
