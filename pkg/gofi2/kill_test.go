package gofi2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKillInstance(t *testing.T) {
	tempDir := os.TempDir()
	lockPath := filepath.Join(tempDir, lockFileName)
	socketPath := filepath.Join(tempDir, socketFileName)

	os.WriteFile(lockPath, []byte("99999"), 0644)
	os.WriteFile(socketPath, []byte("test"), 0644)

	KillInstance()

	if _, err := os.Stat(lockPath); err == nil {
		t.Error("Lock file should be removed")
	}

	if _, err := os.Stat(socketPath); err == nil {
		t.Error("Socket file should be removed")
	}
}
