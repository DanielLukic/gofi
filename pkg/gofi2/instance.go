package gofi2

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

const (
	lockFileName   = "gofi2.lock"
	socketFileName = "gofi2.sock"
	ipcTimeout     = 200 * time.Millisecond
)

type InstanceManager struct {
	lockFile   *os.File
	socketPath string
	listener   net.Listener
}

func NewInstanceManager() *InstanceManager {
	tempDir := os.TempDir()
	return &InstanceManager{
		socketPath: filepath.Join(tempDir, socketFileName),
	}
}

// CheckExistingInstance checks if another instance is running
// Returns true if another instance is running, false if this is the first
func (im *InstanceManager) CheckExistingInstance() bool {
	tempDir := os.TempDir()
	lockPath := filepath.Join(tempDir, lockFileName)

	// Try to read existing lock file
	if data, err := os.ReadFile(lockPath); err == nil {
		if pid, err := strconv.Atoi(string(data)); err == nil {
			// Check if process is still running
			if process, err := os.FindProcess(pid); err == nil {
				// Try to signal the process (doesn't actually send signal, just checks existence)
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// Process exists, try to signal it
					return im.signalExistingInstance()
				}
			}
		}
	}

	// No existing instance, create lock file
	return im.createLockFile(lockPath)
}

func (im *InstanceManager) createLockFile(lockPath string) bool {
	lockFile, err := os.Create(lockPath)
	if err != nil {
		fmt.Printf("Failed to create lock file: %v\n", err)
		return false
	}

	pid := os.Getpid()
	if _, err := lockFile.WriteString(strconv.Itoa(pid)); err != nil {
		lockFile.Close()
		os.Remove(lockPath)
		fmt.Printf("Failed to write PID to lock file: %v\n", err)
		return false
	}

	im.lockFile = lockFile
	return false // This is the first instance
}

func (im *InstanceManager) signalExistingInstance() bool {
	// Try to connect to existing instance's socket with timeout
	conn, err := net.DialTimeout("unix", im.socketPath, ipcTimeout)
	if err != nil {
		// Socket doesn't exist or not accessible, assume stale lock
		return false
	}
	defer conn.Close()

	// Set timeout for the write operation
	conn.SetWriteDeadline(time.Now().Add(ipcTimeout))

	// Send "show" command
	_, err = conn.Write([]byte("show\n"))
	if err != nil {
		// Write failed, likely unresponsive process
		return false
	}

	// Optionally: try to read a response to verify the process handled the command
	conn.SetReadDeadline(time.Now().Add(ipcTimeout))
	buffer := make([]byte, 16)
	_, err = conn.Read(buffer)

	// Don't care about the response content, just that something responded
	return err == nil
}

func (im *InstanceManager) StartIPCServer(app AppInterface) error {
	// Remove existing socket if present
	os.Remove(im.socketPath)

	listener, err := net.Listen("unix", im.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create IPC socket: %w", err)
	}

	im.listener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}

			go func() {
				defer conn.Close()
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err != nil {
					return
				}

				command := string(buffer[:n])
				if command == "show\n" {
					app.Show()
					// Send acknowledgment
					conn.Write([]byte("ok\n"))
				}
			}()
		}
	}()

	return nil
}

func (im *InstanceManager) Cleanup() {
	if im.lockFile != nil {
		im.lockFile.Close()
		os.Remove(im.lockFile.Name())
	}
	if im.listener != nil {
		im.listener.Close()
		os.Remove(im.socketPath)
	}
}
