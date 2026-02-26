package log

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"  INFO  ", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"ERROR", slog.LevelError},
		{"error", slog.LevelError},
		{"WARN", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"invalid", slog.LevelInfo},
	}
	for _, tt := range tests {
		got := parseLevel(tt.in)
		if got != tt.want {
			t.Errorf("parseLevel(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestInit_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	// Use a custom handler writing to buf so we can assert output.
	setHandlerForTest := func(level slog.Level) {
		h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})
		slog.SetDefault(slog.New(h))
	}
	t.Cleanup(func() {
		// Restore default so other tests are not affected
		Init("INFO")
	})

	// ERROR level: only Error should produce output
	buf.Reset()
	setHandlerForTest(slog.LevelError)
	Debug("debug msg")
	Info("info msg")
	Error("error msg")
	out := buf.String()
	if !strings.Contains(out, "error msg") {
		t.Errorf("expected Error output, got %q", out)
	}
	if strings.Contains(out, "debug msg") || strings.Contains(out, "info msg") {
		t.Errorf("ERROR level should not log Debug/Info, got %q", out)
	}

	// INFO level: Info and Error
	buf.Reset()
	setHandlerForTest(slog.LevelInfo)
	Debug("debug msg")
	Info("info msg")
	Error("error msg")
	out = buf.String()
	if !strings.Contains(out, "info msg") || !strings.Contains(out, "error msg") {
		t.Errorf("expected Info and Error output, got %q", out)
	}
	if strings.Contains(out, "debug msg") {
		t.Errorf("INFO level should not log Debug, got %q", out)
	}

	// DEBUG level: all three
	buf.Reset()
	setHandlerForTest(slog.LevelDebug)
	Debug("debug msg")
	Info("info msg")
	Error("error msg")
	out = buf.String()
	for _, want := range []string{"debug msg", "info msg", "error msg"} {
		if !strings.Contains(out, want) {
			t.Errorf("DEBUG level should log %q, got %q", want, out)
		}
	}
}

func TestInit_InvalidLevelNoPanic(t *testing.T) {
	// Invalid level should default to INFO and not panic
	Init("INVALID")
	Init("")
	// If we get here, no panic
}

func TestInit_FromEnv(t *testing.T) {
	// Ensure init() runs with a safe default when env is not set
	os.Unsetenv("NEPTUNE_LOG_LEVEL")
	// Re-run init by calling Init with default
	Init("INFO")
	Error("test error")
	// No panic and default logger works
}

func TestFor_AddsSource(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { Init("INFO") })

	For("config").Info("Loading environment variables")
	out := buf.String()
	if !strings.Contains(out, "source=neptune.config") {
		t.Errorf("expected source=neptune.config in output, got %q", out)
	}
	if !strings.Contains(out, "Loading environment variables") {
		t.Errorf("expected message in output, got %q", out)
	}
}

func TestBanner_NoPanic(t *testing.T) {
	Banner("Test", []string{"line one", "line two"})
	Banner("Short", nil)
	Banner("Long title here", []string{"body"})
}
