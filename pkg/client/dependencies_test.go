package client_test

import (
	"testing"

	"gofi/pkg/client"
)

func TestCheckDependencies(t *testing.T) {
	// Since we assume all commands are installed, this should pass
	if err := client.CheckDependencies(); err != nil {
		t.Errorf("CheckDependencies() failed: %v", err)
	}
}
