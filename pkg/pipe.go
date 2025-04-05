package pkg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	// command_pipe_path is the path to the named pipe for sending commands to the daemon
	command_pipe_path = "/tmp/gofi.command.pipe"
	// response_pipe_path is the path to the named pipe for receiving responses from the daemon
	response_pipe_path = "/tmp/gofi.response.pipe"
)

// setup_pipes creates both command and response pipes if they don't exist
func setup_pipes() error {
	// Create the command pipe if it doesn't exist
	if _, err := os.Stat(command_pipe_path); os.IsNotExist(err) {
		if err := syscall.Mkfifo(command_pipe_path, 0666); err != nil {
			return fmt.Errorf("failed to create command pipe: %v", err)
		}
	}

	// Create the response pipe if it doesn't exist
	if _, err := os.Stat(response_pipe_path); os.IsNotExist(err) {
		if err := syscall.Mkfifo(response_pipe_path, 0666); err != nil {
			return fmt.Errorf("failed to create response pipe: %v", err)
		}
	}

	return nil
}

// cleanup_pipes removes both command and response pipes
func cleanup_pipes() error {
	// Remove the command pipe if it exists
	if _, err := os.Stat(command_pipe_path); err == nil {
		if err := os.Remove(command_pipe_path); err != nil {
			return fmt.Errorf("failed to remove command pipe: %v", err)
		}
	}

	// Remove the response pipe if it exists
	if _, err := os.Stat(response_pipe_path); err == nil {
		if err := os.Remove(response_pipe_path); err != nil {
			return fmt.Errorf("failed to remove response pipe: %v", err)
		}
	}

	return nil
}

// wait_for_pipes waits for both pipes to be created
func wait_for_pipes(timeout_ms int) error {
	for i := 0; i < timeout_ms/100; i++ {
		command_pipe_exists := false
		response_pipe_exists := false

		if _, err := os.Stat(command_pipe_path); err == nil {
			command_pipe_exists = true
		}

		if _, err := os.Stat(response_pipe_path); err == nil {
			response_pipe_exists = true
		}

		if command_pipe_exists && response_pipe_exists {
			return nil
		}

		// Wait 100ms before checking again
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for pipes to be created")
}

// open_command_pipe_read opens the command pipe for reading
func open_command_pipe_read() (*os.File, error) {
	pipe, err := os.OpenFile(command_pipe_path, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to open command pipe for reading: %v", err)
	}
	return pipe, nil
}

// open_command_pipe_write opens the command pipe for writing
func open_command_pipe_write() (*os.File, error) {
	pipe, err := os.OpenFile(command_pipe_path, os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open command pipe for writing: %v", err)
	}
	return pipe, nil
}

// open_response_pipe_read opens the response pipe for reading
func open_response_pipe_read() (*os.File, error) {
	pipe, err := os.OpenFile(response_pipe_path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open response pipe for reading: %v", err)
	}
	return pipe, nil
}

// open_response_pipe_write opens the response pipe for writing
func open_response_pipe_write() (*os.File, error) {
	pipe, err := os.OpenFile(response_pipe_path, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to open response pipe for writing: %v", err)
	}
	return pipe, nil
}

// write_to_daemon writes a message to the daemon through the command pipe
func write_to_daemon(message string) error {
	// Make sure the message ends with a newline
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	// Open the command pipe for writing
	pipe, err := open_command_pipe_write()
	if err != nil {
		return err
	}
	defer pipe.Close()

	// Write the message to the pipe
	if _, err := pipe.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write to command pipe: %v", err)
	}

	return nil
}

// write_to_client writes a message to the client through the response pipe
func write_to_client(message string) error {
	// Make sure the message ends with a newline
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}

	// Open the response pipe for writing
	pipe, err := open_response_pipe_write()
	if err != nil {
		return err
	}
	defer pipe.Close()

	// Write the message to the pipe
	if _, err := pipe.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write to response pipe: %v", err)
	}

	return nil
}

// read_from_pipe reads a message from the given pipe
func read_from_pipe(pipe *os.File) (string, error) {
	// Read a line from the pipe
	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// read_from_daemon reads a message from the daemon through the response pipe
func read_from_daemon() (string, error) {
	// Open the response pipe for reading
	pipe, err := open_response_pipe_read()
	if err != nil {
		return "", err
	}
	defer pipe.Close()

	// Read a line from the pipe
	return read_from_pipe(pipe)
}

// read_json_from_daemon reads a JSON response from the daemon
func read_json_from_daemon() (string, error) {
	// Simply read a single line, which will contain the entire JSON
	return read_from_daemon()
}

// check_pipes_exist checks if both pipes exist
func check_pipes_exist() bool {
	command_pipe_exists := false
	response_pipe_exists := false

	if _, err := os.Stat(command_pipe_path); err == nil {
		command_pipe_exists = true
	}

	if _, err := os.Stat(response_pipe_path); err == nil {
		response_pipe_exists = true
	}

	return command_pipe_exists && response_pipe_exists
}

// send_and_receive_simple sends a command to the daemon and reads a simple response
func send_and_receive_simple(command string) (string, error) {
	// Send the command
	if err := write_to_daemon(command); err != nil {
		return "", err
	}

	// Read the response
	response, err := read_from_daemon()
	if err != nil {
		return "", err
	}

	return response, nil
}

// send_and_receive_json sends a command to the daemon and reads a JSON response
func send_and_receive_json(command string) (string, error) {
	// Send the command
	if err := write_to_daemon(command); err != nil {
		return "", err
	}

	// Read the JSON response
	return read_json_from_daemon()
}
