package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

func TestEnablePullRequestAutoMerge_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		callCount++
		switch callCount {
		case 1:
			// Query for PR node ID
			if len(body.Query) == 0 || body.Query[:5] != "query" {
				t.Errorf("first request expected query, got %q", body.Query)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"repository": map[string]interface{}{
						"pullRequest": map[string]interface{}{
							"id": "PR_kwDOxxxx",
						},
					},
				},
			})
		case 2:
			// enablePullRequestAutoMerge mutation
			if len(body.Query) == 0 || body.Query[:8] != "mutation" {
				t.Errorf("second request expected mutation, got %q", body.Query)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"enablePullRequestAutoMerge": map[string]interface{}{
						"clientMutationId": nil,
					},
				},
			})
		default:
			t.Errorf("unexpected request count %d", callCount)
		}
	}))
	defer server.Close()

	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			GitHub: &domain.GitHubConfig{
				Repository:        "owner/repo",
				PullRequestNumber: "42",
				Token:             "test-token",
			},
		},
	}
	client := NewClient(cfg)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	client.graphQLURL = server.URL

	ctx := context.Background()
	err := client.EnablePullRequestAutoMerge(ctx)
	if err != nil {
		t.Fatalf("EnablePullRequestAutoMerge: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 GraphQL requests, got %d", callCount)
	}
}

func TestEnablePullRequestAutoMerge_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": nil,
			"errors": []map[string]interface{}{
				{"message": "Auto-merge is not enabled for this repository"},
			},
		})
	}))
	defer server.Close()

	cfg := &domain.NeptuneConfig{
		Repository: &domain.RepositoryConfig{
			GitHub: &domain.GitHubConfig{
				Repository:        "owner/repo",
				PullRequestNumber: "1",
				Token:             "test-token",
			},
		},
	}
	client := NewClient(cfg)
	client.graphQLURL = server.URL

	ctx := context.Background()
	err := client.EnablePullRequestAutoMerge(ctx)
	if err == nil {
		t.Fatal("expected error from GraphQL errors response")
	}
	if err.Error() != "GraphQL errors: Auto-merge is not enabled for this repository" {
		t.Errorf("unexpected error: %v", err)
	}
}
