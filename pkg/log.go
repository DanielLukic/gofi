package pkg

import (
	"fmt"
	"os"
)

const (
	LogPath = "daemon.log"
)

func Log(format string, args ...interface{}) {
	// Format the log message
	msg := fmt.Sprintf(format+"\n", args...)

	// Open log file in append mode
	log_file, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to stderr if we can't open the log file
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		fmt.Fprint(os.Stderr, msg)
		return
	}
	defer log_file.Close()

	// Write to log file
	_, err = log_file.WriteString(msg)
	if err != nil {
		// Fall back to stderr if we can't write to the log file
		fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
		fmt.Fprint(os.Stderr, msg)
	}
}
