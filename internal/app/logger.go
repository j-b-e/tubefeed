package app

import (
	"fmt"
	"log/slog"
	"os"
	"tubefeed/internal/config"
)

func (a App) createLogger() *slog.Logger {
	var level slog.Level
	switch config.Load().LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "info":
		fallthrough
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	logger.Info(fmt.Sprintf("logger initialized (%s)", level.String()))
	return logger
}
