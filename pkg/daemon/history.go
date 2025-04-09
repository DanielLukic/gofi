package daemon

import (
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// History manages window history
type History struct {
	windows []*shared.Window
}

// NewHistory creates a new History instance
// Returns:
//
//	*History: New history instance
func NewHistory() *History {
	return &History{
		windows: make([]*shared.Window, 0),
	}
}

// Initialize initializes the history with a list of windows
// Args:
//
//	windows: List of windows to initialize with
func (h *History) Initialize(windows []*shared.Window) {
	h.windows = windows
}

// KeepOnly keeps only the specified windows in history
// Args:
//
//	windows: List of windows to keep
//
// Returns:
//
//	bool: True if the history was modified.
func (h *History) KeepOnly(windows []*shared.Window) bool {
	// Create a set of IDs to keep for efficient lookup
	keepIDs := make(map[int]struct{}, len(windows))
	for _, w := range windows {
		keepIDs[w.ID] = struct{}{}
	}

	// Filter the existing history
	keptWindows := make([]*shared.Window, 0, len(h.windows)) // Pre-allocate capacity
	for _, histWindow := range h.windows {
		if _, shouldKeep := keepIDs[histWindow.ID]; shouldKeep {
			keptWindows = append(keptWindows, histWindow)
		}
	}

	// Check if the list actually changed (length or content)
	changed := len(h.windows) != len(keptWindows)
	if !changed {
		// If length is same, check if elements are the same (order matters here)
		for i := range keptWindows {
			if h.windows[i].ID != keptWindows[i].ID {
				changed = true
				break
			}
		}
	}

	// Replace the old history with the filtered list
	h.windows = keptWindows
	return changed
}

// AddNew adds new windows to history
// Args:
//
//	windows: List of windows to add
//
// Returns:
//
//	bool: True if any new windows were added.
func (h *History) AddNew(windows []*shared.Window) bool {
	changed := false
	// Create a map of existing window IDs
	existingIDs := make(map[int]struct{})
	for _, w := range h.windows {
		existingIDs[w.ID] = struct{}{}
	}

	// Add only new windows
	for _, w := range windows {
		if _, exists := existingIDs[w.ID]; !exists {
			h.windows = append(h.windows, w)
			changed = true // Mark as changed if we add one
		}
	}
	return changed
}

// UpdateActiveWindow updates the active window in history
// Args:
//
//	activeID: Active window ID
//
// Returns:
//
//	bool: True if the active window was found and moved.
func (h *History) UpdateActiveWindow(activeID int) bool {
	changed := false

	// Find active window
	for i, window := range h.windows {
		if i > 0 && window.ID == activeID {
			// Check title is not "gofi":
			if window.Title != "gofi" {
				// Move to front
				h.windows = append([]*shared.Window{window}, append(h.windows[:i], h.windows[i+1:]...)...)
				changed = true // Mark changed only if moved
				log.Debug("Updating active window: %d", activeID)
			}
			break
		}
	}
	return changed
}

// GetActiveID returns the ID of the currently active window
// Returns:
//
//	int: ID of active window or 0 if none
func (h *History) GetActiveID() int {
	if len(h.windows) == 0 {
		return 0
	}
	return h.windows[0].ID
}

// // filterWindows filters windows by IDs
// // Args:
// //
// //	windows: List of windows
// //	ids: List of IDs to keep
// //
// // Returns:
// //
// //	[]*shared.Window: Filtered list of windows
// func filterWindows(windows []*shared.Window, ids []int) []*shared.Window {
// 	result := make([]*shared.Window, 0, len(windows))
// 	idSet := make(map[int]struct{})
// 	for _, id := range ids {
// 		idSet[id] = struct{}{}
// 	}

// 	for _, w := range windows {
// 		if _, exists := idSet[w.ID]; exists {
// 			result = append(result, w)
// 		}
// 	}
// 	return result
// }
//
// // findWindowByID finds a window by its ID
// // Args:
// //
// //	windows: List of windows
// //	id: Window ID to find
// //
// // Returns:
// //
// //	*shared.Window: Found window or nil
// func findWindowByID(windows []*shared.Window, id int) *shared.Window {
// 	for _, w := range windows {
// 		if w.ID == id {
// 			return w
// 		}
// 	}
// 	return nil
// }

// // removeWindowByID removes a window by its ID
// // Args:
// //
// //	windows: List of windows
// //	id: Window ID to remove
// //
// // Returns:
// //
// //	[]*shared.Window: List with window removed
// func removeWindowByID(windows []*shared.Window, id int) []*shared.Window {
// 	result := make([]*shared.Window, 0, len(windows))
// 	for _, w := range windows {
// 		if w.ID != id {
// 			result = append(result, w)
// 		}
// 	}
// 	return result
// }
