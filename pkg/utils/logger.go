package utils

import (
	"log/slog"
	"os"
	"strings"
)

// InitLogger initializes the global structured logger.
func InitLogger() *slog.Logger {
	level := getLogLevel()
	format := getLogFormat()

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func getLogLevel() slog.Level {
	switch strings.ToLower(os.Getenv("MOAI_LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getLogFormat() string {
	format := os.Getenv("MOAI_LOG_FORMAT")
	if format == "json" {
		return "json"
	}
	return "text"
}
