package internal

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func InitLogger(level slog.Level) {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
}

func GetLogger() *slog.Logger {
	if logger == nil {
		InitLogger(slog.LevelInfo)
	}
	return logger
}
