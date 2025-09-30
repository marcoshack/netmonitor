package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Initialize configures and initializes the global zerolog logger.
// If LOG_CONSOLE=1 is set, it uses colored console output.
// Otherwise, it uses standard JSON output.
func Initialize() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var logger zerolog.Logger

	if os.Getenv("LOG_CONSOLE") == "1" {
		// Use colored console output
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		// Use standard JSON output
		logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	// Set global logger
	log.Logger = logger

	return logger
}