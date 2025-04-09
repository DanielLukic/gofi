package log_test

import (
	"testing"

	"gofi/pkg/log"
)

func TestBasicLogging(t *testing.T) {
	// Test each log level
	log.Debug("This is a debug message")
	log.Info("This is an info message")
	log.Warn("This is a warning message")
	log.Error("This is an error message")
}

func TestLoggerSetup(t *testing.T) {
	// Test logger setup with different levels
	levels := []string{
		log.LevelDebug,
		log.LevelInfo,
		log.LevelWarn,
		log.LevelError,
	}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			log.SetupLogger(level, false)
			// Test we can log at this level
			log.Info("Test message for level: %s", level)
		})
	}
}
