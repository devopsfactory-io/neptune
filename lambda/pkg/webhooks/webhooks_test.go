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
	payload, instID, labels, err := ParsePullRequest([]byte(pullRequestOpened))
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
	if len(labels) != 0 {
		t.Errorf("labels: expected empty, got %v", labels)
	}
}

func TestParsePullRequest_ValidSynchronize(t *testing.T) {
	payload, instID, labels, err := ParsePullRequest([]byte(pullRequestSynchronize))
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
	if len(labels) != 0 {
		t.Errorf("labels: expected empty, got %v", labels)
	}
}

func TestParsePullRequest_UnsupportedAction(t *testing.T) {
	payload, instID, labels, err := ParsePullRequest([]byte(pullRequestClosed))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload != nil {
		t.Errorf("expected nil payload for closed, got %+v", payload)
	}
	if instID != 0 {
		t.Errorf("expected zero instID when skipping, got %d", instID)
	}
	if labels != nil {
		t.Errorf("expected nil labels when skipping, got %v", labels)
	}
}

func TestParsePullRequest_InvalidJSON(t *testing.T) {
	_, _, _, err := ParsePullRequest([]byte(`{`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

const pullRequestWithNeptuneLabel = `{
  "action": "opened",
  "number": 1,
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 1},
  "pull_request": {
    "head": {"ref": "branch", "sha": "sha1"},
    "labels": [{"name": "neptune"}, {"name": "infra"}]
  }
}`

const pullRequestWithOtherLabel = `{
  "action": "opened",
  "number": 2,
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 1},
  "pull_request": {
    "head": {"ref": "branch", "sha": "sha1"},
    "labels": [{"name": "other"}]
  }
}`

func TestParsePullRequest_Labels(t *testing.T) {
	payload, _, labels, err := ParsePullRequest([]byte(pullRequestWithNeptuneLabel))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected non-nil payload")
	}
	if len(labels) != 2 {
		t.Fatalf("labels: expected 2, got %v", labels)
	}
	hasNeptune := false
	for _, name := range labels {
		if name == "neptune" {
			hasNeptune = true
			break
		}
	}
	if !hasNeptune {
		t.Errorf("labels: expected to contain neptune, got %v", labels)
	}

	_, _, labelsOther, err := ParsePullRequest([]byte(pullRequestWithOtherLabel))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labelsOther) != 1 || labelsOther[0] != "other" {
		t.Errorf("labels: expected [other], got %v", labelsOther)
	}
}

const issueCommentApply = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 1001, "body": "@neptbot apply"}
}`

const issueCommentPlan = `{
  "action": "created",
  "issue": {"number": 5, "pull_request": {}},
  "repository": {"full_name": "a/b"},
  "installation": {"id": 222},
  "comment": {"id": 2002, "body": "Please @neptbot plan"}
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

const issueCommentBotAuthor = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 3003, "body": "To apply these changes, comment:\n@neptbot apply", "user": {"type": "Bot", "login": "github-actions[bot]"}}
}`

func TestParseIssueComment_ValidApply(t *testing.T) {
	payload, instID, commentID, labels, ok, err := ParseIssueComment([]byte(issueCommentApply), "neptbot")
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
	if commentID != 1001 {
		t.Errorf("comment ID: got %d", commentID)
	}
	if len(labels) != 0 {
		t.Errorf("labels: expected empty, got %v", labels)
	}
}

func TestParseIssueComment_ValidPlan(t *testing.T) {
	payload, _, commentID, _, ok, err := ParseIssueComment([]byte(issueCommentPlan), "neptbot")
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
	if commentID != 2002 {
		t.Errorf("comment ID: got %d", commentID)
	}
}

func TestParseIssueComment_NotPR(t *testing.T) {
	payload, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentNotPR), "neptbot")
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
	_, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentNoMention), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false when body has no @-mention")
	}
}

func TestParseIssueComment_MentionNoCommand(t *testing.T) {
	_, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentMentionNoCommand), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false when body has mention but no apply/plan")
	}
}

func TestParseIssueComment_BotAuthorDoesNotTrigger(t *testing.T) {
	payload, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentBotAuthor), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false when comment author is a bot")
	}
	if payload != nil {
		t.Errorf("expected nil payload for bot comment, got %+v", payload)
	}
}

func TestParseIssueComment_InvalidJSON(t *testing.T) {
	_, _, _, _, _, err := ParseIssueComment([]byte(`not json`), "neptbot")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseIssueComment_DefaultMention(t *testing.T) {
	body := `{"action":"created","issue":{"number":2,"pull_request":{}},"repository":{"full_name":"o/r"},"installation":{"id":1},"comment":{"id":99,"body":"@neptbot plan"}}`
	payload, _, commentID, _, ok, err := ParseIssueComment([]byte(body), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true with default mention")
	}
	if payload == nil || payload.Command != string(CommandPlan) {
		t.Errorf("expected plan command with default mention, got %+v", payload)
	}
	if commentID != 99 {
		t.Errorf("comment ID: got %d", commentID)
	}
}

const issueCommentWithNeptuneLabel = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}, "labels": [{"name": "neptune"}]},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 1001, "body": "@neptbot apply"}
}`

const issueCommentWithOtherLabel = `{
  "action": "created",
  "issue": {"number": 11, "pull_request": {}, "labels": [{"name": "other"}]},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 1002, "body": "@neptbot plan"}
}`

func TestParseIssueComment_Labels(t *testing.T) {
	payload, _, _, labels, ok, err := ParseIssueComment([]byte(issueCommentWithNeptuneLabel), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true")
	}
	if payload == nil {
		t.Fatal("expected non-nil payload")
	}
	if len(labels) != 1 || labels[0] != "neptune" {
		t.Errorf("labels: expected [neptune], got %v", labels)
	}

	_, _, _, labelsOther, ok, err := ParseIssueComment([]byte(issueCommentWithOtherLabel), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true")
	}
	if len(labelsOther) != 1 || labelsOther[0] != "other" {
		t.Errorf("labels: expected [other], got %v", labelsOther)
	}
}
