package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCmd returns the version command.
func NewVersionCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of the tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("neptune version: %s (commit: %s, date: %s)\n", version, commit, date)
			return nil
		},
	}
}
