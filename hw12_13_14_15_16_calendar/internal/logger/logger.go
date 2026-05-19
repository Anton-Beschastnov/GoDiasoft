package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	logger *slog.Logger
	level  slog.Level
}

func New(level string, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	handler := slog.NewTextHandler(output, opts)
	return &Logger{
		logger: slog.New(handler),
		level:  lvl,
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
