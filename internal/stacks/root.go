package stacks

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetRepoRoot returns the directory containing the Neptune config file (working root).
// Prefers the directory containing the file named by NEPTUNE_CONFIG_PATH when set.
func GetRepoRoot() (string, error) {
	configPath := os.Getenv("NEPTUNE_CONFIG_PATH")
	if configPath == "" {
		configPath = ".neptune.yaml"
	}
	path := filepath.Clean(configPath)
	if filepath.IsAbs(path) {
		return filepath.Dir(path), nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return filepath.Join(wd, filepath.Dir(path)), nil
}
