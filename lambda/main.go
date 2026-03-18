package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/devopsfactory-io/neptune/lambda/pkg/config"
	"github.com/devopsfactory-io/neptune/lambda/pkg/github"
	"github.com/devopsfactory-io/neptune/lambda/pkg/verify"
	"github.com/devopsfactory-io/neptune/lambda/pkg/webhooks"
)

var (
	cfgOnce sync.Once
	appCfg  *config.Config
	cfgErr  error
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Only accept POST
	if req.RequestContext.HTTP.Method != "POST" {
		return response(405, "Method Not Allowed"), nil
	}

	body := req.Body
	if req.IsBase64Encoded {
		dec, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			log.Printf("base64 decode body: %v", err)
			return response(400, "Invalid body"), nil
		}
		body = string(dec)
	}

	sig := getHeader(req.Headers, "x-hub-signature-256")
	eventType := getHeader(req.Headers, "x-github-event")

	cfg, err := loadConfig(ctx)
	if err != nil {
		log.Printf("load config: %v", err)
		return response(500, "Configuration error"), nil
	}

	if err := verify.WebhookSignature([]byte(body), sig, cfg.WebhookSecret); err != nil {
		log.Printf("verify signature: %v", err)
		return response(401, "Invalid signature"), nil
	}

	appSlug := cfg.AppSlug
	if appSlug == "" {
		appSlug = "neptune"
	}

	switch eventType {
	case "pull_request":
		payload, instID, labels, addedLabel, err := webhooks.ParsePullRequest([]byte(body))
		if err != nil {
			log.Printf("parse pull_request: %v", err)
			return response(400, "Bad payload"), nil
		}
		if payload == nil {
			return response(200, "OK"), nil // unsupported action
		}
		if !shouldDispatchPullRequest(cfg.PrLabel, payload.PullRequestAction, addedLabel, labels) {
			return response(200, "OK"), nil
		}
		token, err := github.InstallationToken(ctx, cfg.AppID, cfg.PrivateKey, instID)
		if err != nil {
			log.Printf("installation token: %v", err)
			return response(500, "GitHub auth error"), nil
		}
		client := github.NewClient(token)
		if err := client.RepositoryDispatch(ctx, payload.PullRequestRepoFull, payload); err != nil {
			log.Printf("repository_dispatch: %v", err)
			return response(500, "Dispatch error"), nil
		}
		if err := client.CreateReactionForIssue(ctx, payload.PullRequestRepoFull, payload.PullRequestNumber, "eyes"); err != nil {
			log.Printf("eyes reaction on PR #%d failed (dispatch succeeded): %v", payload.PullRequestNumber, err)
		}
		return response(200, "OK"), nil

	case "issue_comment":
		payload, instID, commentID, labels, ok, err := webhooks.ParseIssueComment([]byte(body), appSlug)
		if err != nil {
			log.Printf("parse issue_comment: %v", err)
			return response(400, "Bad payload"), nil
		}
		if !ok || payload == nil {
			return response(200, "OK"), nil
		}
		if cfg.PrLabel != "" && !hasLabel(labels, cfg.PrLabel) {
			return response(200, "OK"), nil
		}
		token, err := github.InstallationToken(ctx, cfg.AppID, cfg.PrivateKey, instID)
		if err != nil {
			log.Printf("installation token: %v", err)
			return response(500, "GitHub auth error"), nil
		}
		client := github.NewClient(token)
		// Fetch PR to get head ref and sha
		branch, sha, err := client.GetPR(ctx, payload.PullRequestRepoFull, payload.PullRequestNumber)
		if err != nil {
			log.Printf("get PR: %v", err)
			return response(500, "GitHub API error"), nil
		}
		payload.PullRequestBranch = branch
		payload.PullRequestSHA = sha
		if err := client.RepositoryDispatch(ctx, payload.PullRequestRepoFull, payload); err != nil {
			log.Printf("repository_dispatch: %v", err)
			return response(500, "Dispatch error"), nil
		}
		if commentID != 0 {
			if err := client.CreateReactionForIssueComment(ctx, payload.PullRequestRepoFull, commentID, "eyes"); err != nil {
				log.Printf("eyes reaction on comment %d failed (dispatch succeeded): %v", commentID, err)
			}
		}
		return response(200, "OK"), nil

	default:
		return response(200, "OK"), nil
	}
}

func loadConfig(ctx context.Context) (*config.Config, error) {
	cfgOnce.Do(func() {
		appCfg, cfgErr = config.Load(ctx)
	})
	return appCfg, cfgErr
}

func response(status int, body string) events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		},
		Body: body,
	}
}

func getHeader(h map[string]string, key string) string {
	if h == nil {
		return ""
	}
	for k, v := range h {
		if len(k) == len(key) && strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

// shouldDispatchPullRequest reports whether a pull_request webhook event should
// trigger a repository_dispatch. prLabel is the configured label gate (empty
// means no gate). action is the webhook action (e.g. "opened", "labeled").
// labels are the PR's current labels. addedLabel is non-empty only for
// "labeled" events and holds the name of the label that was just added.
//
// Deduplication rule: when prLabel is set and the action is "opened",
// GitHub fires both a pull_request.opened event (with the label already in
// pull_request.labels) and a pull_request.labeled event. To avoid two
// dispatches for the same PR creation we suppress "opened" here and rely on
// the labeled event as the sole initial trigger when a label gate is active.
// "reopened" is NOT suppressed because GitHub does not fire a labeled event
// on reopen. synchronize and ready_for_review are also gated on label
// presence because no labeled event accompanies those actions.
func shouldDispatchPullRequest(prLabel, action, addedLabel string, labels []string) bool {
	if addedLabel != "" {
		// labeled event: only dispatch when the added label matches the gate.
		return prLabel != "" && addedLabel == prLabel
	}
	if prLabel == "" {
		return true
	}
	// prLabel is set; non-labeled action.
	switch action {
	case "opened":
		// Suppress: the accompanying labeled event will trigger dispatch.
		return false
	default:
		return hasLabel(labels, prLabel)
	}
}

// hasLabel returns true if want is in labels (case-sensitive).
func hasLabel(labels []string, want string) bool {
	for _, name := range labels {
		if name == want {
			return true
		}
	}
	return false
}

func init() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.SetOutput(os.Stdout)
	}
}
