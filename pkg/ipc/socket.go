package ipc

import (
	"fmt"
	"gofi/pkg/log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	SocketPath = "/tmp/gofi.socket"
	Timeout    = 5 * time.Second
)

// CheckSocketExists checks if the socket file exists
// Returns:
//
//	bool: True if socket exists, False otherwise
func CheckSocketExists() bool {
	_, err := os.Stat(SocketPath)
	return err == nil
}

// SetupSocket creates and binds a Unix domain socket
// Returns:
//
//	net.Listener: The socket listener
//	error: Any error that occurred
func SetupSocket() (net.Listener, error) {
	log.Debug("Create Unix domain socket at %s", SocketPath)

	if err := CleanupSocket(); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		log.Error("Failed to set up socket: %s", err)
		return nil, err
	}

	if err := os.Chmod(SocketPath, 0600); err != nil {
		log.Error("Failed to set permissions on socket: %s", err)
		return nil, err
	}

	return listener, nil
}

// CleanupSocket removes the existing socket file
// Returns:
//
//	error: Any error that occurred
func CleanupSocket() error {
	if _, err := os.Stat(SocketPath); err == nil {
		log.Debug("Remove socket at %s", SocketPath)
		if err := os.Remove(SocketPath); err != nil {
			log.Error("Failed to remove socket: %s", err)
			return err
		}
	}
	return nil
}

// SendMessage sends a message to the socket
// Args:
//
//	message: The message to send
//
// Returns:
//
//	string: The response from the socket
//	error: Any error that occurred
func SendMessage(message string) (string, error) {
	if message == "" {
		return "", fmt.Errorf("empty message")
	}

	if strings.Contains(message, "\n") {
		return "", fmt.Errorf("message contains newline character")
	}

	conn, err := net.DialTimeout("unix", SocketPath, Timeout)
	if err != nil {
		log.Error("Failed to connect to socket: %s", err)
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", message)))
	if err != nil {
		log.Error("Failed to send message: %s", err)
		return "", err
	}

	log.Debug("Send message to socket: %s", message)

	conn.SetReadDeadline(time.Now().Add(Timeout))
	buffer := make([]byte, 4096)

	n, err := conn.Read(buffer)
	if err != nil {
		log.Error("Failed to read response: %s", err)
		return "", err
	}

	response := strings.TrimSpace(string(buffer[:n]))
	logResponse := truncateForLog(response)
	log.Debug("Received response: %s", logResponse)

	return response, nil
}

// ReceiveMessage receives a message from a connection
// Args:
//
//	conn: The connection to receive from
//
// Returns:
//
//	string: The received message
//	error: Any error that occurred
func ReceiveMessage(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(Timeout))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Error("Error receiving message: %s", err)
		return "", err
	}

	message := strings.TrimSpace(string(buffer[:n]))
	if strings.Contains(message, "\n") {
		return "", fmt.Errorf("received message contains newline character")
	}

	// Truncate message for logging if needed
	logMessage := truncateForLog(message)
	log.Debug("Received message: %s", logMessage)

	return message, nil
}

// SendResponse sends a response back to the client
// Args:
//
//	conn: The connection to send to
//	response: The response to send
//
// Returns:
//
//	error: Any error that occurred
func SendResponse(conn net.Conn, response string) error {
	if response == "" {
		return fmt.Errorf("empty response")
	}

	if strings.Contains(response, "\n") {
		return fmt.Errorf("response contains newline character")
	}

	// Truncate response for logging if needed
	logResponse := truncateForLog(response)
	log.Debug("Send response: %s", logResponse)

	_, err := conn.Write([]byte(fmt.Sprintf("%s\n", response)))
	if err != nil {
		log.Error("Failed to send response: %s", err)
		return err
	}

	return nil
}

// truncateForLog shortens a string for logging purposes.
func truncateForLog(s string) string {
	const maxLogLen = 50
	if len(s) > maxLogLen {
		return s[:maxLogLen] + "..."
	}
	return s
}
