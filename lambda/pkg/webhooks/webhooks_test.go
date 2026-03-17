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
	payload, instID, labels, addedLabel, err := ParsePullRequest([]byte(pullRequestOpened))
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
	if addedLabel != "" {
		t.Errorf("addedLabel: expected empty for opened, got %q", addedLabel)
	}
}

func TestParsePullRequest_ValidSynchronize(t *testing.T) {
	payload, instID, labels, addedLabel, err := ParsePullRequest([]byte(pullRequestSynchronize))
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
	if addedLabel != "" {
		t.Errorf("addedLabel: expected empty for synchronize, got %q", addedLabel)
	}
}

func TestParsePullRequest_UnsupportedAction(t *testing.T) {
	payload, instID, labels, addedLabel, err := ParsePullRequest([]byte(pullRequestClosed))
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
	if addedLabel != "" {
		t.Errorf("expected empty addedLabel when skipping, got %q", addedLabel)
	}
}

func TestParsePullRequest_InvalidJSON(t *testing.T) {
	_, _, _, _, err := ParsePullRequest([]byte(`{`))
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
	payload, _, labels, addedLabel, err := ParsePullRequest([]byte(pullRequestWithNeptuneLabel))
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
	if addedLabel != "" {
		t.Errorf("addedLabel: expected empty for opened, got %q", addedLabel)
	}

	_, _, labelsOther, _, err := ParsePullRequest([]byte(pullRequestWithOtherLabel))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labelsOther) != 1 || labelsOther[0] != "other" {
		t.Errorf("labels: expected [other], got %v", labelsOther)
	}
}

const pullRequestLabeledNeptune = `{
  "action": "labeled",
  "number": 11,
  "repository": {"full_name": "devopsfactory-io/neptune"},
  "installation": {"id": 123},
  "pull_request": {
    "head": {"ref": "feat/branch", "sha": "2a2b3f35975b4a761678984cf33e3f0289f41856"},
    "labels": [{"name": "neptune"}]
  },
  "label": {"name": "neptune"}
}`

const pullRequestLabeledOther = `{
  "action": "labeled",
  "number": 12,
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 456},
  "pull_request": {
    "head": {"ref": "other-branch", "sha": "abcde"},
    "labels": [{"name": "other"}]
  },
  "label": {"name": "other"}
}`

func TestParsePullRequest_Labeled(t *testing.T) {
	payload, instID, labels, addedLabel, err := ParsePullRequest([]byte(pullRequestLabeledNeptune))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected non-nil payload")
	}
	if payload.Command != string(CommandPlan) {
		t.Errorf("command: got %q", payload.Command)
	}
	if payload.PullRequestNumber != 11 {
		t.Errorf("number: got %d", payload.PullRequestNumber)
	}
	if payload.PullRequestBranch != "feat/branch" {
		t.Errorf("branch: got %q", payload.PullRequestBranch)
	}
	if payload.PullRequestSHA != "2a2b3f35975b4a761678984cf33e3f0289f41856" {
		t.Errorf("sha: got %q", payload.PullRequestSHA)
	}
	if payload.PullRequestRepoFull != "devopsfactory-io/neptune" {
		t.Errorf("repo: got %q", payload.PullRequestRepoFull)
	}
	if instID != 123 {
		t.Errorf("installation ID: got %d", instID)
	}
	if len(labels) != 1 || labels[0] != "neptune" {
		t.Errorf("labels: expected [neptune], got %v", labels)
	}
	if addedLabel != "neptune" {
		t.Errorf("addedLabel: expected neptune, got %q", addedLabel)
	}
}

