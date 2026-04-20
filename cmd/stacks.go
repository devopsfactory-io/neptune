package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/devopsfactory-io/neptune/internal/config"
	"github.com/devopsfactory-io/neptune/internal/log"
	"github.com/devopsfactory-io/neptune/internal/stacks"
)

const (
	formatJSON      = "json"
	formatYAML      = "yaml"
	formatText      = "text"
	formatFormatted = "formatted"
)

var validFormats = []string{formatJSON, formatYAML, formatText, formatFormatted}

func isValidFormat(f string) bool {
	for _, v := range validFormats {
		if f == v {
			return true
		}
	}
	return false
}

// NewStacksCmd returns the stacks subcommand (list, create).
func NewStacksCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "stacks",
		Short: "Manage local stacks (list, create)",
		Long:  "List or create stacks when using stacks_management: local. For Terramate, use the Terramate CLI to list stacks.",
	}
	c.PersistentFlags().String("format", formatFormatted, "Output format: json, yaml, text, or formatted")
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}
			if !isValidFormat(format) {
				return fmt.Errorf("format must be one of: json, yaml, text, formatted")
			}
			return runStacksList(cmd, format, changed)
		},
	}
	c.Flags().BoolVar(&changed, "changed", false, "Only list stacks that have changes since the base branch")
	return c
}

func runStacksList(cmd *cobra.Command, format string, changedOnly bool) error {
	env, err := config.LoadEnvForLocal()
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
	level, err := effectiveLogLevel(cmd, cfg.LogLevel)
	if err != nil {
		return err
	}
	log.Init(level)
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
	return writeStacksListOutput(os.Stdout, format, paths)
}

// stacksListOutput is the shape for JSON/YAML list output.
type stacksListOutput struct {
	Stacks []string `json:"stacks" yaml:"stacks"`
}

// writeStacksListOutput writes the stack list to w in the requested format.
func writeStacksListOutput(w io.Writer, format string, paths []string) error {
	switch format {
	case formatText:
		for _, p := range paths {
			if _, err := fmt.Fprintln(w, p); err != nil {
				return err
			}
		}
		return nil
	case formatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stacksListOutput{Stacks: paths})
	case formatYAML:
		return yaml.NewEncoder(w).Encode(stacksListOutput{Stacks: paths})
	case formatFormatted:
		lines := []string{""}
		for _, p := range paths {
			lines = append(lines, "  - "+p)
		}
		if err := log.BannerTo(w, "Neptune Stacks", lines); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("format must be one of: json, yaml, text, formatted")
	}
}

// NewStacksCreateCmd returns the stacks create subcommand.
func NewStacksCreateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "create [path]",
		Short: "Create a new local stack (directory + stack.hcl)",
		Long:  "Create a new stack directory and a minimal stack.hcl for stacks_management: local discovery.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: neptune stacks create <path>")
			}
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}
			if !isValidFormat(format) {
				return fmt.Errorf("format must be one of: json, yaml, text, formatted")
			}
			dependsOnStr, err := cmd.Flags().GetString("depends-on")
			if err != nil {
				return err
			}
			var dependsOn []string
			if dependsOnStr != "" {
				for _, s := range strings.Split(dependsOnStr, ",") {
					if p := strings.TrimSpace(s); p != "" {
						dependsOn = append(dependsOn, p)
					}
				}
			}
			return runStacksCreate(cmd, format, args[0], dependsOn)
		},
	}
	c.Flags().String("depends-on", "", "Comma-separated list of stack paths this stack depends on (e.g. foundation/base,foundation/iam)")
	return c
}

func runStacksCreate(cmd *cobra.Command, format string, name string, dependsOn []string) error {
	env, err := config.LoadEnvForLocal()
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
	level, err := effectiveLogLevel(cmd, cfg.LogLevel)
	if err != nil {
		return err
	}
	log.Init(level)
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
	if err := os.MkdirAll(stackDir, 0755); err != nil { //nolint:gosec // G301: standard directory permissions
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	stackHcl := filepath.Join(stackDir, "stack.hcl")
	content := stackHclContent(name, dependsOn)
	if err := os.WriteFile(stackHcl, []byte(content), 0644); err != nil { //nolint:gosec // G306: standard file permissions
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	log.For("cli").Info("Created stack", "path", name, "file", stackHcl)
	return writeStackCreateOutput(os.Stdout, format, name)
}

// stackHclContent returns the contents of a stack.hcl file for the given name and optional depends_on paths.
// Attributes are aligned so the "=" signs line up (name padded to match depends_on width).
func stackHclContent(name string, dependsOn []string) string {
	const attrWidth = 10 // width of "depends_on" for alignment
	pad := func(s string) string { return s + strings.Repeat(" ", attrWidth-len(s)) }
	var b strings.Builder
	b.WriteString("// Neptune local stack: " + name + "\n")
	if len(dependsOn) == 0 {
		b.WriteString("// Optional: depends_on = [\"other-stack\"] or [\"../foundation\"] for run order.\n")
	}
	b.WriteString("stack {\n")
	b.WriteString("  " + pad("name") + " = " + fmt.Sprintf("%q", name) + "\n")
	if len(dependsOn) > 0 {
		quoted := make([]string, len(dependsOn))
		for i, p := range dependsOn {
			quoted[i] = fmt.Sprintf("%q", p)
		}
		b.WriteString("  " + pad("depends_on") + " = [" + strings.Join(quoted, ", ") + "]\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// stackCreateOutput is the shape for JSON/YAML create output.
type stackCreateOutput struct {
	Path string `json:"path" yaml:"path"`
}

// writeStackCreateOutput writes the created stack path to w in the requested format.
func writeStackCreateOutput(w io.Writer, format string, path string) error {
	switch format {
	case formatText:
		if _, err := fmt.Fprintln(w, path); err != nil {
			return err
		}
		return nil
	case formatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stackCreateOutput{Path: path})
	case formatYAML:
		return yaml.NewEncoder(w).Encode(stackCreateOutput{Path: path})
	case formatFormatted:
		lines := []string{"", "  - Path: " + path}
		if err := log.BannerTo(w, "Neptune Stack Created", lines); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("format must be one of: json, yaml, text, formatted")
	}
}
