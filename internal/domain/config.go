package domain

// WorkflowStep is a single step in a workflow phase.
type WorkflowStep struct {
	Run string `yaml:"run"`
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
	Repository           string
	PullRequestBranch    string
	PullRequestNumber    string
	PullRequestCommentID string
	RunID                string
	Token                string
}

// RepositoryConfig is the repository section of .neptune.yaml.
type RepositoryConfig struct {
	ObjectStorage     string        `yaml:"object_storage"`
	Branch            string        `yaml:"branch"`
	PlanRequirements  []string      `yaml:"plan_requirements"`
	ApplyRequirements []string      `yaml:"apply_requirements"`
	AllowedWorkflow   string        `yaml:"allowed_workflow"`
	GitHub            *GitHubConfig `yaml:"-"`
}

// NeptuneConfig is the full loaded config.
type NeptuneConfig struct {
	Repository *RepositoryConfig
	Workflows  *Workflows
}
