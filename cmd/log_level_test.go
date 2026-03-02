package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestEffectiveLogLevel_FlagEmpty_ReturnsConfig(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	root.AddCommand(child)
	// Don't set --log-level; default is "" so GetString returns "" and we use fromConfig.
	level, err := effectiveLogLevel(child, "ERROR")
	if err != nil {
		t.Fatalf("effectiveLogLevel: %v", err)
	}
	if level != "ERROR" {
		t.Errorf("expected ERROR (from config), got %q", level)
	}
}

func TestEffectiveLogLevel_FlagSet_ReturnsFlagValue(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	// Simulate user having passed --log-level=DEBUG by setting the flag value.
	// Persistent flags are inherited; we need the child to have root as parent for Root() to work.
	root.AddCommand(child)
	if err := root.PersistentFlags().Set("log-level", "DEBUG"); err != nil {
		t.Fatalf("Set log-level: %v", err)
	}
	level, err := effectiveLogLevel(child, "ERROR")
	if err != nil {
		t.Fatalf("effectiveLogLevel: %v", err)
	}
	if level != "DEBUG" {
		t.Errorf("expected DEBUG, got %q", level)
	}
}

func TestEffectiveLogLevel_FlagSetCaseInsensitive(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	root.AddCommand(child)
	if err := root.PersistentFlags().Set("log-level", "info"); err != nil {
		t.Fatalf("Set log-level: %v", err)
	}
	level, err := effectiveLogLevel(child, "ERROR")
	if err != nil {
		t.Fatalf("effectiveLogLevel: %v", err)
	}
	if level != "INFO" {
		t.Errorf("expected INFO (normalized), got %q", level)
	}
}

func TestEffectiveLogLevel_FlagInvalid_ReturnsError(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	root.AddCommand(child)
	if err := root.PersistentFlags().Set("log-level", "TRACE"); err != nil {
		t.Fatalf("Set log-level: %v", err)
	}
	_, err := effectiveLogLevel(child, "ERROR")
	if err == nil {
		t.Error("expected error for invalid log-level TRACE")
	}
	if err != nil && err.Error() != "log-level must be one of: DEBUG, INFO, ERROR" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseLogLevelFlag_Empty_ReturnsEmpty(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	root.AddCommand(child)
	level, err := ParseLogLevelFlag(child)
	if err != nil {
		t.Fatalf("ParseLogLevelFlag: %v", err)
	}
	if level != "" {
		t.Errorf("expected empty when flag unset, got %q", level)
	}
}

func TestParseLogLevelFlag_Valid_ReturnsNormalized(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	root.AddCommand(child)
	if err := root.PersistentFlags().Set("log-level", "debug"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	level, err := ParseLogLevelFlag(child)
	if err != nil {
		t.Fatalf("ParseLogLevelFlag: %v", err)
	}
	if level != "DEBUG" {
		t.Errorf("expected DEBUG, got %q", level)
	}
}

func TestParseLogLevelFlag_Invalid_ReturnsError(t *testing.T) {
	root := NewRootCmd("", "", "")
	child := &cobra.Command{}
	root.AddCommand(child)
	if err := root.PersistentFlags().Set("log-level", "WARN"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	level, err := ParseLogLevelFlag(child)
	if err == nil {
		t.Error("expected error for invalid log-level WARN")
	}
	if level != "" {
		t.Errorf("expected empty level on error, got %q", level)
	}
}

// TestPreRunE_InvalidLogLevel_ReturnsError ensures the root PersistentPreRunE runs and
// returns an error when --log-level is set to an invalid value.
func TestPreRunE_InvalidLogLevel_ReturnsError(t *testing.T) {
	root := NewRootCmd("", "", "")
	root.SetArgs([]string{"--log-level", "TRACE", "version"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --log-level=TRACE")
	}
	if err != nil && err.Error() != "log-level must be one of: DEBUG, INFO, ERROR" {
		t.Errorf("unexpected error: %v", err)
	}
}
