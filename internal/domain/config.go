package domain

// WorkflowStep is a single step in a workflow phase.
// When Once is false or nil (default), Neptune runs the command once per changed stack (in run order) with CWD set to each stack.
// When Once is true, the command runs once in the process CWD (repo root).
type WorkflowStep struct {
	Run  string `yaml:"run"`
	Once *bool  `yaml:"once"` // true = run once in repo root; false or unset = run in each changed stack
}

// WorkflowPhase is a phase in a workflow (e.g. plan or apply).
type WorkflowPhase struct {
	Steps     []WorkflowStep `yaml:"steps"`
	DependsOn []string       `yaml:"depends_on"`
}

// WorkflowStatement is a full workflow with phases.
type WorkflowStatement struct {
	Name   string                   `yaml:"-"`
	Phases map[string]WorkflowPhase `yaml:"-"` // keyed by phase name
}

// Workflows holds named workflows.
type Workflows struct {
	Workflows map[string]WorkflowStatement `yaml:"-"` // populated from YAML
}

// GitHubConfig holds GitHub-related config from env.
type GitHubConfig struct {
	Repository        string
	PullRequestBranch string
	PullRequestNumber string
	RunID             string
	Token             string
}

// StackEntry is a single stack entry for local stacks_management (config-based).
type StackEntry struct {
	Path      string   `yaml:"path"`
	DependsOn []string `yaml:"depends_on"`
}

// LocalStacksConfig is the optional config for stacks_management: local.
type LocalStacksConfig struct {
	Source string       `yaml:"source"` // "config" or "discovery"
	Stacks []StackEntry `yaml:"stacks"` // used when source is "config"
}

// RepositoryConfig is the repository section of .neptune.yaml.
type RepositoryConfig struct {
	ObjectStorage     string             `yaml:"object_storage"`
	StacksManagement  string             `yaml:"stacks_management"` // "terramate" (default) or "local"
	LocalStacks       *LocalStacksConfig `yaml:"-"`                 // populated from root-level local_stacks key when stacks_management is "local"
	Branch            string             `yaml:"branch"`
	PlanRequirements  []string           `yaml:"plan_requirements"`
	ApplyRequirements []string           `yaml:"apply_requirements"`
	AllowedWorkflow   string             `yaml:"allowed_workflow"`
	Automerge         bool               `yaml:"automerge"` // when true, enable PR auto-merge after successful apply
	GitHub            *GitHubConfig      `yaml:"-"`
}

// NeptuneConfig is the full loaded config.
type NeptuneConfig struct {
	Repository *RepositoryConfig
	Workflows  *Workflows
	// LogLevel is the effective log level (DEBUG, INFO, ERROR); env NEPTUNE_LOG_LEVEL overrides config log_level.
	LogLevel string
}
