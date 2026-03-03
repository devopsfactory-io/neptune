package main

import (
	"os"

	"github.com/devopsfactory-io/neptune/cmd"
	"github.com/devopsfactory-io/neptune/internal/log"
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
