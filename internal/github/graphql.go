package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devopsfactory-io/neptune/internal/log"
)

const graphQLURL = "https://api.github.com/graphql"

// graphQLResponse is the generic shape of a GraphQL response (data or errors).
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

// EnablePullRequestAutoMerge enables auto-merge on the current PR via GitHub's GraphQL API.
// The PR will merge when all required checks pass. Uses merge method MERGE.
// On failure (e.g. repo does not allow auto-merge, or token lacks permission), returns an error.
func (c *Client) EnablePullRequestAutoMerge(ctx context.Context) error {
	url := graphQLURL
	if c.graphQLURL != "" {
		url = c.graphQLURL
	}
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, name := parts[0], parts[1]
	prNum := parseInt(c.prNum)
	if prNum <= 0 {
		return fmt.Errorf("invalid PR number: %s", c.prNum)
	}

	// Query pull request node ID
	queryPR := `query($owner: String!, $name: String!, $number: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      id
    }
  }
}`
	varsPR := map[string]interface{}{
		"owner":  owner,
		"name":   name,
		"number": prNum,
	}
	body, err := json.Marshal(map[string]interface{}{
		"query":     queryPR,
		"variables": varsPR,
	})
	if err != nil {
		return fmt.Errorf("build GraphQL request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GraphQL request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.For("github").Error("close response body", "err", err)
		}
	}()
	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return fmt.Errorf("decode GraphQL response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %s", gqlResp.Errors[0].Message)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GraphQL status %d", resp.StatusCode)
	}
	// Parse data.repository.pullRequest.id
	var prData struct {
		Repository struct {
			PullRequest struct {
				ID string `json:"id"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(gqlResp.Data, &prData); err != nil {
		return fmt.Errorf("parse PR id: %w", err)
	}
	prID := prData.Repository.PullRequest.ID
	if prID == "" {
		return fmt.Errorf("pull request not found or has no id")
	}

	// Enable auto-merge
	mutation := `mutation($pullRequestId: ID!) {
  enablePullRequestAutoMerge(input: { pullRequestId: $pullRequestId, mergeMethod: MERGE }) {
    clientMutationId
  }
}`
	varsMerge := map[string]interface{}{"pullRequestId": prID}
	body, err = json.Marshal(map[string]interface{}{
		"query":     mutation,
		"variables": varsMerge,
	})
	if err != nil {
		return fmt.Errorf("build enable automerge request: %w", err)
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create automerge request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("enable automerge request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.For("github").Error("close response body", "err", err)
		}
	}()
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return fmt.Errorf("decode automerge response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL automerge errors: %s", gqlResp.Errors[0].Message)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("enable automerge status %d", resp.StatusCode)
	}
	return nil
}
