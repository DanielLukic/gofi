package pkg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func RunDaemon() {
	if err := KillDaemon(); err != nil {
		Log("Error killing existing daemon: %v", err)
		os.Exit(1)
	}
	if err := _start_daemon(); err != nil {
		Log("Error starting daemon: %v", err)
		os.Exit(1)
	}
}

func KillDaemon() error {
	// Clean up the pipes
	Log("Cleaning up pipes...")
	if err := cleanup_pipes(); err != nil {
		Log("Error cleaning up pipes: %v", err)
	} else {
		Log("Pipes cleaned up.")
	}

	// Get our own PID to avoid killing ourselves
	selfPid := os.Getpid()

	// Find and kill all gofi daemon processes (except self)
	Log("Finding gofi daemon processes...")
	cmd := exec.Command("pgrep", "-f", "gofi.*-daemon")
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes match, which is fine
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			Log("No gofi daemon processes found.")
			return nil
		}
		Log("Error checking for gofi processes:", err)
		return fmt.Errorf("error checking for gofi processes: %v", err)
	}

	// Process each PID
	for _, pidBytes := range bytes.Split(bytes.TrimSpace(output), []byte{'\n'}) {
		if len(pidBytes) == 0 {
			continue
		}

		pidStr := string(pidBytes)
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("invalid PID: %s", pidStr)
		}

		// Skip if this is our own process
		if pid == selfPid {
			continue
		}

		// Kill the process
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("failed to find process %d: %v", pid, err)
		}

		Log("Killing process %d...", pid)
		if err := process.Signal(syscall.SIGTERM); err != nil {
			Log("Error killing process %d: %v", pid, err)
			return fmt.Errorf("failed to terminate process %d: %v", pid, err)
		}
		Log("Killed process %d.", pid)
	}

	return nil
}

func _start_daemon() error {
	// Set up the pipes
	Log("Setting up pipes...")
	if err := setup_pipes(); err != nil {
		Log("Error setting up pipes: %v", err)
		return fmt.Errorf("failed to set up pipes: %v", err)
	}
	Log("Pipes set up.")

	// Open the command pipe for reading
	Log("Opening command pipe...")
	commandPipe, err := open_command_pipe_read()
	if err != nil {
		Log("Error opening command pipe: %v", err)
		return fmt.Errorf("failed to open command pipe: %v", err)
	}
	defer commandPipe.Close()
	Log("Command pipe opened.")

	Log("Daemon started. Listening for commands...")

	ActiveWindowList()

	Log("Waiting for commands on %s", command_pipe_path)
	for {
		Log("Waiting for command...")
		// Read a command from the pipe
		command, err := read_from_pipe(commandPipe)
		if err != nil {
			if err == io.EOF {
				// Pipe was closed, reopen it
				commandPipe.Close()
				commandPipe, err = open_command_pipe_read()
				if err != nil {
					Log("Error reopening command pipe: %v", err)
					return fmt.Errorf("failed to reopen command pipe: %v", err)
				}
				Log("Command pipe reopened.")
				continue
			}
			Log("Error reading from command pipe: %v", err)
			return fmt.Errorf("error reading from command pipe: %v", err)
		}

		// Prepare to write a response
		var responseErr error

		// Process the command
		switch command {
		case "HELLO":
			Log("Received HELLO command")
			responseErr = write_to_client("HELLO")
			if responseErr != nil {
				Log("Error writing HELLO response: %v", responseErr)
			}
		case "ACTIVE_WINDOW_LIST":
			Log("Received ACTIVE_WINDOW_LIST command")
			// Get the current active window list
			windows := ActiveWindowList()
			// Send the window list as JSON
			if err := _send_active_window_list(windows); err != nil {
				Log("Error sending active window list: %v", err)
			}
		case "QUIT":
			Log("Received QUIT command")
			responseErr = write_to_client("BYE")
			if responseErr != nil {
				Log("Error writing BYE response: %v", responseErr)
			}
			commandPipe.Close()
			return nil
		default:
			Log("Received unknown command: %s", command)
			responseErr = write_to_client("UNKNOWN")
			if responseErr != nil {
				Log("Error writing UNKNOWN response: %v", responseErr)
			}
		}

		// Continue to the next command
	}
}

func EnsureDaemon(restart bool) error {
	// Check if daemon is already running
	fmt.Println("Checking daemon...")
	daemon, err := _check_daemon()
	if err != nil {
		return fmt.Errorf("error checking daemon: %v", err)
	}

	// If daemon is already running, nothing to do
	if daemon && !restart {
		return nil
	}

	cmd := exec.Command("/home/dl/Projects/gofi/gofi", "-daemon")

	fmt.Println("Spawning daemon...")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting daemon: %v", err)
	}

	err = _await_daemon()
	if err != nil {
		return err
	}

	fmt.Println("Daemon started.")

	return nil
}

func _await_daemon() error {
	// Wait for both named pipes to be created
	fmt.Println("Waiting for named pipes...")
	if err := wait_for_pipes(1000); err != nil {
		return fmt.Errorf("timeout waiting for daemon: %v", err)
	}

	// Wait a bit more for the daemon to be ready
	time.Sleep(100 * time.Millisecond)

	// Try to connect to the daemon
	return HelloDaemon()
}

func _check_daemon() (bool, error) {
	// Check if both pipes exist
	if !check_pipes_exist() {
		return false, nil
	}

	// Check if there's a running daemon process
	return _find_gofi_daemon()
}

func _find_gofi_daemon() (bool, error) {
	cmd := exec.Command("pgrep", "-f", "gofi.*-daemon")
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes match
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("error checking for gofi process: %v", err)
	}
	return len(output) > 0, nil
}

func HelloDaemon() error {
	fmt.Println("Hello daemon...")

	// Send HELLO command
	fmt.Println("Sending HELLO...")
	response, err := send_and_receive_simple("HELLO")
	if err != nil {
		return fmt.Errorf("failed to communicate with daemon: %v", err)
	}

	// Check response
	if response != "HELLO" {
		return fmt.Errorf("expected HELLO response, got %s", response)
	}
	fmt.Println("Received HELLO response.")

	return nil
}

func _send_active_window_list(list []Window) error {
	Log("Sending active window list...")

	// TODO we need to swap the first two entries:
	if len(list) > 1 {
		list[0], list[1] = list[1], list[0]
	}

	// Marshal the window list to JSON
	json_str, err := marshal_window_list(list)
	if err != nil {
		Log("Error marshaling window list: %v", err)
		return fmt.Errorf("failed to marshal window list: %v", err)
	}

	// Send the JSON data
	Log("Sending JSON data...")
	if err := write_to_client(json_str); err != nil {
		Log("Error writing JSON data: %v", err)
		return fmt.Errorf("failed to write window list to pipe: %v", err)
	}

	Log("Active window list sent.")
	return nil
}

func AcquireActiveWindowList() ([]Window, error) {
	// Send the ACTIVE_WINDOW_LIST command and receive JSON response
	fmt.Println("Sending ACTIVE_WINDOW_LIST command...")
	json_response, err := send_and_receive_json("ACTIVE_WINDOW_LIST")
	if err != nil {
		return nil, fmt.Errorf("failed to communicate with daemon: %v", err)
	}

	fmt.Println("Got JSON response, length:", len(json_response))

	// Parse the JSON response
	windows, err := unmarshal_window_list(json_response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse window list: %v", err)
	}

	return windows, nil
}
