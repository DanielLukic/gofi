package client

import (
	"net"
	"testing"

	"gofi/pkg/ipc"
)

func TestSendHello(t *testing.T) {
	// Arrange
	listener, err := ipc.SetupSocket()
	if err != nil {
		t.Fatalf("Failed to setup socket: %v", err)
	}
	defer listener.Close()
	defer ipc.CleanupSocket()

	// Act
	go handleHelloRequest(t, listener)
	ok := SendHello()

	// Assert
	if !ok {
		t.Error("SendHello failed")
	}
}

func TestActiveWindowList(t *testing.T) {
	// Arrange
	listener, err := ipc.SetupSocket()
	if err != nil {
		t.Fatalf("Failed to setup socket: %v", err)
	}
	defer listener.Close()
	defer ipc.CleanupSocket()

	// Act
	go handleWindowListRequest(t, listener)
	windows := ActiveWindowList()

	// Assert
	if len(windows) != 1 {
		t.Fatalf("Expected 1 window, got %d", len(windows))
	}
	w := windows[0]
	if w.Title != "Test Window" || w.ClassName != "TestApp" {
		t.Errorf("Unexpected window data: %+v", w)
	}
}

// Helper functions
func handleHelloRequest(t *testing.T, listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		t.Errorf("Failed to accept connection: %v", err)
		return
	}
	defer conn.Close()

	msg, err := ipc.ReceiveMessage(conn)
	if err != nil {
		t.Errorf("Failed to receive message: %v", err)
		return
	}
	if msg != "HELLO" {
		t.Errorf("Expected HELLO message, got %q", msg)
		return
	}

	if err := ipc.SendResponse(conn, "HELLO"); err != nil {
		t.Errorf("Failed to send response: %v", err)
	}
}

func handleWindowListRequest(t *testing.T, listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		t.Errorf("Failed to accept connection: %v", err)
		return
	}
	defer conn.Close()

	msg, err := ipc.ReceiveMessage(conn)
	if err != nil {
		t.Errorf("Failed to receive message: %v", err)
		return
	}
	if msg != "ACTIVE_WINDOW_LIST" {
		t.Errorf("Expected ACTIVE_WINDOW_LIST message, got %q", msg)
		return
	}

	mockResponse := `[{"id":1,"title":"Test Window","class_name":"TestApp","instance":"test","type":"NORMAL","desktop":1,"pid":123}]`
	if err := ipc.SendResponse(conn, mockResponse); err != nil {
		t.Errorf("Failed to send response: %v", err)
	}
}
