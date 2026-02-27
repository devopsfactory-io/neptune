package webhooks

import (
	"testing"
)

const pullRequestOpened = `{
  "action": "opened",
  "number": 42,
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 12345},
  "pull_request": {
    "head": {"ref": "feature-branch", "sha": "abc123"}
  }
}`

const pullRequestSynchronize = `{
  "action": "synchronize",
  "number": 7,
  "repository": {"full_name": "org/proj"},
  "installation": {"id": 999},
  "pull_request": {
    "head": {"ref": "main", "sha": "sha789"}
  }
}`

const pullRequestClosed = `{
  "action": "closed",
  "number": 1,
  "repository": {"full_name": "x/y"},
  "installation": {"id": 1},
  "pull_request": {"head": {"ref": "br", "sha": "s"}}
}`

func TestParsePullRequest_ValidOpened(t *testing.T) {
	payload, instID, err := ParsePullRequest([]byte(pullRequestOpened))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected non-nil payload")
	}
	if payload.Command != string(CommandPlan) {
		t.Errorf("command: got %q", payload.Command)
	}
	if payload.PullRequestNumber != 42 {
		t.Errorf("number: got %d", payload.PullRequestNumber)
	}
	if payload.PullRequestBranch != "feature-branch" {
		t.Errorf("branch: got %q", payload.PullRequestBranch)
	}
	if payload.PullRequestSHA != "abc123" {
		t.Errorf("sha: got %q", payload.PullRequestSHA)
	}
	if payload.PullRequestRepoFull != "owner/repo" {
		t.Errorf("repo: got %q", payload.PullRequestRepoFull)
	}
	if instID != 12345 {
		t.Errorf("installation ID: got %d", instID)
	}
}

func TestParsePullRequest_ValidSynchronize(t *testing.T) {
	payload, instID, err := ParsePullRequest([]byte(pullRequestSynchronize))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected non-nil payload")
	}
	if payload.Command != string(CommandPlan) {
		t.Errorf("command: got %q", payload.Command)
	}
	if payload.PullRequestNumber != 7 || payload.PullRequestBranch != "main" || payload.PullRequestSHA != "sha789" {
		t.Errorf("unexpected payload: %+v", payload)
	}
	if instID != 999 {
		t.Errorf("installation ID: got %d", instID)
	}
}

func TestParsePullRequest_UnsupportedAction(t *testing.T) {
	payload, instID, err := ParsePullRequest([]byte(pullRequestClosed))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload != nil {
		t.Errorf("expected nil payload for closed, got %+v", payload)
	}
	if instID != 0 {
		t.Errorf("expected zero instID when skipping, got %d", instID)
	}
}

func TestParsePullRequest_InvalidJSON(t *testing.T) {
	_, _, err := ParsePullRequest([]byte(`{`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

const issueCommentApply = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"body": "@neptbot apply"}
}`

const issueCommentPlan = `{
  "action": "created",
  "issue": {"number": 5, "pull_request": {}},
  "repository": {"full_name": "a/b"},
  "installation": {"id": 222},
  "comment": {"body": "Please @neptbot plan"}
}`

const issueCommentNotPR = `{
  "action": "created",
  "issue": {"number": 3},
  "repository": {"full_name": "x/y"},
  "comment": {"body": "@neptbot apply"}
}`

const issueCommentNoMention = `{
  "action": "created",
  "issue": {"number": 1, "pull_request": {}},
  "repository": {"full_name": "x/y"},
  "comment": {"body": "just a comment"}
}`

const issueCommentMentionNoCommand = `{
  "action": "created",
  "issue": {"number": 1, "pull_request": {}},
  "repository": {"full_name": "x/y"},
  "comment": {"body": "@neptbot hello"}
}`

func TestParseIssueComment_ValidApply(t *testing.T) {
	payload, instID, ok, err := ParseIssueComment([]byte(issueCommentApply), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true")
	}
	if payload == nil {
		t.Fatal("expected non-nil payload")
	}
	if payload.Command != string(CommandApply) {
		t.Errorf("command: got %q", payload.Command)
	}
	if payload.PullRequestNumber != 10 {
		t.Errorf("number: got %d", payload.PullRequestNumber)
	}
	if payload.PullRequestRepoFull != "owner/repo" {
		t.Errorf("repo: got %q", payload.PullRequestRepoFull)
	}
	if instID != 111 {
		t.Errorf("installation ID: got %d", instID)
	}
}

func TestParseIssueComment_ValidPlan(t *testing.T) {
	payload, _, ok, err := ParseIssueComment([]byte(issueCommentPlan), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true")
	}
	if payload.Command != string(CommandPlan) {
		t.Errorf("command: got %q", payload.Command)
	}
	if payload.PullRequestNumber != 5 {
		t.Errorf("number: got %d", payload.PullRequestNumber)
	}
}

func TestParseIssueComment_NotPR(t *testing.T) {
	payload, _, ok, err := ParseIssueComment([]byte(issueCommentNotPR), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false when issue is not a PR")
	}
	if payload != nil {
		t.Errorf("expected nil payload, got %+v", payload)
	}
}

func TestParseIssueComment_NoMention(t *testing.T) {
	_, _, ok, err := ParseIssueComment([]byte(issueCommentNoMention), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false when body has no @-mention")
	}
}

func TestParseIssueComment_MentionNoCommand(t *testing.T) {
	_, _, ok, err := ParseIssueComment([]byte(issueCommentMentionNoCommand), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false when body has mention but no apply/plan")
	}
}

func TestParseIssueComment_InvalidJSON(t *testing.T) {
	_, _, _, err := ParseIssueComment([]byte(`not json`), "neptbot")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseIssueComment_DefaultMention(t *testing.T) {
	body := `{"action":"created","issue":{"number":2,"pull_request":{}},"repository":{"full_name":"o/r"},"installation":{"id":1},"comment":{"body":"@neptbot plan"}}`
	payload, _, ok, err := ParseIssueComment([]byte(body), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true with default mention")
	}
	if payload == nil || payload.Command != string(CommandPlan) {
		t.Errorf("expected plan command with default mention, got %+v", payload)
	}
}
