package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var LogFile string
var logHandle *os.File

// New initializes and returns a logger and a close function
func New(logDir string, debug bool) (zerolog.Logger, func(), error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return zerolog.Nop(), nil, err
	}

	LogFile = filepath.Join(logDir, "netmonitor.log")
	file, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return zerolog.Nop(), nil, err
	}
	logHandle = file

	// Use multi-level writer: file + console (formatted)
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multi := zerolog.MultiLevelWriter(consoleWriter, file)

	l := zerolog.New(multi).With().Timestamp().Logger()

	// Set global level (still useful if using zerolog types, but we are returning instance)
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	l.Info().Str("path", LogFile).Msg("Logger initialized")

	closeFunc := func() {
		if logHandle != nil {
			_ = logHandle.Close()
		}
	}

	return l, closeFunc, nil
}

// GetLogPath returns the absolute path to the log file
func GetLogPath() string {
	abs, err := filepath.Abs(LogFile)
	if err != nil {
		return LogFile
	}
	return abs
}
