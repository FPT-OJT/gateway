package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(level string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	lvl := parseLevel(level)

	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
