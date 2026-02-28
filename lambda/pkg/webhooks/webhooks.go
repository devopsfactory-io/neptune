package webhooks

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Command is the Neptune command to run (plan or apply).
type Command string

const (
	CommandPlan  Command = "plan"
	CommandApply Command = "apply"
)

// DispatchPayload is the data we send in repository_dispatch client_payload.
type DispatchPayload struct {
	Command             string `json:"command"`
	PullRequestNumber   int    `json:"pull_request_number"`
	PullRequestBranch   string `json:"pull_request_branch"`
	PullRequestSHA      string `json:"pull_request_sha,omitempty"`
	PullRequestRepoFull string `json:"pull_request_repo_full,omitempty"`
}

// PullRequestPayload is the relevant part of GitHub pull_request webhook.
type PullRequestPayload struct {
	Action       string        `json:"action"`
	Number       int           `json:"number"`
	Repository   Repo          `json:"repository"`
	Installation *Installation `json:"installation"`
	PullRequest  struct {
		Head struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Labels []struct{ Name string } `json:"labels"`
	} `json:"pull_request"`
	// Label is set for "labeled" / "unlabeled" actions (the label that was added or removed).
	Label *struct{ Name string } `json:"label,omitempty"`
}

// IssueCommentPayload is the relevant part of GitHub issue_comment webhook.
type IssueCommentPayload struct {
	Action string `json:"action"`
	Issue  struct {
		Number      int                     `json:"number"`
		PullRequest *struct{}               `json:"pull_request,omitempty"`
		Labels      []struct{ Name string } `json:"labels"`
	} `json:"issue"`
	Repository   Repo          `json:"repository"`
	Installation *Installation `json:"installation"`
	Comment      struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
		User struct {
			Type string `json:"type"`
		} `json:"user"`
	} `json:"comment"`
}

// Repo is repository info from webhook.
type Repo struct {
	FullName string `json:"full_name"`
}

// Installation is GitHub App installation from webhook.
type Installation struct {
	ID int64 `json:"id"`
}

// ParsePullRequest parses the pull_request webhook body and returns dispatch payload for "plan" if action is supported, label names from pull_request.labels, and for "labeled" the name of the added label (addedLabel); otherwise addedLabel is "".
func ParsePullRequest(body []byte) (*DispatchPayload, int64, []string, string, error) {
	var p PullRequestPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, 0, nil, "", err
	}
	allowed := map[string]bool{
		"opened": true, "reopened": true, "synchronize": true, "ready_for_review": true, "labeled": true,
	}
	if !allowed[p.Action] {
		return nil, 0, nil, "", nil
	}
	var instID int64
	if p.Installation != nil {
		instID = p.Installation.ID
	}
	labels := make([]string, 0, len(p.PullRequest.Labels))
	for _, l := range p.PullRequest.Labels {
		if l.Name != "" {
			labels = append(labels, l.Name)
		}
	}
	var addedLabel string
	if p.Action == "labeled" && p.Label != nil {
		addedLabel = p.Label.Name
	}
	return &DispatchPayload{
		Command:             string(CommandPlan),
		PullRequestNumber:   p.Number,
		PullRequestBranch:   p.PullRequest.Head.Ref,
		PullRequestSHA:      p.PullRequest.Head.SHA,
		PullRequestRepoFull: p.Repository.FullName,
	}, instID, labels, addedLabel, nil
}

// ParseIssueComment parses the issue_comment webhook body. If the comment is on a PR and mentions the app with a command, returns (dispatch payload, installation ID, comment ID, label names from issue.labels, true). appMention is the app login/slug (e.g. "neptbot").
func ParseIssueComment(body []byte, appMention string) (*DispatchPayload, int64, int64, []string, bool, error) {
	var p IssueCommentPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, 0, 0, nil, false, err
	}
	if p.Action != "created" {
		return nil, 0, 0, nil, false, nil
	}
	if p.Issue.PullRequest == nil {
		return nil, 0, 0, nil, false, nil
	}
	bodyLower := strings.ToLower(strings.TrimSpace(p.Comment.Body))
	mentionLower := strings.ToLower(strings.TrimSpace(appMention))
	if mentionLower == "" {
		mentionLower = "neptbot"
	}
	if !strings.Contains(bodyLower, "@"+mentionLower) {
		return nil, 0, 0, nil, false, nil
	}
	var cmd Command
	if matchApply.MatchString(bodyLower) {
		cmd = CommandApply
	} else if matchPlan.MatchString(bodyLower) {
		cmd = CommandPlan
	} else {
		return nil, 0, 0, nil, false, nil
	}
	if p.Comment.User.Type == "Bot" {
		return nil, 0, 0, nil, false, nil
	}
	var instID int64
	if p.Installation != nil {
		instID = p.Installation.ID
	}
	commentID := p.Comment.ID
	labels := make([]string, 0, len(p.Issue.Labels))
	for _, l := range p.Issue.Labels {
		if l.Name != "" {
			labels = append(labels, l.Name)
		}
	}
	return &DispatchPayload{
		Command:             string(cmd),
		PullRequestNumber:   p.Issue.Number,
		PullRequestRepoFull: p.Repository.FullName,
	}, instID, commentID, labels, true, nil
}

var (
	matchApply = regexp.MustCompile(`\bapply\b`)
	matchPlan  = regexp.MustCompile(`\bplan\b`)
)
