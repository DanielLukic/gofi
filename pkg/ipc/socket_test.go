package ipc_test

import (
	"testing"

	"gofi/pkg/ipc"
)

func TestSocket(t *testing.T) {
	// Test socket setup
	listener, err := ipc.SetupSocket()
	if err != nil {
		t.Fatalf("Failed to setup socket: %v", err)
	}
	defer listener.Close()

	// Test sending a message
	go func() {
		response, err := ipc.SendMessage("HELLO")
		if err != nil {
			t.Errorf("Failed to send message: %v", err)
			return
		}
		if response != "HELLO" {
			t.Errorf("Unexpected response: got %q, want %q", response, "HELLO")
		}
	}()

	// Accept and handle one connection
	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Failed to accept connection: %v", err)
	}
	defer conn.Close()

	// Read the message
	message, err := ipc.ReceiveMessage(conn)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}
	if message != "HELLO" {
		t.Errorf("Unexpected message: got %q, want %q", message, "HELLO")
	}

	// Send response
	if err := ipc.SendResponse(conn, "HELLO"); err != nil {
		t.Fatalf("Failed to send response: %v", err)
	}

	// Test cleanup
	if err := ipc.CleanupSocket(); err != nil {
		t.Errorf("Failed to cleanup socket: %v", err)
	}
}
