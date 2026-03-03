package lock

import (
	"context"
	"fmt"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

// ObjectStorage is the abstraction for lock file storage (GCS, S3, or S3-compatible).
type ObjectStorage interface {
	GetLockFile(ctx context.Context, stackPath string) (*domain.LockFile, error)
	CreateOrUpdateLockFile(ctx context.Context, stackPath string, lockData *domain.LockFile) error
	DeleteLockFile(ctx context.Context, stackPath string) error
	Close() error
}

// NewObjectStorage creates an ObjectStorage from a bucket URL. bucketURL must be gs://bucket or gs://bucket/prefix for GCS,
// or s3://bucket or s3://bucket/prefix for S3 (including S3-compatible backends like MinIO). parentFolder is sanitized and used as prefix (e.g. repo name).
func NewObjectStorage(ctx context.Context, bucketURL, parentFolder string) (ObjectStorage, error) {
	bucketURL = strings.TrimSpace(bucketURL)
	switch {
	case strings.HasPrefix(bucketURL, "gs://"):
		return NewGCSStorage(ctx, bucketURL, parentFolder)
	case strings.HasPrefix(bucketURL, "s3://"):
		return NewS3Storage(ctx, bucketURL, parentFolder)
	default:
		return nil, fmt.Errorf("object_storage URL must start with gs:// or s3://, got %q", bucketURL)
	}
}
