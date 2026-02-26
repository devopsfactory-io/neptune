package lock

import (
	"context"
	"testing"
)

func TestNewObjectStorage_InvalidScheme(t *testing.T) {
	ctx := context.Background()
	tests := []string{"", "ftp://bucket", "http://bucket", "file:///tmp", "bucket"}
	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			_, err := NewObjectStorage(ctx, url, "repo")
			if err == nil {
				t.Fatal("expected error for invalid scheme")
			}
			if err != nil && err.Error() == "" {
				t.Fatal("expected non-empty error message")
			}
		})
	}
}

func TestNewObjectStorage_ValidSchemeDispatches(t *testing.T) {
	ctx := context.Background()
	// Should not return scheme error; may return backend error (missing creds).
	for _, url := range []string{"gs://bucket", "s3://bucket"} {
		t.Run(url, func(t *testing.T) {
			_, err := NewObjectStorage(ctx, url, "repo")
			// We expect either success (if creds exist) or a backend error, never "must start with gs:// or s3://"
			if err != nil && err.Error() == "object_storage URL must start with gs:// or s3://, got \""+url+"\"" {
				t.Fatalf("factory should accept %s and dispatch to backend, got scheme error", url)
			}
		})
	}
}
