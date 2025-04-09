package daemon

import (
	"testing"

	"gofi/pkg/desktop"
)

func TestWindowList(t *testing.T) {
	// Create mock window manager
	wm := desktop.NewMockWindowManager()

	// Create window list with mock window manager
	wl := NewWindowList(wm, nil)

	// Initialize with mock windows
	wl.Initialize()
	wl.UpdateWindowList()

	// Test getting window list
	windows := wl.ClientList()
	if len(windows) == 0 {
		t.Error("Expected at least one window in the list")
	}

	// Test window properties
	for _, win := range windows {
		if win == nil {
			t.Error("Found nil window in list")
			continue
		}
		if win.ID == 0 {
			t.Error("Found window with invalid ID 0")
		}
		if win.Title == "" {
			t.Error("Found window with empty title")
		}
		if win.Type == "" {
			t.Error("Found window with empty type")
		}
		if win.Desktop < -1 {
			t.Error("Found window with invalid desktop number")
		}
	}
}
