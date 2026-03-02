package log

import (
	"log/slog"
	"os"
	"strings"
)

func init() {
	level := os.Getenv("NEPTUNE_LOG_LEVEL")
	if level == "" {
		level = "ERROR"
	}
	Init(level)
}

// Init sets the global default logger to the given level (DEBUG, INFO, ERROR; case-insensitive).
// Invalid levels fall back to INFO.
func Init(level string) {
	lvl := parseLevel(level)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(h))
}

func parseLevel(s string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO", "WARN":
		return slog.LevelInfo
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug logs at DEBUG level.
func Debug(msg string, args ...any) {
	slog.Default().Debug(msg, args...)
}

// Info logs at INFO level.
func Info(msg string, args ...any) {
	slog.Default().Info(msg, args...)
}

// Error logs at ERROR level.
func Error(msg string, args ...any) {
	slog.Default().Error(msg, args...)
}

// Logger is a logger that includes a source attribute (e.g. neptune.config) on every message.
type Logger struct {
	l *slog.Logger
}

// For returns a logger that adds source "neptune.<source>" to all messages.
func For(source string) *Logger {
	return &Logger{l: slog.Default().With("source", "neptune."+source)}
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, args ...any) {
	l.l.Debug(msg, args...)
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, args ...any) {
	l.l.Info(msg, args...)
}

// Error logs at ERROR level.
func (l *Logger) Error(msg string, args ...any) {
	l.l.Error(msg, args...)
}
