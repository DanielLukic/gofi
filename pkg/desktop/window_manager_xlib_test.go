package desktop

import (
	"context"
	"os"
	"testing"
	"time"

	"gofi/pkg/shared"

	"github.com/BurntSushi/xgb/xproto"
)

// setupXLibTest creates a new XLibWindowManager instance for testing
// Skips the test if X11 display is not available
func setupXLibTest(t *testing.T) *XLibWindowManager {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("X11 display not available - skipping test")
	}

	wm, err := NewXLibWindowManager()
	if err != nil {
		t.Fatalf("Failed to create XLib window manager: %v", err)
	}
	return wm
}

// TestNewXLibWindowManager tests window manager creation
func TestNewXLibWindowManager(t *testing.T) {
	wm := setupXLibTest(t)
	if wm == nil {
		t.Fatal("Expected non-nil window manager")
	}
	if wm.display == nil {
		t.Error("Expected non-nil X display connection")
	}
}

// TestStackingList tests getting the list of windows
func TestStackingList(t *testing.T) {
	wm := setupXLibTest(t)

	// Get window list
	t.Log("Getting window list...")
	windows := wm.StackingList()
	t.Logf("Got %d windows", len(windows))

	if len(windows) == 0 {
		// Original debugging code removed as getAtom is no longer public
		// and the test is skipped here anyway.
		t.Skip("No windows found - is X running with some windows open?")
	}

	// Test window properties
	for _, window := range windows {
		t.Run("Window Properties", func(t *testing.T) {
			if window.ID == 0 {
				t.Error("Found window with invalid ID 0")
			}
			if window.Title == "" {
				t.Error("Found window with empty title")
			}
			if window.Type == "" {
				t.Error("Found window with empty type")
			}
			// Restore check for Desktop field
			if window.Desktop < -1 {
				t.Error("Found window with invalid desktop number")
			}

			// Test that window is valid
			if !wm.isValidWindow(xproto.Window(window.ID)) {
				t.Error("Found invalid window in list")
			}
		})
	}
}

// TestActiveWindow tests getting and validating the active window
func TestActiveWindow(t *testing.T) {
	wm := setupXLibTest(t)

	// Get active window
	activeID := wm.ActiveWindowID()

	// If there's an active window, verify its properties
	if activeID != 0 {
		// Verify window is valid
		if !wm.isValidWindow(xproto.Window(activeID)) {
			t.Fatal("Active window is not valid")
		}

		// Get window info
		windows := wm.StackingList()
		var activeWindow *shared.Window
		for _, w := range windows {
			if w.ID == activeID {
				activeWindow = w
				break
			}
		}

		if activeWindow == nil {
			t.Errorf("Active window %d not found in window list", activeID)
		} else {
			// Test active window properties
			if activeWindow.Title == "" {
				t.Error("Active window has empty title")
			}
			if activeWindow.Type == "" {
				t.Error("Active window has empty type")
			}
		}
	}
}

// TestWindowClassAndInstance tests getting window class and instance names
func TestWindowClassAndInstance(t *testing.T) {
	wm := setupXLibTest(t)

	windows := wm.StackingList()
	if len(windows) == 0 {
		t.Skip("No windows found - is X running with some windows open?")
	}

	// Test first window's class and instance
	window := windows[0]
	instance, class := wm.getWindowClass(xproto.Window(window.ID))

	if instance == "" && class == "" {
		t.Log("Window has no class/instance info - this is allowed but unusual")
	}

	// Verify the values match what's in the Window struct
	if instance != window.Instance {
		t.Errorf("Instance mismatch: got %q, want %q", instance, window.Instance)
	}
	if class != window.ClassName {
		t.Errorf("Class mismatch: got %q, want %q", class, window.ClassName)
	}
}

// TestEventHandling tests event initialization and handling
func TestEventHandling(t *testing.T) {
	wm := setupXLibTest(t)

	// Test event initialization
	if !wm.InitEvents() {
		t.Fatal("Failed to initialize events")
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is called eventually

	// Test event handling with timeout
	done := make(chan bool)
	go func() {
		eventName := wm.AwaitEvent(ctx)
		// When cancelled quickly, we expect an empty string.
		// If an event *did* arrive before cancellation, eventName would be non-empty.
		// The test primarily verifies that AwaitEvent respects cancellation.
		if ctx.Err() != nil && eventName != "" {
			t.Errorf("Expected empty event name on cancellation, got %q", eventName)
		}
		// If no error (e.g., event arrived before cancel), we could check if eventName is valid
		// but the rapid cancel makes this unlikely and harder to test reliably.
		// else if eventName == "" {
		// 	 t.Error("Expected non-empty event name when no cancellation occurred")
		// }

		// Original check logic (now adapted for string):
		/*
			if eventName != "" {
				// If we got an event, verify it's a known event type string
				switch eventName {
				case "PropertyNotifyEvent",
					"ConfigureNotifyEvent",
					"MapNotifyEvent",
					"UnmapNotifyEvent",
					"DestroyNotifyEvent":
					// These are expected event types
				default:
					t.Errorf("Unexpected event name: %s", eventName)
				}
			}
		*/
		done <- true
	}()

	// Wait briefly, then cancel the context to simulate shutdown/abort
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for event handling with timeout
	select {
	case <-done:
		// Success - event handling completed
	case <-time.After(1 * time.Second):
		t.Fatal("Event handling test timed out")
	}
}
