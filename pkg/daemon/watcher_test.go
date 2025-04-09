package daemon

import (
	"testing"

	"gofi/pkg/desktop"
	"gofi/pkg/shared"
)

func TestWatcher(t *testing.T) {
	// Create mock window manager
	wm := desktop.NewMockWindowManager()

	// Create API
	api := NewAPI()

	// Create watcher
	watcher := NewWindowWatcher(wm, api)

	// Test starting watcher
	if !watcher.Start() {
		t.Error("Failed to start watcher")
	}

	// Test stopping watcher
	if !watcher.Stop() {
		t.Error("Failed to stop watcher")
	}
}

// mockAPI is a mock implementation of the API interface
type mockAPI struct {
	windows []*shared.Window
}

func (m *mockAPI) ClientList() []*shared.Window {
	return m.windows
}

func (m *mockAPI) AddWindow(w *shared.Window) {
	m.windows = append(m.windows, w)
}

func (m *mockAPI) RemoveWindow(w *shared.Window) {
	for i, win := range m.windows {
		if win.ID == w.ID {
			m.windows = append(m.windows[:i], m.windows[i+1:]...)
			break
		}
	}
}
