package client

import (
	"encoding/json"
	"gofi/pkg/ipc"
	"gofi/pkg/log"
	"gofi/pkg/shared"
)

// SendHello sends a HELLO command to check if daemon is responsive
// Returns:
//
//	bool: True if daemon responded with HELLO, False otherwise
func SendHello() bool {
	response, err := ipc.SendMessage("HELLO")
	if err != nil {
		log.Error("Failed to send HELLO: %s", err)
		return false
	}

	return response == "HELLO"
}

// ActiveWindowList gets the active window list from daemon
// Returns:
//
//	[]Window: List of window dictionaries or nil on error
func ActiveWindowList() []shared.Window {
	response, err := ipc.SendMessage("ACTIVE_WINDOW_LIST")
	if err != nil {
		log.Error("Failed to get window list: %s", err)
		return nil
	}

	var windows []shared.Window
	if err := json.Unmarshal([]byte(response), &windows); err != nil {
		log.Error("Failed to parse window list: %s", err)
		return nil
	}

	return windows
}

// sendCommand sends a command to the daemon
// Args:
//
//	command: The command to send
//
// Returns:
//
//	string: Response from daemon or empty string if error
func SendCommand(command string) string {
	if !ipc.CheckSocketExists() {
		log.Error("Socket doesn't exist. Is the daemon running?")
		return ""
	}

	response, err := ipc.SendMessage(command)
	if err != nil {
		log.Error("Error sending command: %s", err)
		return ""
	}

	return response
}
