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

// graphQLRequest sends a GraphQL request and returns the decoded response.
func (c *Client) graphQLRequest(ctx context.Context, query string, variables map[string]interface{}) (*graphQLResponse, error) {
	url := graphQLURL
	if c.graphQLURL != "" {
		url = c.graphQLURL
	}
	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("build GraphQL request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Do(req) //nolint:gosec // G704: URL from internal config constant
	if err != nil {
		return nil, fmt.Errorf("GraphQL request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.For("github").Error("close response body", "err", err)
		}
	}()
	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("decode GraphQL response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %s", gqlResp.Errors[0].Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL status %d", resp.StatusCode)
	}
	return &gqlResp, nil
}

// queryPRNodeID fetches the GraphQL node ID for a pull request.
func (c *Client) queryPRNodeID(ctx context.Context, owner, name string, number int) (string, error) {
	query := `query($owner: String!, $name: String!, $number: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      id
    }
  }
}`
	vars := map[string]interface{}{
		"owner":  owner,
		"name":   name,
		"number": number,
	}
	gqlResp, err := c.graphQLRequest(ctx, query, vars)
	if err != nil {
		return "", err
	}
	var prData struct {
		Repository struct {
			PullRequest struct {
				ID string `json:"id"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(gqlResp.Data, &prData); err != nil {
		return "", fmt.Errorf("parse PR id: %w", err)
	}
	if prData.Repository.PullRequest.ID == "" {
		return "", fmt.Errorf("pull request not found or has no id")
	}
	return prData.Repository.PullRequest.ID, nil
}

// EnablePullRequestAutoMerge enables auto-merge on the current PR via GitHub's GraphQL API.
// The PR will merge when all required checks pass. Uses merge method MERGE.
// On failure (e.g. repo does not allow auto-merge, or token lacks permission), returns an error.
func (c *Client) EnablePullRequestAutoMerge(ctx context.Context) error {
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, name := parts[0], parts[1]
	prNum := parseInt(c.prNum)
	if prNum <= 0 {
		return fmt.Errorf("invalid PR number: %s", c.prNum)
	}

	prID, err := c.queryPRNodeID(ctx, owner, name, prNum)
	if err != nil {
		return err
	}

	mutation := `mutation($pullRequestId: ID!) {
  enablePullRequestAutoMerge(input: { pullRequestId: $pullRequestId, mergeMethod: MERGE }) {
    clientMutationId
  }
}`
	vars := map[string]interface{}{"pullRequestId": prID}
	_, err = c.graphQLRequest(ctx, mutation, vars)
	if err != nil {
		return fmt.Errorf("enable automerge: %w", err)
	}
	return nil
}
