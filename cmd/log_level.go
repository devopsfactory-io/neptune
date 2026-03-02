package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var allowedLogLevels = map[string]bool{"DEBUG": true, "INFO": true, "ERROR": true}

// ParseLogLevelFlag reads the root --log-level flag. It returns the normalized level (DEBUG, INFO, ERROR)
// when the flag is set and valid, "" and nil when unset, and "" and an error when set but invalid.
// Use this in root PersistentPreRunE so the level is applied before config load.
func ParseLogLevelFlag(cmd *cobra.Command) (string, error) {
	flagVal, err := cmd.Root().PersistentFlags().GetString("log-level")
	if err != nil {
		return "", err
	}
	flagVal = strings.TrimSpace(flagVal)
	if flagVal == "" {
		return "", nil
	}
	upper := strings.ToUpper(flagVal)
	if !allowedLogLevels[upper] {
		return "", fmt.Errorf("log-level must be one of: DEBUG, INFO, ERROR")
	}
	return upper, nil
}

// effectiveLogLevel returns the log level to use: the --log-level flag value if set and valid,
// otherwise fromConfig. It overrides config and NEPTUNE_LOG_LEVEL.
func effectiveLogLevel(cmd *cobra.Command, fromConfig string) (string, error) {
	level, err := ParseLogLevelFlag(cmd)
	if err != nil {
		return "", err
	}
	if level != "" {
		return level, nil
	}
	return fromConfig, nil
}
