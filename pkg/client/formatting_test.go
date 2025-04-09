package client_test

import (
	"testing"

	"gofi/pkg/client"
	"gofi/pkg/shared"
)

var testWidths = map[string]int{
	"desktop":  4,
	"instance": 20,
	"title":    20,
	"class":    18,
}

func makeWindow(t *testing.T, opts ...func(*shared.Window)) *shared.Window {
	w := &shared.Window{
		ID:        16777216,
		Title:     "Window Title",
		ClassName: "Class Name",
		Instance:  "Instance Name",
		Type:      "NORMAL",
		Desktop:   1,
		PID:       1234,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func TestFormatWindows(t *testing.T) {
	windows := []shared.Window{
		*makeWindow(t, func(w *shared.Window) {
			w.ID = 1
			w.Title = "Win1"
			w.ClassName = "Class1"
			w.Instance = "i1"
			w.Desktop = 1
		}),
		*makeWindow(t, func(w *shared.Window) {
			w.ID = 2
			w.Title = "LongWindowTitle NeedsTruncate"
			w.ClassName = "Class2"
			w.Instance = "Instance2"
			w.Desktop = 2
		}),
	}

	result := client.FormatWindows(windows, testWidths, nil)

	if len(result) != 2 {
		t.Fatalf("formatWindows length incorrect: got %d, want 2", len(result))
	}
	expected1 := "[1]  i1                   Win1                 Class1             0x1"
	expected2 := "[2]  Class2               LongWindowTitle Need Instance2          0x2"

	if result[0] != expected1 {
		t.Errorf("formatWindows first window incorrect:\n GOT: %q\nWANT: %q", result[0], expected1)
	}
	if result[1] != expected2 {
		t.Errorf("formatWindows second window incorrect:\n GOT: %q\nWANT: %q", result[1], expected2)
	}
}

func TestFormatWindowsEmpty(t *testing.T) {
	result := client.FormatWindows([]shared.Window{}, testWidths, nil)

	if len(result) != 0 {
		t.Errorf("formatWindows empty list incorrect: got %d, want 0", len(result))
	}
}
