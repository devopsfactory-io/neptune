package lock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"

	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/log"
)

// Ensure GCSStorage implements ObjectStorage.
var _ ObjectStorage = (*GCSStorage)(nil)

// GCSStorage is a GCS-backed store for lock files.
type GCSStorage struct {
	bucketName   string
	prefix       string
	parentFolder string
	client       *storage.Client
	bucket       *storage.BucketHandle
}

// NewGCSStorage creates a GCS storage client. bucketURL is e.g. gs://bucket-name or gs://bucket/prefix; parentFolder is sanitized and used as path segment.
func NewGCSStorage(ctx context.Context, bucketURL, parentFolder string) (*GCSStorage, error) {
	if !strings.HasPrefix(bucketURL, "gs://") {
		return nil, fmt.Errorf("bucket URL must start with gs://")
	}
	pathPart := strings.TrimPrefix(bucketURL, "gs://")
	pathPart = strings.Trim(pathPart, "/")
	var bucketName, prefix string
	if idx := strings.Index(pathPart, "/"); idx >= 0 {
		bucketName = pathPart[:idx]
		prefix = pathPart[idx+1:]
	} else {
		bucketName = pathPart
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	parent := strings.ReplaceAll(parentFolder, "/", "-")
	parent = strings.ReplaceAll(parent, ":", "")
	parent = strings.ReplaceAll(parent, ".", "-")
	return &GCSStorage{
		bucketName:   bucketName,
		prefix:       prefix,
		parentFolder: parent,
		client:       client,
		bucket:       client.Bucket(bucketName),
	}, nil
}

func (s *GCSStorage) objectPath(stackPath string) string {
	if s.prefix != "" {
		return s.prefix + "/" + s.parentFolder + "/" + stackPath + "/lock.json"
	}
	return s.parentFolder + "/" + stackPath + "/lock.json"
}

// GetLockFile returns the lock file for the given stack path, or nil if not found.
func (s *GCSStorage) GetLockFile(ctx context.Context, stackPath string) (*domain.LockFile, error) {
	log.For("lock").Debug("Getting lock file for stack " + stackPath)
	obj := s.bucket.Object(s.objectPath(stackPath))
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lock file for stack %s: %w", stackPath, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Error("close reader", "err", err)
		}
	}()
	var lf domain.LockFile
	if err := json.NewDecoder(reader).Decode(&lf); err != nil {
		return nil, fmt.Errorf("failed to decode lock file for stack %s: %w", stackPath, err)
	}
	return &lf, nil
}

// CreateOrUpdateLockFile writes the lock file for the given stack.
func (s *GCSStorage) CreateOrUpdateLockFile(ctx context.Context, stackPath string, lockData *domain.LockFile) error {
	log.For("lock").Debug("Creating or updating lock file for stack " + stackPath)
	obj := s.bucket.Object(s.objectPath(stackPath))
	data, err := json.MarshalIndent(lockData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}
	w := obj.NewWriter(ctx)
	w.ContentType = "application/json"
	if _, err := w.Write(data); err != nil {
		if closeErr := w.Close(); closeErr != nil {
			return fmt.Errorf("write failed: %w; close: %v", err, closeErr)
		}
		return err
	}
	return w.Close()
}

// DeleteLockFile removes the lock file for the given stack.
func (s *GCSStorage) DeleteLockFile(ctx context.Context, stackPath string) error {
	obj := s.bucket.Object(s.objectPath(stackPath))
	if err := obj.Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf("failed to delete lock file for stack %s: %w", stackPath, err)
	}
	return nil
}

// Close closes the GCS client.
func (s *GCSStorage) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
