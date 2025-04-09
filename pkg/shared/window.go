package shared

import (
	"encoding/json"
	"fmt"
)

// Window represents a window in the window manager
// Fields:
//
//	ID: Window ID
//	Title: Window title
//	ClassName: Window class name
//	Type: Window type
//	Instance: Window instance
//	Desktop: Desktop number
//	PID: Process ID
type Window struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	ClassName string `json:"class_name"`
	Type      string `json:"type"`
	Instance  string `json:"instance"`
	Desktop   int    `json:"desktop"`
	PID       int    `json:"pid"`
}

// HexID returns the window ID in hex format for wmctrl
// Returns:
//
//	string: Window ID in hex format
func (w Window) HexID() string {
	return fmt.Sprintf("0x%x", w.ID)
}

// DesktopStr returns the desktop number for display in selector
// Returns:
//
//	string: Desktop number in format [X] or [S] if invalid
func (w Window) DesktopStr() string {
	if w.Desktop < 0 || w.Desktop > 99 {
		return "[S]"
	}
	return fmt.Sprintf("[%d]", w.Desktop)
}

// String returns a string representation of the window
// Returns:
//
//	string: Window information
func (w Window) String() string {
	return fmt.Sprintf("Window{%d, %s, %s, %s, %s, %d, %d}",
		w.ID, w.Title, w.ClassName, w.Type, w.Instance, w.Desktop, w.PID)
}

// NewWindow creates a new Window instance
// Args:
//
//	id: Window ID
//	title: Window title
//	className: Window class name
//	type: Window type
//	instance: Window instance
//	desktop: Desktop number
//	pid: Process ID
//
// Returns:
//
//	*Window: New window instance
func NewWindow(id int, title, className, typeStr, instance string, desktop, pid int) *Window {
	return &Window{
		ID:        id,
		Title:     title,
		ClassName: className,
		Type:      typeStr,
		Instance:  instance,
		Desktop:   desktop,
		PID:       pid,
	}
}

// MarshalJSON implements json.Marshaler interface
// Returns:
//
//	[]byte: JSON representation of Window
//	error: Any error that occurred
func (w Window) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		ClassName string `json:"class_name"`
		Type      string `json:"type"`
		Instance  string `json:"instance"`
		Desktop   int    `json:"desktop"`
		PID       int    `json:"pid"`
	}{
		ID:        w.ID,
		Title:     w.Title,
		ClassName: w.ClassName,
		Type:      w.Type,
		Instance:  w.Instance,
		Desktop:   w.Desktop,
		PID:       w.PID,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface
// Args:
//
//	data: JSON data to unmarshal
//
// Returns:
//
//	error: Any error that occurred
func (w *Window) UnmarshalJSON(data []byte) error {
	type Alias Window
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	return json.Unmarshal(data, &aux)
}