func TestParsePullRequest_LabeledOtherLabel(t *testing.T) {
	payload, instID, labels, addedLabel, err := ParsePullRequest([]byte(pullRequestLabeledOther))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected non-nil payload for labeled (handler filters by PrLabel)")
	}
	if payload.PullRequestNumber != 12 || payload.PullRequestBranch != "other-branch" {
		t.Errorf("unexpected payload: %+v", payload)
	}
	if instID != 456 {
		t.Errorf("installation ID: got %d", instID)
	}
	if len(labels) != 1 || labels[0] != "other" {
		t.Errorf("labels: expected [other], got %v", labels)
	}
	if addedLabel != "other" {
		t.Errorf("addedLabel: expected other, got %q", addedLabel)
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
  "comment": {"id": 3003, "body": "To apply these changes, comment:\n@neptbot apply", "user": {"type": "Bot", "login": "neptbot[bot]"}}
}`

const issueCommentExternalBot = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 4004, "body": "@neptbot apply", "user": {"type": "Bot", "login": "neptune-ci[bot]"}}
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

// TestParseIssueComment_BotAuthorDoesNotTrigger tests the self-bot login filter:
// the comment author is neptbot[bot], which matches the selfBotLogin guard and
// returns early before reaching the instructional text check.
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

func TestParseIssueComment_ExternalBotCanTrigger(t *testing.T) {
	payload, instID, commentID, _, ok, err := ParseIssueComment([]byte(issueCommentExternalBot), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true for external bot comment")
	}
	if payload == nil {
		t.Fatal("expected non-nil payload for external bot comment")
	}
	if payload.Command != string(CommandApply) {
		t.Errorf("command: got %q, want %q", payload.Command, CommandApply)
	}
	if instID != 111 {
		t.Errorf("installation ID: got %d, want 111", instID)
	}
	if commentID != 4004 {
		t.Errorf("comment ID: got %d, want 4004", commentID)
	}
}

// issueCommentPlanTextGitHubActions simulates the plan comment posted by github-actions[bot]
// when GITHUB_TOKEN is used instead of the app's installation token. The comment body
// contains "To apply these changes, comment: @neptbot apply" instructional text and must
// not be interpreted as a command by the webhook handler.
const issueCommentPlanTextGitHubActions = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 5005, "body": "To apply these changes, comment:\n` + "`" + `\n@neptbot apply\n` + "`" + `", "user": {"type": "Bot", "login": "github-actions[bot]"}}
}`

func TestParseIssueComment_PlanCommentInstructionalTextDoesNotTrigger(t *testing.T) {
	payload, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentPlanTextGitHubActions), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false: plan comment with instructional text should not trigger a command")
	}
	if payload != nil {
		t.Errorf("expected nil payload for plan comment with instructional text, got %+v", payload)
	}
}

// issueCommentExternalBotApply simulates an external CI bot (not the Neptune app bot)
// posting a plain "@neptbot apply" command — this should still trigger apply.
const issueCommentExternalBotApply = `{
  "action": "created",
  "issue": {"number": 11, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 6006, "body": "@neptbot apply", "user": {"type": "Bot", "login": "neptune-ci[bot]"}}
}`

func TestParseIssueComment_ExternalBotApplyTriggers(t *testing.T) {
	payload, _, commentID, _, ok, err := ParseIssueComment([]byte(issueCommentExternalBotApply), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok true: external bot plain apply comment should trigger a command")
	}
	if payload == nil {
		t.Fatal("expected non-nil payload for external bot apply comment")
	}
	if payload.Command != string(CommandApply) {
		t.Errorf("command: got %q, want %q", payload.Command, CommandApply)
	}
	if commentID != 6006 {
		t.Errorf("comment ID: got %d, want 6006", commentID)
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

// issueCommentExternalBotInstructionalText simulates an external bot (not the Neptune
// app bot) posting a comment that contains the instructional text phrase used in plan
// results. Despite being a different bot login, the instructional text guard applies to
// all Bot-type users and must prevent triggering.
const issueCommentExternalBotInstructionalText = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 7007, "body": "To apply these changes, comment:\n@neptbot apply", "user": {"type": "Bot", "login": "some-other-ci[bot]"}}
}`

// issueCommentBotInstructionalTextNoApply simulates a bot comment containing "To apply
// these changes" but without any recognisable command. The early-exit on the
// instructional text check means the mention guard is never reached for apply/plan
// matching; the result must still be no-trigger.
const issueCommentBotInstructionalTextNoApply = `{
  "action": "created",
  "issue": {"number": 10, "pull_request": {}},
  "repository": {"full_name": "owner/repo"},
  "installation": {"id": 111},
  "comment": {"id": 8008, "body": "To apply these changes, see the run log.", "user": {"type": "Bot", "login": "some-other-ci[bot]"}}
}`

func TestParseIssueComment_ExternalBotInstructionalTextDoesNotTrigger(t *testing.T) {
	payload, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentExternalBotInstructionalText), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false: external bot posting instructional text should not trigger a command")
	}
	if payload != nil {
		t.Errorf("expected nil payload for external bot instructional text comment, got %+v", payload)
	}
}

func TestParseIssueComment_BotInstructionalTextWithoutApplyDoesNotTrigger(t *testing.T) {
	payload, _, _, _, ok, err := ParseIssueComment([]byte(issueCommentBotInstructionalTextNoApply), "neptbot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok false: bot instructional text comment without apply command should not trigger")
	}
	if payload != nil {
		t.Errorf("expected nil payload, got %+v", payload)
	}
}

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
