package cmd

import (
	"github.com/spf13/cobra"

	"github.com/devopsfactory-io/neptune/internal/log"
)

// NewRootCmd returns the root command for the Neptune CLI.
func NewRootCmd(version, commit, date string) *cobra.Command {
	root := &cobra.Command{
		Use:   "neptune",
		Short: "Neptune CLI - Terraform pull request automation tool inspired by Atlantis",
		Long:  "Neptune is a Terraform pull request automation tool. Run plans and applies on PRs with Terramate or local stack management and object-storage locking (GCS or S3).",
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().String("log-level", "", "Log level (DEBUG, INFO, ERROR); overrides config and NEPTUNE_LOG_LEVEL")
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		level, err := ParseLogLevelFlag(cmd)
		if err != nil {
			return err
		}
		if level != "" {
			log.Init(level)
		}
		return nil
	}
	root.AddCommand(NewVersionCmd(version, commit, date))
	root.AddCommand(NewCommandCmd())
	root.AddCommand(NewUnlockCmd())
	root.AddCommand(NewStacksCmd())
	return root
}
