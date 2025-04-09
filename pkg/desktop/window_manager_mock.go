package desktop

import (
	"context"
	"fmt"
	"sync"

	"gofi/pkg/shared"
)

// MockWindowManager is a mock implementation of the window manager protocol
type MockWindowManager struct {
	mu           sync.Mutex
	events       chan interface{}
	eventCount   int
	eventsInit   bool
	windows      map[int]*shared.Window
	activeWindow int
	windowIDs    []int
}

// NewMockWindowManager creates a new mock window manager instance
// Returns:
//
//	*MockWindowManager: New mock window manager instance
func NewMockWindowManager() *MockWindowManager {
	wm := &MockWindowManager{
		events:       make(chan interface{}, 10), // Buffered channel for events
		eventsInit:   true,
		windows:      make(map[int]*shared.Window),
		activeWindow: 1,
		windowIDs:    []int{1, 2, 3},
	}

	// Initialize default windows
	wm.addWindow(shared.NewWindow(
		1, "Terminal", "gnome-terminal", "Normal", "gnome-terminal", 0, 1234,
	))
	wm.addWindow(shared.NewWindow(
		2, "Browser", "firefox", "Normal", "firefox", 0, 5678,
	))
	wm.addWindow(shared.NewWindow(
		3, "Text Editor", "gedit", "Normal", "gedit", 1, 9101,
	))

	return wm
}

// EnqueueEvent adds an event to the event queue
// Args:
//
//	event: Event to enqueue
func (wm *MockWindowManager) EnqueueEvent(event interface{}) {
	wm.events <- event
}

// InitEvents initializes event handling
// Returns:
//
//	bool: True if events were initialized successfully
func (wm *MockWindowManager) InitEvents() bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.eventsInit = true
	return true
}

// AwaitEvent waits for and handles events with optional abort mechanism
// Args:
//
//	ctx: Context for cancellation
//
// Returns:
//
//	string: The event name (if a string was enqueued) or "" if aborted, error occurred, or non-string event
func (wm *MockWindowManager) AwaitEvent(ctx context.Context) string {
	select {
	case <-ctx.Done():
		return "" // Return empty string on cancellation
	default:
		// Non-blocking check for context done before acquiring lock
		select {
		case <-ctx.Done():
			return ""
		default:
		}

		wm.mu.Lock()
		defer wm.mu.Unlock()

		// Check for events
		select {
		case event := <-wm.events:
			wm.eventCount++
			// Return event name if it's a string, otherwise empty string
			if eventName, ok := event.(string); ok {
				return eventName
			}
			// Log if a non-string event was dequeued in the mock?
			// log.Warn("MockWindowManager dequeued non-string event: %T", event)
			return "" // Return empty string for non-string types
		default:
			// No event available
			return "" // Return empty string if no event
		}
	}
}

// ActiveWindowID gets the ID of the currently active window
// Returns:
//
//	int: The window ID or 0 if no active window
func (wm *MockWindowManager) ActiveWindowID() int {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.activeWindow
}

// StackingList gets information about all windows
// Returns:
//
//	[]*shared.Window: List of windows
func (wm *MockWindowManager) StackingList() []*shared.Window {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	windows := make([]*shared.Window, len(wm.windowIDs))
	for i, id := range wm.windowIDs {
		windows[i] = wm.windows[id]
	}
	return windows
}

// WindowTitle gets the title of a window
// Args:
//
//	windowID: Window ID
//
// Returns:
//
//	string: Window title or "<no title>" if not found
func (wm *MockWindowManager) WindowTitle(windowID int) string {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	window := wm.windows[windowID]
	if window != nil {
		return window.Title
	}
	return "<no title>"
}

// WindowClass gets the class of a window
// Args:
//
//	windowID: Window ID
//
// Returns:
//
//	string: Window class name
//	string: Window instance
func (wm *MockWindowManager) WindowClass(windowID int) (string, string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	window := wm.windows[windowID]
	if window != nil {
		return window.Instance, window.ClassName
	}
	return "", ""
}

// WindowDesktop gets the desktop number of a window
// Args:
//
//	windowID: Window ID
//
// Returns:
//
//	int: Desktop number or -1 if not found
func (wm *MockWindowManager) WindowDesktop(windowID int) int {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	window := wm.windows[windowID]
	if window != nil {
		return window.Desktop
	}
	return -1
}

// WindowPID gets the process ID of a window
// Args:
//
//	windowID: Window ID
//
// Returns:
//
//	int: Process ID or -1 if not found
func (wm *MockWindowManager) WindowPID(windowID int) int {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	window := wm.windows[windowID]
	if window != nil {
		return window.PID
	}
	return -1
}

// SetActiveWindow sets the active window for testing
// Args:
//
//	windowID: Window ID to set as active
func (wm *MockWindowManager) SetActiveWindow(windowID int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if _, ok := wm.windows[windowID]; ok {
		wm.activeWindow = windowID
		// Move to front of MRU list
		wm.windowIDs = append([]int{windowID}, wm.windowIDs...)
		for i := 1; i < len(wm.windowIDs); i++ {
			if wm.windowIDs[i] == windowID {
				wm.windowIDs = append(wm.windowIDs[:i], wm.windowIDs[i+1:]...)
				break
			}
		}
	}
}

// AddWindow adds a new window for testing
// Args:
//
//	window: Window to add
func (wm *MockWindowManager) AddWindow(window *shared.Window) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.windows[window.ID] = window
	wm.windowIDs = append([]int{window.ID}, wm.windowIDs...)
}

// RemoveWindow removes a window for testing
// Args:
//
//	windowID: Window ID to remove
func (wm *MockWindowManager) RemoveWindow(windowID int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if _, ok := wm.windows[windowID]; ok {
		delete(wm.windows, windowID)
		// Remove from MRU list
		for i := 0; i < len(wm.windowIDs); i++ {
			if wm.windowIDs[i] == windowID {
				wm.windowIDs = append(wm.windowIDs[:i], wm.windowIDs[i+1:]...)
				break
			}
		}
		// Update active window if it was removed
		if wm.activeWindow == windowID {
			if len(wm.windowIDs) > 0 {
				wm.activeWindow = wm.windowIDs[0]
			} else {
				wm.activeWindow = 0
			}
		}
	}
}

// CloseWindow simulates closing a window by removing it from the mock state.
// Returns an error if the windowID does not exist.
func (wm *MockWindowManager) CloseWindow(windowID int) error {
	wm.mu.Lock()
	_, exists := wm.windows[windowID]
	wm.mu.Unlock() // Unlock before potentially returning error or calling RemoveWindow (which locks again)

	if !exists {
		return fmt.Errorf("mock window %d not found, cannot close", windowID)
	}

	// Use the existing RemoveWindow logic which handles maps, slices, and active window update
	wm.RemoveWindow(windowID)
	// Optionally, enqueue a DestroyNotify event here if needed for testing downstream consumers
	// wm.EnqueueEvent(fmt.Sprintf("DestroyNotify:%d", windowID)) // Example event
	return nil
}

// addWindow adds a window to the internal map
// Args:
//
//	window: Window to add
func (wm *MockWindowManager) addWindow(window *shared.Window) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.windows[window.ID] = window
	wm.windowIDs = append([]int{window.ID}, wm.windowIDs...)
}
