package ipc

import (
	"fmt"
	"gofi/pkg/log"
	"net"
)

// CommandHandler defines the interface for command handlers
// Args:
//
//	command: The command to handle
//
// Returns:
//
//	string: The response to send
//	error: Any error that occurred
type CommandHandler interface {
	Handle(command string) (string, error)
}

// CommandHandlers is a map of command names to handlers
type CommandHandlers map[string]CommandHandler

// ProcessCommands runs the daemon command processing loop
// Args:
//
//	handlers: Map of command names to handler functions
//
// Returns:
//
//	error: Any error that occurred
func ProcessCommands(handlers CommandHandlers) error {
	listener, err := SetupSocket()
	if err != nil {
		return fmt.Errorf("failed to open command socket: %w", err)
	}
	defer listener.Close()

	for {
		if err := processOneCommand(listener, handlers); err != nil {
			log.Error("Error processing command: %s", err)
			continue
		} else if !isAlive(listener) {
			break
		}
	}

	return nil
}

// isAlive checks if the socket is still alive
// Args:
//
//	listener: The socket listener to check
//
// Returns:
//
//	bool: True if socket is alive, False otherwise
func isAlive(listener net.Listener) bool {
	addr := listener.Addr()
	return addr != nil
}

// processOneCommand processes a single command from a client
// Args:
//
//	listener: The socket listener
//	handlers: Map of command names to handler functions
//
// Returns:
//
//	error: Any error that occurred
func processOneCommand(listener net.Listener, handlers CommandHandlers) error {
	conn, err := listener.Accept()
	if err != nil {
		log.Error("Error accepting connection: %s", err)
		return err
	}
	defer conn.Close()

	command, err := ReceiveMessage(conn)
	if err != nil {
		log.Error("Error receiving command: %s", err)
		return err
	}

	if handler, ok := handlers[command]; ok {
		response, err := handler.Handle(command)
		if err != nil {
			log.Error("Error handling command: %s", err)
			return err
		}

		if err := SendResponse(conn, response); err != nil {
			log.Error("Failed to send response: %s", err)
			return err
		}

		if command == "QUIT" && response == "BYE" {
			log.Info("Received QUIT command, stopping command loop")
			return nil
		}
	} else {
		log.Error("Unknown command: %s", command)
		if err := SendResponse(conn, "ERROR: Unknown command"); err != nil {
			log.Error("Failed to send error response: %s", err)
			return err
		}
	}

	return nil
}
