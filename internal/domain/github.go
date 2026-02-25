package domain

// PRRequirementsStatus is the result of checking PR requirements.
type PRRequirementsStatus struct {
	IsCompliant        bool
	FailedRequirements []string
	ErrorMessage       string
}

// PRInfo holds PR data from the GitHub API.
type PRInfo struct {
	Response map[string]interface{}
	PRNumber string
	Repo     string
	APIURL   string
}

// PullRequestComment is the payload for a PR comment.
type PullRequestComment struct {
	StepsOutput   *StepsOutput
	Stacks        *TerraformStacks
	SimpleOutput  string
	OverallStatus int
}
