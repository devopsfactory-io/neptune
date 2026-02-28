#!/usr/bin/env bash
# Check eyes reactions on a PR and its comments (for Neptune webhook Lambda).
# Usage: ./check-reactions.sh [PR_NUMBER]
#        REPO=owner/repo ./check-reactions.sh [PR_NUMBER]
# Requires: gh CLI (gh auth login), jq
set -e

REPO="${REPO:-devopsfactory-io/neptune}"
PR_NUM="${1:-}"

if ! command -v gh &>/dev/null; then
  echo "Error: gh CLI not found. Install from https://cli.github.com/"
  exit 1
fi

if ! command -v jq &>/dev/null; then
  echo "Error: jq not found. Install jq to run this script."
  exit 1
fi

if [[ -z "$PR_NUM" ]]; then
  echo "Usage: $0 PR_NUMBER"
  echo "       REPO=owner/repo $0 PR_NUMBER"
  echo ""
  echo "Checks whether the PR and its comments have eyes (👀) reactions,"
  echo "which the Neptune Lambda adds when it handles pull_request and"
  echo "issue_comment webhooks."
  exit 0
fi

echo "Repo: $REPO"
echo "PR: #$PR_NUM"
echo ""

# Reactions on the PR (issue)
echo "--- Reactions on PR #$PR_NUM ---"
PR_REACTIONS=$(gh api "repos/$REPO/issues/$PR_NUM/reactions" 2>/dev/null || echo "[]")
EYES_PR=$(echo "$PR_REACTIONS" | jq '[.[] | select(.content == "eyes")] | length')
TOTAL_PR=$(echo "$PR_REACTIONS" | jq 'length')
if [[ "$EYES_PR" -gt 0 ]]; then
  echo "  Eyes reaction: yes ($EYES_PR)"
else
  echo "  Eyes reaction: no (total reactions: $TOTAL_PR)"
fi
echo ""

# Comments and their reactions
echo "--- Reactions on PR comments ---"
COMMENTS_JSON=$(gh api "repos/$REPO/issues/$PR_NUM/comments" 2>/dev/null || echo "[]")
COUNT=$(echo "$COMMENTS_JSON" | jq 'length')
if [[ "$COUNT" -eq 0 ]]; then
  echo "  No comments."
else
  for i in $(seq 0 $((COUNT - 1))); do
    CID=$(echo "$COMMENTS_JSON" | jq -r ".[$i].id")
    BODY=$(echo "$COMMENTS_JSON" | jq -r ".[$i].body" | head -c 50)
    USER=$(echo "$COMMENTS_JSON" | jq -r ".[$i].user.login")
    REACTIONS=$(gh api "repos/$REPO/issues/comments/$CID/reactions" 2>/dev/null || echo "[]")
    EYES=$(echo "$REACTIONS" | jq '[.[] | select(.content == "eyes")] | length')
    if [[ "$EYES" -gt 0 ]]; then
      echo "  Comment $CID ($USER): eyes=yes"
    else
      echo "  Comment $CID ($USER): eyes=no  \"${BODY}...\""
    fi
  done
fi
echo ""

echo "--- Note ---"
echo "If eyes reactions are missing, the Lambda may be failing to add them."
echo "Check: 1) Lambda CloudWatch logs for 'eyes reaction on PR' or 'eyes reaction on comment'"
echo "       2) GitHub App has Issues: Read and write (Permissions and events)"
echo "       3) Installation has accepted the new permissions"
