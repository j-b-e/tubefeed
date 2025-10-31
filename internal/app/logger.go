package app

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"tubefeed/internal/config"
)

var level = new(slog.LevelVar)

func createLogger() *slog.Logger {

	logLevel := strings.ToLower(config.GetEnvOrDefault("LOG_LEVEL", "info"))

	switch logLevel {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	options := slog.HandlerOptions{Level: level}

	var handler slog.Handler
	logFormat := strings.ToLower(config.GetEnvOrDefault("LOG_FORMAT", "text"))
	switch logFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &options)
	default:
		handler = slog.NewTextHandler(os.Stdout, &options)
	}

	logger := slog.New(handler)
	logger.Info(fmt.Sprintf("logger initialized (%s)", level))
	return logger
}
