package desktop

import (
	"context"
	"gofi/pkg/shared"
)

// WindowManager is the interface for window management operations
type WindowManager interface {
	// InitEvents initializes event handling
	InitEvents() bool

	// AwaitEvent waits for and handles events with optional abort mechanism
	// Args:
	//     ctx: Context for cancellation
	// Returns:
	//     The event or nil if aborted or error occurred
	AwaitEvent(ctx context.Context) string

	// ActiveWindowID gets the ID of the currently active window
	// Returns:
	//     The window ID or 0 if no active window
	ActiveWindowID() int

	// StackingList gets information about all windows
	// Returns:
	//     List of windows
	StackingList() []*shared.Window

	// CloseWindow requests the closing of a window
	// Args:
	//     windowID: ID of the window to close
	// Returns:
	//     Error if the window could not be closed
	CloseWindow(windowID int) error

	// WindowTitle gets the title of a window
	// Args:
	//     windowID: ID of the window
	// Returns:
	//     The title of the window
	WindowTitle(windowID int) string

	// WindowClass gets the class and instance of a window
	// Args:
	//     windowID: ID of the window
	// Returns:
	//     The class and instance of the window
	WindowClass(windowID int) (string, string)
}
