package client

import (
	"fmt"
	"regexp"
	"strings"

	"gofi/pkg/shared"
)

// ColumnOrder defines the order of columns in the formatted output
var ColumnOrder = []string{
	"desktop",
	"instance",
	"title",
	"class",
	"window_id",
}

// ColumnWidths defines the width of each column
var ColumnWidths = map[string]int{
	"desktop":  4,  // e.g., "[0] "
	"instance": 20, // Increased width
	"title":    55,
	"class":    18, // Increased width
}

// FormatWindows formats windows for display
// Args:
//
//	windows: List of windows to format
//	widths: Optional column widths
//	order: Optional column order
//
// Returns:
//
//	[]string: List of formatted window lines
func FormatWindows(
	windows []shared.Window,
	widths map[string]int,
	order []string,
) []string {
	if len(windows) == 0 {
		return nil
	}

	if widths == nil {
		widths = ColumnWidths
	}
	if order == nil {
		order = ColumnOrder
	}

	lines := make([]string, len(windows))
	for i, window := range windows {
		props := formatWindow(window, widths)
		lines[i] = formatLine(props, order)
	}

	return lines
}

// formatWindow formats a single window
// Args:
//
//	window: Window to format
//	widths: Column widths
//
// Returns:
//
//	map[string]string: Formatted window properties
func formatWindow(window shared.Window, widths map[string]int) map[string]string {
	windowID := window.HexID()
	desktop := window.DesktopStr()

	// Get original instance and class names
	instanceName := window.Instance
	className := window.ClassName
	title := window.Title

	// Swap class_name and instance if instance starts uppercase
	if len(instanceName) > 0 && instanceName[0] >= 'A' && instanceName[0] <= 'Z' {
		instanceName, className = className, instanceName
	}

	// Now fit the potentially swapped names to columns
	instanceFitted := fitColumn(instanceName, widths["instance"])
	classFitted := fitColumn(className, widths["class"])
	titleFitted := fitColumn(title, widths["title"])
	desktopFitted := fitColumn(desktop, widths["desktop"])

	return map[string]string{
		"desktop":   desktopFitted,
		"instance":  instanceFitted,
		"title":     titleFitted,
		"class":     classFitted,
		"window_id": windowID, // window_id is not fitted/padded
	}
}

// fitColumn fits text to column width, padding with spaces.
// Args:
//
//	text: Text to fit
//	width: Width of the column
//
// Returns:
//
//	string: Fitted text
func fitColumn(text string, width int) string {
	if text == "" {
		return strings.Repeat(" ", width) // Return padding if empty
	}

	// Remove newlines and extra spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(` {2,}`).ReplaceAllString(text, " ")

	// Truncate if needed *before* padding (using runes for safety)
	runes := []rune(text)
	if len(runes) > width {
		text = string(runes[:width])
	}

	// Use Sprintf for left-aligned padding
	return fmt.Sprintf("%-*s", width, text)
}

// formatLine formats a window line with given column order
// Args:
//
//	window: Window properties
//	order: Column order
//
// Returns:
//
//	string: Formatted window line
func formatLine(window map[string]string, order []string) string {
	if order == nil {
		order = ColumnOrder
	}

	formatted := make([]string, len(order))
	for i, key := range order {
		if value, ok := window[key]; ok {
			formatted[i] = value
		}
	}
	return strings.Join(formatted, " ")
}
