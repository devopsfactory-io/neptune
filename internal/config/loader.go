package config

import (
	"os"
	"path/filepath"
	"strings"

	"neptune/internal/domain"
	"neptune/internal/log"

	"gopkg.in/yaml.v3"
)

// Required env vars (all must be set).
var requiredEnvVars = []string{
	"NEPTUNE_CONFIG_PATH",
	"GITHUB_REPOSITORY",
	"GITHUB_PULL_REQUEST_BRANCH",
	"GITHUB_PULL_REQUEST_NUMBER",
	"GITHUB_RUN_ID",
	"GITHUB_TOKEN",
}

type rawRepository struct {
	ObjectStorage     string   `yaml:"object_storage"`
	Branch            string   `yaml:"branch"`
	PlanRequirements  []string `yaml:"plan_requirements"`
	ApplyRequirements []string `yaml:"apply_requirements"`
	AllowedWorkflow   string   `yaml:"allowed_workflow"`
	Automerge         bool     `yaml:"automerge"`
}

type rawStep struct {
	Run       string `yaml:"run"`
	Terramate *bool  `yaml:"terramate"`
	Changed   *bool  `yaml:"changed"`
}

type rawPhase struct {
	Steps     []rawStep `yaml:"steps"`
	DependsOn []string  `yaml:"depends_on"`
}

type rawConfig struct {
	LogLevel   string                         `yaml:"log_level"`
	Repository rawRepository                  `yaml:"repository"`
	Workflows  map[string]map[string]rawPhase `yaml:"workflows"`
}

// LoadEnv loads required environment variables. Returns a map of all required vars or error.
// When NEPTUNE_E2E=1, GitHub vars may be empty and get default values so e2e tests can run without a real token.
func LoadEnv() (map[string]string, error) {
	log.For("config").Info("Loading environment variables")
	env := make(map[string]string)
	e2eMode := os.Getenv("NEPTUNE_E2E") == "1"

	env["NEPTUNE_CONFIG_PATH"] = getEnv("NEPTUNE_CONFIG_PATH", ".neptune.yaml")
	env["GITHUB_REPOSITORY"] = getEnv("GITHUB_REPOSITORY", "e2e/neptune-test")
	env["GITHUB_PULL_REQUEST_BRANCH"] = getEnv("GITHUB_PULL_REQUEST_BRANCH", "pr-1")
	env["GITHUB_PULL_REQUEST_NUMBER"] = getEnv("GITHUB_PULL_REQUEST_NUMBER", "1")
	env["GITHUB_RUN_ID"] = getEnv("GITHUB_RUN_ID", "1")
	env["GITHUB_TOKEN"] = os.Getenv("GITHUB_TOKEN")

	var missing []string
	for _, k := range requiredEnvVars {
		if env[k] != "" {
			continue
		}
		if e2eMode {
			continue
		}
		missing = append(missing, k)
	}
	if len(missing) > 0 {
		return nil, &LoadError{Message: "environment variables " + joinQuoted(missing) + " are required"}
	}
	return env, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads config from env and the YAML file. Caller should then call Validate.
func Load(env map[string]string) (*domain.NeptuneConfig, error) {
	configPath := env["NEPTUNE_CONFIG_PATH"]
	if configPath == "" {
		configPath = ".neptune.yaml"
	}
	log.For("config").Info("Loading config file")
	path := filepath.Clean(configPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &LoadError{Message: "config file not found: " + configPath}
	}
	return parseConfig(data, env)
}

// LoadWithContent reads config from env and the given YAML content (e.g. from git show).
// Caller should then call Validate.
func LoadWithContent(env map[string]string, content []byte) (*domain.NeptuneConfig, error) {
	return parseConfig(content, env)
}

func parseConfig(data []byte, env map[string]string) (*domain.NeptuneConfig, error) {
	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, &LoadError{Message: "invalid YAML: " + err.Error()}
	}

	githubCfg := &domain.GitHubConfig{
		Repository:        env["GITHUB_REPOSITORY"],
		PullRequestBranch: env["GITHUB_PULL_REQUEST_BRANCH"],
		PullRequestNumber: env["GITHUB_PULL_REQUEST_NUMBER"],
		RunID:             env["GITHUB_RUN_ID"],
		Token:             env["GITHUB_TOKEN"],
	}

	repo := &domain.RepositoryConfig{
		ObjectStorage:     raw.Repository.ObjectStorage,
		Branch:            raw.Repository.Branch,
		PlanRequirements:  raw.Repository.PlanRequirements,
		ApplyRequirements: raw.Repository.ApplyRequirements,
		AllowedWorkflow:   raw.Repository.AllowedWorkflow,
		Automerge:         raw.Repository.Automerge,
		GitHub:            githubCfg,
	}
	if repo.Branch == "" {
		repo.Branch = "master"
	}
	if repo.PlanRequirements == nil {
		repo.PlanRequirements = []string{}
	}
	if repo.ApplyRequirements == nil {
		repo.ApplyRequirements = []string{}
	}

	workflows := &domain.Workflows{Workflows: make(map[string]domain.WorkflowStatement)}
	for wfName, phases := range raw.Workflows {
		st := domain.WorkflowStatement{Name: wfName, Phases: make(map[string]domain.WorkflowPhase)}
		for phaseName, rp := range phases {
			steps := make([]domain.WorkflowStep, 0, len(rp.Steps))
			for _, s := range rp.Steps {
				steps = append(steps, domain.WorkflowStep{
					Run:       s.Run,
					Terramate: s.Terramate,
					Changed:   s.Changed,
				})
			}
			st.Phases[phaseName] = domain.WorkflowPhase{
				Steps:     steps,
				DependsOn: rp.DependsOn,
			}
		}
		workflows.Workflows[wfName] = st
	}

	effectiveLevel := strings.TrimSpace(raw.LogLevel)
	if effectiveLevel == "" {
		effectiveLevel = "INFO"
	}
	if v := os.Getenv("NEPTUNE_LOG_LEVEL"); v != "" {
		effectiveLevel = strings.TrimSpace(v)
	}

	return &domain.NeptuneConfig{
		Repository: repo,
		Workflows:  workflows,
		LogLevel:   effectiveLevel,
	}, nil
}

// LoadError is a config load error (no GitHub comment).
type LoadError struct {
	Message string
}

func (e *LoadError) Error() string {
	return e.Message
}

func joinQuoted(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	b := ""
	for i, s := range ss {
		if i > 0 {
			b += ", "
		}
		b += "'" + s + "'"
	}
	return b
}
