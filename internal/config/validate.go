package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/log"
)

var allowedRequirements = map[string]bool{
	"undiverged": true,
	"approved":   true,
	"mergeable":  true,
	"rebased":    true,
}

var allowedLogLevels = map[string]bool{"DEBUG": true, "INFO": true, "ERROR": true}

// Validate checks the loaded config. Returns an error suitable for GitHub comment if validation fails.
func Validate(cfg *domain.NeptuneConfig) error {
	log.For("config").Info("Checking config options")
	repo := cfg.Repository
	if repo == nil {
		return errors.New("repository config is required")
	}
	if repo.ObjectStorage == "" {
		return errors.New("repository object storage locking is required")
	}
	if !strings.HasPrefix(repo.ObjectStorage, "gs://") && !strings.HasPrefix(repo.ObjectStorage, "s3://") {
		return errors.New("repository object storage must be a valid GCS (gs://) or S3 (s3://) URL")
	}
	stacksMgmt := strings.TrimSpace(strings.ToLower(repo.StacksManagement))
	if stacksMgmt == "" {
		stacksMgmt = "terramate"
	}
	if stacksMgmt != "terramate" && stacksMgmt != "local" {
		return errors.New("repository stacks_management must be terramate or local")
	}
	if stacksMgmt == "local" && repo.LocalStacks != nil && strings.TrimSpace(strings.ToLower(repo.LocalStacks.Source)) == "config" {
		if len(repo.LocalStacks.Stacks) == 0 {
			return errors.New("local_stacks.source is config but local_stacks.stacks is empty")
		}
	}
	for _, r := range repo.PlanRequirements {
		if !allowedRequirements[r] {
			return fmt.Errorf("repository plan requirements must be one of: undiverged, approved, mergeable, rebased")
		}
	}
	for _, r := range repo.ApplyRequirements {
		if !allowedRequirements[r] {
			return fmt.Errorf("repository apply requirements must be one of: undiverged, approved, mergeable, rebased")
		}
	}
	if repo.AllowedWorkflow == "" {
		return errors.New("repository allowed workflow is required")
	}
	if cfg.Workflows == nil || cfg.Workflows.Workflows == nil {
		return errors.New("workflows are required")
	}
	if _, ok := cfg.Workflows.Workflows[repo.AllowedWorkflow]; !ok {
		return fmt.Errorf("repository allowed workflow must be one of: %s", keys(cfg.Workflows.Workflows))
	}
	if repo.Branch == "" {
		return errors.New("repository branch is required, check the GitHub Action configuration")
	}
	if repo.GitHub != nil && repo.GitHub.PullRequestBranch == repo.Branch {
		return errors.New("the repository.branch (default branch) should not be used to execute the workflows; run workflows in the pull request branch")
	}
	if cfg.LogLevel != "" && !allowedLogLevels[strings.ToUpper(strings.TrimSpace(cfg.LogLevel))] {
		return errors.New("log_level must be one of: DEBUG, INFO, ERROR")
	}

	for _, wf := range cfg.Workflows.Workflows {
		for phaseName, phase := range wf.Phases {
			for _, dep := range phase.DependsOn {
				if _, ok := wf.Phases[dep]; !ok {
					return fmt.Errorf("phase %s depends on %s, but %s is not a valid workflow phase", phaseName, dep, dep)
				}
			}
			if _, hasPlan := wf.Phases["plan"]; !hasPlan {
				return errors.New("phases should include at least plan and apply phases")
			}
			if _, hasApply := wf.Phases["apply"]; !hasApply {
				return errors.New("phases should include at least plan and apply phases")
			}
			for _, step := range phase.Steps {
				if step.Run == "" {
					return errors.New("at least one step is required in each phase")
				}
				run := step.Run
				// When once is true, the run string is executed once in repo root; it must then use terramate CLI and --changed if it runs terraform/terragrunt.
				// When once is false or nil (default), Neptune runs the command per stack; no need for "terramate" or "--changed" in run.
				runOnceInRoot := step.Once != nil && *step.Once
				if runOnceInRoot && (strings.Contains(run, "terragrunt") || strings.Contains(run, "terraform")) {
					if !strings.Contains(run, "terramate") || !strings.Contains(run, "--changed") {
						return errors.New("the step run must use both the terramate command AND the --changed flag when using terragrunt or terraform with once: true (or set once: false so Neptune runs the command per stack)")
					}
				}
			}
		}
	}
	return nil
}

func keys(m map[string]domain.WorkflowStatement) string {
	var k []string
	for s := range m {
		k = append(k, s)
	}
	return strings.Join(k, ", ")
}
