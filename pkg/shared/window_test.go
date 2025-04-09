package shared

import (
	"encoding/json"
	"testing"
)

func TestWindowJSON(t *testing.T) {
	// Test marshaling
	window := Window{
		ID:        123,
		Title:     "Test Window",
		ClassName: "TestClass",
		Type:      "Normal",
		Instance:  "test-instance",
		Desktop:   1,
		PID:       456,
	}

	jsonData, err := json.Marshal(window)
	if err != nil {
		t.Fatalf("Failed to marshal window: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Window
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal window: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != window.ID {
		t.Errorf("ID mismatch: got %d, want %d", unmarshaled.ID, window.ID)
	}
	if unmarshaled.Title != window.Title {
		t.Errorf("Title mismatch: got %s, want %s", unmarshaled.Title, window.Title)
	}
	if unmarshaled.ClassName != window.ClassName {
		t.Errorf("ClassName mismatch: got %s, want %s", unmarshaled.ClassName, window.ClassName)
	}
	if unmarshaled.Type != window.Type {
		t.Errorf("Type mismatch: got %s, want %s", unmarshaled.Type, window.Type)
	}
	if unmarshaled.Instance != window.Instance {
		t.Errorf("Instance mismatch: got %s, want %s", unmarshaled.Instance, window.Instance)
	}
	if unmarshaled.Desktop != window.Desktop {
		t.Errorf("Desktop mismatch: got %d, want %d", unmarshaled.Desktop, window.Desktop)
	}
	if unmarshaled.PID != window.PID {
		t.Errorf("PID mismatch: got %d, want %d", unmarshaled.PID, window.PID)
	}
}

func TestWindowString(t *testing.T) {
	window := Window{
		ID:        123,
		Title:     "Test Window",
		ClassName: "TestClass",
		Type:      "Normal",
		Instance:  "test-instance",
		Desktop:   1,
		PID:       456,
	}

	expected := "Window{123, Test Window, TestClass, Normal, test-instance, 1, 456}"
	if window.String() != expected {
		t.Errorf("String() mismatch: got %s, want %s", window.String(), expected)
	}
}

func TestWindowHexID(t *testing.T) {
	window := Window{ID: 123}
	expected := "0x7b"
	if window.HexID() != expected {
		t.Errorf("HexID() mismatch: got %s, want %s", window.HexID(), expected)
	}
}

func TestWindowDesktopStr(t *testing.T) {
	tests := []struct {
		desktop int
		want    string
	}{
		{0, "[0]"},
		{1, "[1]"},
		{99, "[99]"},
		{-1, "[S]"},
		{100, "[S]"},
	}

	for _, tt := range tests {
		window := Window{Desktop: tt.desktop}
		if got := window.DesktopStr(); got != tt.want {
			t.Errorf("DesktopStr() for %d: got %s, want %s", tt.desktop, got, tt.want)
		}
	}
}

func TestWindowUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Window
		wantErr bool
	}{
		{
			name: "complete window",
			json: `{
				"id": 123,
				"title": "Test Window",
				"class_name": "TestClass",
				"type": "Normal",
				"instance": "test-instance",
				"desktop": 1,
				"pid": 456
			}`,
			want: Window{
				ID:        123,
				Title:     "Test Window",
				ClassName: "TestClass",
				Type:      "Normal",
				Instance:  "test-instance",
				Desktop:   1,
				PID:       456,
			},
			wantErr: false,
		},
		{
			name: "minimal window",
			json: `{
				"id": 123,
				"title": "",
				"class_name": "",
				"type": "",
				"instance": "",
				"desktop": 0,
				"pid": 0
			}`,
			want: Window{
				ID:        123,
				Title:     "",
				ClassName: "",
				Type:      "",
				Instance:  "",
				Desktop:   0,
				PID:       0,
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid json}`,
			want:    Window{},
			wantErr: true,
		},
		{
			name: "missing id field",
			json: `{
				"title": "Test Window",
				"class_name": "TestClass",
				"type": "Normal",
				"instance": "test-instance",
				"desktop": 1,
				"pid": 456
			}`,
			want: Window{
				ID:        0, // Default value for int
				Title:     "Test Window",
				ClassName: "TestClass",
				Type:      "Normal",
				Instance:  "test-instance",
				Desktop:   1,
				PID:       456,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Window
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("Unmarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWindowMarshal(t *testing.T) {
	tests := []struct {
		name    string
		window  Window
		want    string
		wantErr bool
	}{
		{
			name: "complete window",
			window: Window{
				ID:        123,
				Title:     "Test Window",
				ClassName: "TestClass",
				Type:      "Normal",
				Instance:  "test-instance",
				Desktop:   1,
				PID:       456,
			},
			want:    `{"id":123,"title":"Test Window","class_name":"TestClass","type":"Normal","instance":"test-instance","desktop":1,"pid":456}`,
			wantErr: false,
		},
		{
			name:    "zero values",
			window:  Window{},
			want:    `{"id":0,"title":"","class_name":"","type":"","instance":"","desktop":0,"pid":0}`,
			wantErr: false,
		},
		{
			name: "special characters",
			window: Window{
				ID:        123,
				Title:     "Test \"Window\" with \\ and /",
				ClassName: "Test\nClass",
				Type:      "Normal\t",
				Instance:  "test-instance",
				Desktop:   1,
				PID:       456,
			},
			want:    `{"id":123,"title":"Test \"Window\" with \\ and /","class_name":"Test\nClass","type":"Normal\t","instance":"test-instance","desktop":1,"pid":456}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.window)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("Marshal() got = %v, want %v", string(got), tt.want)
			}
		})
	}
}
