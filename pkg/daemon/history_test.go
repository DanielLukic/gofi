package daemon

import (
	"testing"

	"gofi/pkg/shared"
)

func TestHistory(t *testing.T) {
	// Test adding and retrieving history
	h := NewHistory()
	window := &shared.Window{
		ID:        1,
		Title:     "Test Window",
		ClassName: "Test",
		Instance:  "test",
		Type:      "NORMAL",
		Desktop:   1,
		PID:       1234,
	}

	// Add window to history
	h.Initialize([]*shared.Window{window})
	if len(h.windows) != 1 {
		t.Errorf("History length incorrect: got %d", len(h.windows))
	}

	// Add same window again - should not be duplicated
	h.AddNew([]*shared.Window{window})
	if len(h.windows) != 1 {
		t.Errorf("History should not have duplicates: got %d", len(h.windows))
	}

	// Test clearing history
	h.KeepOnly([]*shared.Window{})
	if len(h.windows) != 0 {
		t.Errorf("History not cleared: got %d", len(h.windows))
	}
}
