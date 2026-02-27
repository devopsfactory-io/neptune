package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"neptune-lambda/pkg/webhooks"
)

const (
	githubAPI = "https://api.github.com"
	eventType = "neptune-command"
)

// Client calls GitHub API with an installation access token.
type Client struct {
	httpClient *http.Client
	token      string
}

// NewClient creates a client that uses the given installation access token.
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		token:      token,
	}
}

// InstallationToken obtains an installation access token for the given installation ID using the GitHub App JWT.
func InstallationToken(ctx context.Context, appID, privateKeyPEM string, installationID int64) (string, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	appIDNum, _ := strconv.ParseInt(appID, 10, 64)
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": appIDNum,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	jwtStr, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/app/installations/%d/access_tokens", githubAPI, installationID),
		nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request installation token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("installation token: status %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Token, nil
}

// GetPR fetches pull request by number and returns head ref and sha.
func (c *Client) GetPR(ctx context.Context, ownerRepo string, prNumber int) (branch, sha string, err error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d", githubAPI, ownerRepo, prNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("get PR: status %d", resp.StatusCode)
	}
	var pr struct {
		Head struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", "", err
	}
	return pr.Head.Ref, pr.Head.SHA, nil
}

// RepositoryDispatch triggers the repository_dispatch event in the given repo.
func (c *Client) RepositoryDispatch(ctx context.Context, ownerRepo string, payload *webhooks.DispatchPayload) error {
	body := struct {
		EventType     string                    `json:"event_type"`
		ClientPayload *webhooks.DispatchPayload `json:"client_payload"`
	}{
		EventType:     eventType,
		ClientPayload: payload,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/repos/%s/dispatches", githubAPI, ownerRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("repository_dispatch: status %d", resp.StatusCode)
	}
	return nil
}
