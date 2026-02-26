package main

import (
	"os"

	"neptune/cmd"
	"neptune/internal/log"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cmd.NewRootCmd(version, commit, date)
	if err := rootCmd.Execute(); err != nil {
		log.Error("Error", "err", err)
		os.Exit(1)
	}
}
