package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Constants for log levels matching Python's
const (
	// Log level "off" - no logging
	LevelOff = "off"
	// Log level "error" - log errors only
	LevelError = "error"
	// Log level "warning" - log warnings and errors
	LevelWarn = "warning"
	// Log level "info" - log info, warnings, and errors
	LevelInfo = "info"
	// Log level "debug" - log debug, info, warnings, and errors
	LevelDebug = "debug"

	// Default log file path
	LogFilePath = "/tmp/gofi.log"

	// Maximum file size (16KB)
	MaxFileSize = 16 * 1024
)

// LevelMap maps string levels to logrus levels
var LevelMap = map[string]logrus.Level{
	LevelOff:   logrus.Level(7), // Custom level higher than Fatal
	LevelError: logrus.ErrorLevel,
	LevelWarn:  logrus.WarnLevel,
	LevelInfo:  logrus.InfoLevel,
	LevelDebug: logrus.DebugLevel,
}

// logger is the singleton logger instance
var (
	logger *logrus.Logger
	once   sync.Once
)

// getLogger returns the singleton logger instance
// This function is safe to call from multiple goroutines
func getLogger() *logrus.Logger {
	once.Do(func() {
		logger = logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			DisableColors:   true,
			DisableSorting:  true,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
		logger.SetOutput(os.Stdout)
	})

	return logger
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	getLogger().Debug(fmt.Sprintf(format, args...))
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	getLogger().Info(fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	getLogger().Warn(fmt.Sprintf(format, args...))
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	getLogger().Error(fmt.Sprintf(format, args...))
}

// SetupLogger configures the logger with the specified settings
// Args:
//
//	logLevel: Logging level ("off", "error", "warning", "info", "debug")
//	isDaemon: Whether this is a daemon process
func SetupLogger(logLevel string, isDaemon bool) {
	l := getLogger()

	// Set log level
	level, ok := LevelMap[strings.ToLower(logLevel)]
	if !ok {
		level = logrus.InfoLevel
	}
	l.SetLevel(level)

	// Add file output if enabled
	if LogFilePath != "" {
		if err := setupLogFile(l, isDaemon); err != nil {
			Error("Failed to setup log file: %s", err)
		}
	}
}

// setupLogFile configures the logger to write to a file and console
func setupLogFile(l *logrus.Logger, isDaemon bool) error {
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(LogFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file
	file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Check file size and truncate if needed
	info, err := file.Stat()
	if err == nil && info.Size() >= MaxFileSize {
		// Use fmt.Fprintf to stderr for warnings during logger setup
		fmt.Fprintf(os.Stderr, "[WARN] Log file size limit reached (%d bytes). Truncating %s to zero.\n", info.Size(), LogFilePath)
		// Truncate the file to 0 bytes
		if err := file.Truncate(0); err != nil {
			// Close the file before returning error, as we might not be able to use it
			file.Close()
			return fmt.Errorf("failed to truncate log file: %w", err)
		}
		// Reset the write offset to the beginning of the file
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			file.Close()
			return fmt.Errorf("failed to seek to start after truncate: %w", err)
		}
		// Log file is now empty and ready for new writes
	}

	// Determine process type
	processType := "CLIENT"
	if isDaemon {
		processType = "DAEMON"
	}
	pid := os.Getpid()

	// Create a multi-writer to write to both original output and file
	multiWriter := io.MultiWriter(l.Out, file)

	// Create custom formatter
	formatter := &customFormatter{
		ProcessType: processType,
		PID:         pid,
		isDaemon:    isDaemon,
	}

	// Add file output with custom writer, keep original output
	l.SetOutput(multiWriter)
	l.SetFormatter(formatter)

	return nil
}

// prefixWriter adds a prefix to each log line
/*
type prefixWriter struct {
	file        *os.File
	processType string
	pid         int
}

func (w *prefixWriter) Write(p []byte) (n int, err error) {
	prefix := fmt.Sprintf("[%s:%d] ", w.processType, w.pid)
	// Write prefix
	_, err = w.file.WriteString(prefix)
	if err != nil {
		return 0, err
	}
	// Write original message
	return w.file.Write(p)
}
*/

// --- Custom Formatter ---

type customFormatter struct {
	ProcessType string
	PID         int
	isDaemon    bool
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := strings.ToUpper(entry.Level.String())
	if len(level) > 1 {
		level = level[:1] // Use first letter for level (D, I, W, E)
	}

	prefix := fmt.Sprintf("%s [%s][%d] %s\n",
		entry.Time.Format("15:04:05"),
		level,
		f.PID,
		entry.Message,
	)

	if f.isDaemon {
		prefix = fmt.Sprintf("%s [DAEMON:%d][%s] %s\n",
			entry.Time.Format("15:04:05"),
			f.PID,
			level,
			entry.Message,
		)
	}

	return []byte(prefix), nil
}

// --- Setup ---
