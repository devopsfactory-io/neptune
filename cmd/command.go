package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"neptune/internal/config"
	"neptune/internal/domain"
	"neptune/internal/git"
	"neptune/internal/github"
	"neptune/internal/lock"
	"neptune/internal/log"
	githubnotify "neptune/internal/notifications/github"
	"neptune/internal/run"
)

// NewCommandCmd returns the command (workflow) subcommand.
func NewCommandCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "command [workflow]",
		Short: "Run a workflow phase (e.g. plan, apply)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCommand,
	}
	return c
}

// repoRootFromEnv returns the directory containing the config file (same logic as lock's repoRoot).
func repoRootFromEnv(env map[string]string) (string, error) {
	configPath := env["NEPTUNE_CONFIG_PATH"]
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

// loadConfig loads .neptune.yaml from the default branch when not in E2E mode (for security),
// falling back to PR branch or local file. When NEPTUNE_E2E=1, reads from the local file only.
func loadConfig(env map[string]string) (*domain.NeptuneConfig, error) {
	if os.Getenv("NEPTUNE_E2E") == "1" {
		log.For("cli").Info("Config loaded from local file (E2E mode)")
		return config.Load(env)
	}
	configPath := env["NEPTUNE_CONFIG_PATH"]
	if configPath == "" {
		configPath = ".neptune.yaml"
	}
	pathForGit := filepath.Base(configPath)
	if pathForGit == "" || pathForGit == "." {
		pathForGit = ".neptune.yaml"
	}
	repoRoot, err := repoRootFromEnv(env)
	if err != nil {
		log.For("cli").Info("Config loaded from local file (could not resolve repo root)", "err", err)
		return config.Load(env)
	}
	defaultBranch, err := git.DefaultBranch(repoRoot)
	if err != nil {
		log.For("cli").Info("Config loaded from local file (could not get default branch)", "err", err)
		return config.Load(env)
	}
	_ = git.FetchBranch(repoRoot, defaultBranch) // best-effort; ShowFileFromRef may still work if ref exists
	refDefault := "origin/" + defaultBranch
	content, err := git.ShowFileFromRef(repoRoot, refDefault, pathForGit)
	if err == nil {
		log.For("cli").Info("Config loaded from default branch", "branch", defaultBranch)
		return config.LoadWithContent(env, content)
	}
	content, err = git.ShowFileFromRef(repoRoot, "HEAD", pathForGit)
	if err == nil {
		log.For("cli").Info("Config loaded from PR branch")
		return config.LoadWithContent(env, content)
	}
	return nil, fmt.Errorf(".neptune.yaml not found on default branch (%s) or PR branch (HEAD), and refusing to use local file: %w", defaultBranch, err)
}

func runCommand(_ *cobra.Command, args []string) error {
	workflow := args[0]
	ctx := context.Background()

	env, err := config.LoadEnv()
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	cfg, err := loadConfig(env)
	if err != nil {
		log.For("cli").Error("Error", "err", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		notifyAndExit(cfg, err.Error(), 1)
	}
	log.Init(cfg.LogLevel)

	// Config summary banner (after validate, like Python)
	configSummaryLines := configSummaryLines(cfg)
	log.Banner("Neptune Config Summary", configSummaryLines)

	log.Banner("Neptune Command", []string{"Neptune is running: " + workflow})

	wfs := cfg.Workflows.Workflows[cfg.Repository.AllowedWorkflow]
	if _, ok := wfs.Phases[workflow]; !ok {
		notifyAndExit(cfg, "Workflow "+workflow+" is not valid, check the allowed workflow in the Neptune config", 1)
	}
	log.For("cli").Info("Workflow " + workflow + " is valid")

	var requirements []string
	switch workflow {
	case "plan":
		requirements = cfg.Repository.PlanRequirements
	case "apply":
		requirements = cfg.Repository.ApplyRequirements
	}

	e2eMode := os.Getenv("NEPTUNE_E2E") == "1"
	var ghClient *github.Client
	var isPROpen func(context.Context, string) (bool, error)
	if e2eMode {
		isPROpen = func(_ context.Context, _ string) (bool, error) { return true, nil }
	} else {
		ghClient = github.NewClient(cfg)
		if ghClient == nil {
			notifyAndExit(cfg, "GITHUB_TOKEN, GITHUB_REPOSITORY, and GITHUB_PULL_REQUEST_NUMBER are required", 1)
		}
		status := ghClient.CheckRequirements(ctx, requirements)
		if !status.IsCompliant {
			notifyAndExit(cfg, "Cannot run "+workflow+" workflow: "+status.ErrorMessage, 1)
		}
		log.For("cli").Info("PR requirements check passed for workflow " + workflow)
		reqLine := "PR requirements (" + strings.Join(requirements, ", ") + ") check passed for workflow: " + workflow
		log.Banner("Neptune Plan/Apply Requirements Check", []string{reqLine})
		isPROpen = func(ctx context.Context, prNumber string) (bool, error) {
			return ghClient.IsPROpen(ctx, prNumber)
		}
	}
	lockIface, err := lock.NewInterface(ctx, cfg, isPROpen)
	if err != nil {
		notifyAndExit(cfg, "Failed to get changed Terraform stacks: "+err.Error(), 1)
	}
	defer func() {
		if err := lockIface.Close(); err != nil {
			log.For("cli").Error("close lock", "err", err)
		}
	}()

	if len(lockIface.TerraformStacks.Stacks) == 0 {
		notifyAndExit(cfg, "No Terraform stacks found", 0)
	}

	lockBannerLines := []string{"Neptune is considering the following stacks in the current PR:"}
	lockBannerLines = append(lockBannerLines, strings.Join(lockIface.TerraformStacks.Stacks, ", "))
	log.Banner("Neptune Lock", lockBannerLines)

	locked, err := lockIface.StacksLocked(ctx)
	if err != nil {
		notifyAndExit(cfg, "Failed to check lock status: "+err.Error(), 1)
	}
	if locked.Locked {
		msg := fmt.Sprintf("Some stacks (%v) are locked by other PRs: %v", locked.StackPath, locked.PRs)
		notifyAndExit(cfg, msg, 1)
	}

	if !lockIface.DependsOnCompleted(workflow) {
		notifyAndExit(cfg, "Dependency is not met for phase "+workflow, 1)
	}

	if err := lockIface.LockStacks(ctx, workflow, lockIface.TerraformStacks.Stacks, domain.WorkflowStatusPending); err != nil {
		notifyAndExit(cfg, "Failed to lock stacks: "+err.Error(), 1)
	}

	runURL := buildRunURL(cfg)
	var headSHA string
	if ghClient != nil {
		sha, err := ghClient.GetHeadSHA(ctx)
		if err != nil {
			log.For("cli").Error("failed to get PR head SHA for status", "err", err)
		} else {
			headSHA = sha
			ctxName := "neptune " + workflow
			pendingDesc := "Plan in progress…"
			if workflow == "apply" {
				pendingDesc = "Apply in progress…"
			}
			if err := ghClient.CreateCommitStatus(ctx, headSHA, ctxName, "pending", pendingDesc, runURL); err != nil {
				log.For("cli").Error("failed to set commit status", "context", ctxName, "err", err)
			}
		}
	}

	phase := wfs.Phases[workflow]
	runner := &run.Runner{
		Config: cfg,
		Phase:  workflow,
		Locks:  lockIface,
		Stacks: lockIface.TerraformStacks.Stacks,
		Steps:  phase.Steps,
	}
	stepsOut, err := runner.Execute(ctx)
	if err != nil {
		notifyAndExit(cfg, "Failed to run steps: "+err.Error(), 1)
	}

	stepsSummaryLines := stepsSummaryLines(workflow, stepsOut)
	log.Banner("Neptune Steps Summary", stepsSummaryLines)

	comment := &domain.PullRequestComment{
		StepsOutput:   stepsOut,
		Stacks:        lockIface.TerraformStacks,
		OverallStatus: stepsOut.OverallStatus,
	}
	notifier := githubnotify.NewNotifier(cfg)
	if notifier != nil {
		if err := notifier.CreateComment(comment); err != nil {
			log.For("cli").Error("failed to post comment", "err", err)
		}
	}

	if ghClient != nil && headSHA != "" {
		ctxName := "neptune " + workflow
		state := "success"
		desc := workflow + " completed successfully"
		if stepsOut.OverallStatus != 0 {
			state = "failure"
			desc = workflow + " failed"
		}
		if err := ghClient.CreateCommitStatus(ctx, headSHA, ctxName, state, desc, runURL); err != nil {
			log.For("cli").Error("failed to set commit status", "context", ctxName, "err", err)
		}
		if workflow == "plan" && stepsOut.OverallStatus == 0 {
			applyPendingDesc := "Waiting for status to be reported — The PR cannot be merged because the apply command was not executed with success"
			if err := ghClient.CreateCommitStatus(ctx, headSHA, "neptune apply", "pending", applyPendingDesc, runURL); err != nil {
				log.For("cli").Error("failed to set neptune apply pending status", "err", err)
			}
		}
	}

	msg := fmt.Sprintf("Workflow %s completed with status %d", workflow, stepsOut.OverallStatus)
	if stepsOut.OverallStatus != 0 {
		log.For("cli").Error("workflow failed", "msg", msg)
		for _, o := range stepsOut.Outputs {
			if o.Status != 0 {
				log.For("cli").Error("Failed command", "command", o.Command, "stdout", o.Output, "stderr", o.Error)
				break
			}
		}
		os.Exit(stepsOut.OverallStatus)
	}
	log.For("cli").Info("Success", "msg", msg)
	return nil
}

func buildRunURL(cfg *domain.NeptuneConfig) string {
	if cfg == nil || cfg.Repository == nil || cfg.Repository.GitHub == nil {
		return ""
	}
	repo := strings.TrimPrefix(cfg.Repository.GitHub.Repository, "https://github.com/")
	repo = strings.TrimSuffix(repo, "/")
	if cfg.Repository.GitHub.RunID == "" {
		return ""
	}
	return "https://github.com/" + repo + "/actions/runs/" + cfg.Repository.GitHub.RunID
}

func configSummaryLines(cfg *domain.NeptuneConfig) []string {
	repo := cfg.Repository
	if repo == nil {
		return nil
	}
	var lines []string
	if repo.GitHub != nil {
		lines = append(lines,
			"- Repository: "+repo.GitHub.Repository,
			"- Pull request branch: "+repo.GitHub.PullRequestBranch,
			"- Pull request number: "+repo.GitHub.PullRequestNumber,
		)
	}
	lines = append(lines,
		"- Object storage:",
		"  "+repo.ObjectStorage,
		"- Plan requirements: "+strings.Join(repo.PlanRequirements, ", "),
		"- Apply requirements: "+strings.Join(repo.ApplyRequirements, ", "),
		"- Allowed workflow: "+repo.AllowedWorkflow,
	)
	wf, ok := cfg.Workflows.Workflows[repo.AllowedWorkflow]
	if ok {
		var phases []string
		for p := range wf.Phases {
			phases = append(phases, p)
		}
		lines = append(lines, "- Workflow phases: "+strings.Join(phases, ", "))
	}
	return lines
}

func stepsSummaryLines(phase string, out *domain.StepsOutput) []string {
	if out == nil {
		return []string{"- Phase: " + phase}
	}
	lines := []string{"", "- Phase: " + phase, "- Steps:"}
	for _, o := range out.Outputs {
		statusStr := fmt.Sprint(o.Status)
		lines = append(lines, "  - Command: "+o.Command)
		if o.Stack != "" {
			lines = append(lines, "    - Stack: "+o.Stack)
		}
		lines = append(lines, "    - Status: "+statusStr)
	}
	return lines
}

func notifyAndExit(cfg *domain.NeptuneConfig, output string, status int) {
	notifier := githubnotify.NewNotifier(cfg)
	if notifier != nil {
		if err := notifier.CreateComment(&domain.PullRequestComment{
			SimpleOutput:  output,
			OverallStatus: status,
		}); err != nil {
			log.For("cli").Error("failed to post comment", "err", err)
		}
	}
	if status != 0 {
		log.For("cli").Error("Error", "msg", output)
		os.Exit(status)
	}
	log.For("cli").Info("Success", "msg", output)
	os.Exit(0)
}
