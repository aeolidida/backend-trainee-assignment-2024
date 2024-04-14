package logger

import (
	"backend-trainee-assignment-2024/internal/config"
	"os"

	"log/slog"
)

const (
	debug = "debug"
	info  = "info"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func New(cfg config.Logger) (Logger, error) {
	var log Logger

	if cfg.LogLevel == debug {
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	} else if cfg.LogLevel == info {
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log, nil
}
