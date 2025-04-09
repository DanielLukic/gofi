package daemon

import (
	"encoding/json"
	"fmt"

	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// HandleHello handles the HELLO command
// Returns:
//
//	string: Simple HELLO response
func HandleHello() string {
	return "HELLO"
}

// HandleActiveWindowList handles the ACTIVE_WINDOW_LIST command
// Args:
//
//	windows: List of windows
//
// Returns:
//
//	string: JSON string containing window list data
func HandleActiveWindowList(windows []shared.Window) string {
	jsonData, err := json.Marshal(windows)
	if err != nil {
		log.Error(fmt.Sprintf("Error marshaling window list: %s", err))
		return fmt.Sprintf("ERROR: %s", err)
	}

	return string(jsonData)
}

// HandleQuit handles the QUIT command
// Returns:
//
//	string: BYE response to acknowledge the quit request
func HandleQuit() string {
	log.Info("Received QUIT command")
	return "BYE"
}
