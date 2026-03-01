package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"neptune/internal/config"
	"neptune/internal/log"
	"neptune/internal/stacks"
)

// NewStacksCmd returns the stacks subcommand (list, create).
func NewStacksCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "stacks",
		Short: "Manage local stacks (list, create)",
		Long:  "List or create stacks when using stacks_management: local. For Terramate, use the Terramate CLI to list stacks.",
	}
	c.AddCommand(NewStacksListCmd())
	c.AddCommand(NewStacksCreateCmd())
	return c
}

// NewStacksListCmd returns the stacks list subcommand.
func NewStacksListCmd() *cobra.Command {
	var changed bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List stack paths",
		Long:  "List stack paths (from config or stack.hcl discovery). With --changed, only list stacks with changes since the base branch.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStacksList(changed)
		},
	}
	c.Flags().BoolVar(&changed, "changed", false, "Only list stacks that have changes since the base branch")
	return c
}

func runStacksList(changedOnly bool) error {
	env, err := config.LoadEnv()
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	cfg, err := config.Load(env)
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	log.Init(cfg.LogLevel)
	sm := strings.TrimSpace(strings.ToLower(cfg.Repository.StacksManagement))
	if sm == "" {
		sm = "terramate"
	}
	if sm != "local" {
		fmt.Fprintln(os.Stderr, "stacks list is only available when repository.stacks_management is local. Use the Terramate CLI to list stacks when using Terramate.")
		os.Exit(1)
	}
	var paths []string
	if changedOnly {
		provider := stacks.NewProvider(cfg)
		result, err := provider.ChangedStacks(context.Background(), cfg)
		if err != nil {
			log.For("cli").Error("Error", "err", err)
			os.Exit(1)
		}
		paths = result.Stacks
	} else {
		paths, err = stacks.ListAllLocal(cfg)
		if err != nil {
			log.For("cli").Error("Error", "err", err)
			os.Exit(1)
		}
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}

// NewStacksCreateCmd returns the stacks create subcommand.
func NewStacksCreateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create [stack-name]",
		Short: "Create a new local stack (directory + stack.hcl)",
		Long:  "Create a new stack directory and a minimal stack.hcl for stacks_management: local discovery.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: neptune stacks create <stack-name>")
			}
			return runStacksCreate(args[0])
		},
	}
	return c
}

func runStacksCreate(name string) error {
	env, err := config.LoadEnv()
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	cfg, err := config.Load(env)
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	log.Init(cfg.LogLevel)
	sm := strings.TrimSpace(strings.ToLower(cfg.Repository.StacksManagement))
	if sm == "" {
		sm = "terramate"
	}
	if sm != "local" {
		fmt.Fprintln(os.Stderr, "stacks create is only available when repository.stacks_management is local.")
		os.Exit(1)
	}
	rootDir, err := stacks.GetRepoRoot()
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	stackDir := filepath.Join(rootDir, name)
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	stackHcl := filepath.Join(stackDir, "stack.hcl")
	content := fmt.Sprintf("// Neptune local stack: %s\nstack {\n  name = %q\n}\n", name, name)
	if err := os.WriteFile(stackHcl, []byte(content), 0644); err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	log.For("cli").Info("Created stack", "path", name, "file", stackHcl)
	fmt.Println(name)
	return nil
}
