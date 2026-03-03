package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devopsfactory-io/neptune/internal/domain"

	"github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"
)

// Client wraps GitHub API for PR checks and optional comment posting.
type Client struct {
	client     *github.Client
	repo       string // owner/repo
	prNum      string
	token      string
	cfg        *domain.NeptuneConfig
	graphQLURL string // if set (e.g. in tests), used for GraphQL requests instead of graphQLURL constant
}

// NewClient builds a GitHub client from config. Repo is normalized from GITHUB_REPOSITORY (may include URL).
func NewClient(cfg *domain.NeptuneConfig) *Client {
	if cfg == nil || cfg.Repository == nil || cfg.Repository.GitHub == nil {
		return nil
	}
	repo := cfg.Repository.GitHub.Repository
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimSuffix(repo, "/")
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Repository.GitHub.Token})
	tc := oauth2.NewClient(context.Background(), ts)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	tc.Transport = &oauth2.Transport{
		Source: ts,
		Base:   httpClient.Transport,
	}
	return &Client{
		client: github.NewClient(tc),
		repo:   repo,
		prNum:  cfg.Repository.GitHub.PullRequestNumber,
		token:  cfg.Repository.GitHub.Token,
		cfg:    cfg,
	}
}

// GetPRInfo fetches the pull request and returns PRInfo or error.
func (c *Client) GetPRInfo(ctx context.Context) (*domain.PRInfo, error) {
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, repoName := parts[0], parts[1]
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repoName, parseInt(c.prNum))
	if err != nil {
		return nil, err
	}
	m := prToMap(pr)
	return &domain.PRInfo{
		Response: m,
		PRNumber: c.prNum,
		Repo:     c.repo,
		APIURL:   "https://api.github.com",
	}, nil
}

// GetHeadSHA returns the head commit SHA of the current PR for use with commit statuses.
func (c *Client) GetHeadSHA(ctx context.Context) (string, error) {
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, repoName := parts[0], parts[1]
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repoName, parseInt(c.prNum))
	if err != nil {
		return "", err
	}
	if pr == nil || pr.Head == nil {
		return "", fmt.Errorf("PR or PR head is nil")
	}
	sha := pr.Head.GetSHA()
	if sha == "" {
		return "", fmt.Errorf("PR head SHA is empty")
	}
	return sha, nil
}

// CreateCommitStatus sets a commit status on the given SHA (context is the status name shown in the UI).
// targetURL may be empty; state must be pending, success, failure, or error.
func (c *Client) CreateCommitStatus(ctx context.Context, sha, contextName, state, description, targetURL string) error {
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, repoName := parts[0], parts[1]
	status := github.RepoStatus{
		State:       &state,
		Context:     &contextName,
		Description: &description,
	}
	if targetURL != "" {
		status.TargetURL = &targetURL
	}
	_, _, err := c.client.Repositories.CreateStatus(ctx, owner, repoName, sha, status)
	return err
}

// IsPROpen returns true if the given PR number is open.
func (c *Client) IsPROpen(ctx context.Context, prNumber string) (bool, error) {
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, repoName := parts[0], parts[1]
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repoName, parseInt(prNumber))
	if err != nil {
		return false, err
	}
	return pr != nil && pr.GetState() == "open", nil
}

// ListReviews returns reviews for the current PR.
func (c *Client) ListReviews(ctx context.Context) ([]*github.PullRequestReview, error) {
	parts := strings.SplitN(c.repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: %s", c.repo)
	}
	owner, repoName := parts[0], parts[1]
	reviews, _, err := c.client.PullRequests.ListReviews(ctx, owner, repoName, parseInt(c.prNum), nil)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func prToMap(pr *github.PullRequest) map[string]interface{} {
	m := make(map[string]interface{})
	if pr == nil {
		return m
	}
	if pr.State != nil {
		m["state"] = *pr.State
	}
	// Mergeable can be nil if GitHub hasn't computed it yet
	if pr.Mergeable != nil {
		m["mergeable"] = *pr.Mergeable
	} else {
		m["mergeable"] = false
	}
	if pr.MergeableState != nil {
		m["mergeable_state"] = *pr.MergeableState
	}
	return m
}
