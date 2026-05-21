package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// LoggerIface интерфейс для логгера
type LoggerIface interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Logger struct {
	logger *log.Logger
	level  Level
}

func New(level string, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		logger: log.New(output, "", log.LstdFlags),
		level:  parseLevel(level),
	}
}

func parseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	if l.level <= LevelDebug {
		l.log("DEBUG", msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	if l.level <= LevelInfo {
		l.log("INFO", msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...any) {
	if l.level <= LevelWarn {
		l.log("WARN", msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	if l.level <= LevelError {
		l.log("ERROR", msg, args...)
	}
}

func (l *Logger) log(level, msg string, args ...any) {
	if len(args) > 0 {
		l.logger.Printf("[%s] %s %v", level, msg, formatArgs(args))
	} else {
		l.logger.Printf("[%s] %s", level, msg)
	}
}

func formatArgs(args []any) string {
	if len(args) == 0 {
		return ""
	}
	var result string
	for i := 0; i < len(args)-1; i += 2 {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%v=%v", args[i], args[i+1])
	}
	return result
}
