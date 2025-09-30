package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Initialize configures and initializes the global zerolog logger.
// If LOG_CONSOLE=1 is set, it uses colored console output.
// Otherwise, it uses standard JSON output.
// If LOG_CALLER=1 is set, it includes file and line number in log messages.
func Initialize() zerolog.Logger {
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"

	var logger zerolog.Logger
	enableCaller := os.Getenv("LOG_CALLER") == "1"

	if os.Getenv("LOG_CONSOLE") == "1" {
		// Use colored console output
		ctx := zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02T15:04:05.000Z07:00",
		}).
			With().
			Timestamp()
		if enableCaller {
			ctx = ctx.Caller()
		}
		logger = ctx.Logger()
	} else {
		// Use standard JSON output
		ctx := zerolog.New(os.Stdout).
			With().
			Timestamp()
		if enableCaller {
			ctx = ctx.Caller()
		}
		logger = ctx.Logger()
	}

	// Set global logger
	log.Logger = logger

	return logger
}
