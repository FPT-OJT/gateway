package logger_test

import (
	"testing"

	"github.com/FPT-OJT/gateway/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNew_InfoLevel(t *testing.T) {
	log := logger.New("info")
	assert.Equal(t, zerolog.InfoLevel, log.GetLevel())
}

func TestNew_DebugLevel(t *testing.T) {
	log := logger.New("debug")
	assert.Equal(t, zerolog.DebugLevel, log.GetLevel())
}

func TestNew_WarnLevel(t *testing.T) {
	log := logger.New("warn")
	assert.Equal(t, zerolog.WarnLevel, log.GetLevel())
}

func TestNew_ErrorLevel(t *testing.T) {
	log := logger.New("error")
	assert.Equal(t, zerolog.ErrorLevel, log.GetLevel())
}

func TestNew_UnknownLevel_DefaultsToInfo(t *testing.T) {
	log := logger.New("verbose")
	assert.Equal(t, zerolog.InfoLevel, log.GetLevel())
}

func TestNew_EmptyString_DefaultsToInfo(t *testing.T) {
	log := logger.New("")
	assert.Equal(t, zerolog.InfoLevel, log.GetLevel())
}

func TestNew_UpperCaseLevel_StillParsed(t *testing.T) {
	// parseLevel uses strings.ToLower internally, so "DEBUG" → "debug" → DebugLevel
	log := logger.New("DEBUG")
	assert.Equal(t, zerolog.DebugLevel, log.GetLevel())
}

func TestNew_ReturnsZerologLogger(t *testing.T) {
	log := logger.New("info")
	// Verify it is a usable logger (no panic)
	assert.NotPanics(t, func() {
		log.Info().Msg("test message")
	})
}
