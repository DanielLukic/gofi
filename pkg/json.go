package pkg

import (
	"encoding/json"
	"fmt"
)

// marshal_json converts a Go object to a compact JSON string without newlines
func marshal_json(v interface{}) (string, error) {
	// Marshal the object to JSON bytes
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %v", err)
	}

	// Convert bytes to string (no pretty printing)
	return string(bytes), nil
}

// unmarshal_json parses a JSON string into a Go object
func unmarshal_json(data string, v interface{}) error {
	// Convert string to bytes and unmarshal
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return nil
}

// marshal_window_list converts a slice of Window objects to a compact JSON string
func marshal_window_list(windows []Window) (string, error) {
	return marshal_json(windows)
}

// unmarshal_window_list parses a JSON string into a slice of Window objects
func unmarshal_window_list(data string) ([]Window, error) {
	var windows []Window
	if err := unmarshal_json(data, &windows); err != nil {
		return nil, err
	}
	return windows, nil
}
