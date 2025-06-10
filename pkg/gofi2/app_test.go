package gofi2

import (
	"gofi/pkg/client"
	"gofi/pkg/shared"
	"strings"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.api == nil {
		t.Error("app.api is nil")
	}

	if app.watcher == nil {
		t.Error("app.watcher is nil")
	}

	if app.isVisible {
		t.Error("app should start hidden")
	}

	if app.exitChannel == nil {
		t.Error("app.exitChannel is nil")
	}

	app.Cleanup()
}

func TestAppStartStop(t *testing.T) {
	app := NewApp()
	defer app.Cleanup()

	err := app.Start()
	if err != nil {
		t.Skipf("Skipping test that requires X11: %v", err)
	}

	app.Hide()
	if app.isVisible {
		t.Error("app should be hidden after Hide()")
	}
}

func TestFormatWindows(t *testing.T) {
	windows := []shared.Window{
		{ID: 0x123, Title: "Test Window", Desktop: 1, Instance: "test", ClassName: "TestClass"},
		{ID: 0x456, Title: "Another Window", Desktop: 0, Instance: "app", ClassName: "AppClass"},
	}

	formatted := client.FormatWindows(windows, nil, nil)

	if len(formatted) != 2 {
		t.Errorf("Expected 2 formatted lines, got %d", len(formatted))
	}

	if !strings.Contains(formatted[0], "Test Window") {
		t.Errorf("First line should contain 'Test Window', got %q", formatted[0])
	}

	if !strings.Contains(formatted[1], "Another Window") {
		t.Errorf("Second line should contain 'Another Window', got %q", formatted[1])
	}
}

func TestCreateTempFiles(t *testing.T) {
	tempFiles := createTempFiles()
	if tempFiles == nil {
		t.Fatal("createTempFiles() returned nil")
	}

	defer cleanupTempFiles(tempFiles)

	expectedKeys := []string{"list", "exec", "result"}
	for _, key := range expectedKeys {
		if _, exists := tempFiles[key]; !exists {
			t.Errorf("Missing temp file key: %s", key)
		}
	}
}
