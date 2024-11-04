// logger/logger.go
package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup initializes the logger with the given environment
func Setup(environment string) {
	// Set global logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(getLogLevel(environment))

	// Create multi writer for console output
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			return fmt.Sprintf("| %-6s|", i)
		},
	}

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Str("environment", environment).
		Logger()
}

// getLogLevel returns the appropriate log level based on environment
func getLogLevel(environment string) zerolog.Level {
	switch environment {
	case "development":
		return zerolog.DebugLevel
	case "test":
		return zerolog.DebugLevel
	default:
		return zerolog.InfoLevel
	}
}
