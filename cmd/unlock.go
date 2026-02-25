package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"neptune/internal/config"
	"neptune/internal/github"
	"neptune/internal/lock"
)

// NewUnlockCmd returns the unlock subcommand.
func NewUnlockCmd() *cobra.Command {
	var allStacks bool
	c := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock all stacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !allStacks {
				return fmt.Errorf("you need to use the flag --all to run this command")
			}
			return runUnlock()
		},
	}
	c.Flags().BoolVarP(&allStacks, "all", "a", false, "Unlock all stacks (required)")
	return c
}

func runUnlock() error {
	ctx := context.Background()
	env, err := config.LoadEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	cfg, err := config.Load(env)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	ghClient := github.NewClient(cfg)
	isPROpen := func(ctx context.Context, prNumber string) (bool, error) {
		if ghClient == nil {
			return false, nil
		}
		return ghClient.IsPROpen(ctx, prNumber)
	}
	lockIface, err := lock.NewInterface(ctx, cfg, isPROpen)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer lockIface.Close()
	if err := lockIface.UnlockAllStacks(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println("Success: All changed stacks unlocked")
	return nil
}
