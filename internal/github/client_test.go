package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"

	gh "github.com/google/go-github/v83/github"
)

const baseURLPath = "/api-v3"

// testSetup creates an httptest server and a Neptune github Client configured to use it.
func testSetup(t *testing.T, mux *http.ServeMux) (*Client, func()) {
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	server := httptest.NewServer(apiHandler)

	ghClient := gh.NewClient(nil)
	parsed, err := url.Parse(server.URL + baseURLPath + "/")
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	ghClient.BaseURL = parsed
	ghClient.UploadURL = parsed

	c := &Client{
		client: ghClient,
		repo:   "owner/repo",
		prNum:  "1",
		cfg:    &domain.NeptuneConfig{},
	}
	return c, server.Close
}

func TestGetHeadSHA(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		pr := map[string]interface{}{
			"head": map[string]interface{}{"sha": "abc123"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pr)
	})

	client, teardown := testSetup(t, mux)
	defer teardown()

	ctx := context.Background()
	sha, err := client.GetHeadSHA(ctx)
	if err != nil {
		t.Fatalf("GetHeadSHA: %v", err)
	}
	if sha != "abc123" {
		t.Errorf("GetHeadSHA = %q, want abc123", sha)
	}
}

func TestGetHeadSHA_NoHead(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		pr := map[string]interface{}{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pr)
	})

	client, teardown := testSetup(t, mux)
	defer teardown()

	ctx := context.Background()
	_, err := client.GetHeadSHA(ctx)
	if err == nil {
		t.Error("GetHeadSHA: expected error when PR head is nil")
	}
}

func TestCreateCommitStatus(t *testing.T) {
	mux := http.NewServeMux()
	var gotBody struct {
		State       *string `json:"state"`
		Context     *string `json:"context"`
		Description *string `json:"description"`
		TargetURL   *string `json:"target_url"`
	}
	mux.HandleFunc("/repos/owner/repo/statuses/abc123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	})

	client, teardown := testSetup(t, mux)
	defer teardown()

	ctx := context.Background()
	err := client.CreateCommitStatus(ctx, "abc123", "neptune apply", "pending", "Waiting for apply…", "https://example.com/run")
	if err != nil {
		t.Fatalf("CreateCommitStatus: %v", err)
	}
	if gotBody.State == nil || *gotBody.State != "pending" {
		t.Errorf("state = %v, want pending", gotBody.State)
	}
	if gotBody.Context == nil || *gotBody.Context != "neptune apply" {
		t.Errorf("context = %v, want neptune apply", gotBody.Context)
	}
	if gotBody.Description == nil || *gotBody.Description != "Waiting for apply…" {
		t.Errorf("description = %v", gotBody.Description)
	}
	if gotBody.TargetURL == nil || *gotBody.TargetURL != "https://example.com/run" {
		t.Errorf("target_url = %v", gotBody.TargetURL)
	}
}

func TestCreateCommitStatus_EmptyTargetURL(t *testing.T) {
	mux := http.NewServeMux()
	var gotBody struct {
		TargetURL *string `json:"target_url"`
	}
	mux.HandleFunc("/repos/owner/repo/statuses/sha", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	})

	client, teardown := testSetup(t, mux)
	defer teardown()

	ctx := context.Background()
	err := client.CreateCommitStatus(ctx, "sha", "neptune plan", "success", "Plan completed", "")
	if err != nil {
		t.Fatalf("CreateCommitStatus: %v", err)
	}
	if gotBody.TargetURL != nil && *gotBody.TargetURL != "" {
		t.Errorf("target_url should be omitted or empty, got %q", *gotBody.TargetURL)
	}
}
