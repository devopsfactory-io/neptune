package lock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"neptune/internal/domain"
)

// GCSStorage is a GCS-backed store for lock files.
type GCSStorage struct {
	bucketName   string
	parentFolder string
	client       *storage.Client
	bucket       *storage.BucketHandle
}

// NewGCSStorage creates a GCS storage client. bucketURL is e.g. gs://bucket-name; parentFolder is sanitized and used as prefix.
func NewGCSStorage(ctx context.Context, bucketURL, parentFolder string) (*GCSStorage, error) {
	if !strings.HasPrefix(bucketURL, "gs://") {
		return nil, fmt.Errorf("bucket URL must start with gs://")
	}
	bucketName := strings.TrimPrefix(bucketURL, "gs://")
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	parent := strings.ReplaceAll(parentFolder, "/", "-")
	parent = strings.ReplaceAll(parent, ":", "")
	parent = strings.ReplaceAll(parent, ".", "-")
	return &GCSStorage{
		bucketName:   bucketName,
		parentFolder: parent,
		client:       client,
		bucket:       client.Bucket(bucketName),
	}, nil
}

// GetLockFile returns the lock file for the given stack path, or nil if not found.
func (s *GCSStorage) GetLockFile(ctx context.Context, stackPath string) (*domain.LockFile, error) {
	obj := s.bucket.Object(s.parentFolder + "/" + stackPath + "/lock.json")
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lock file for stack %s: %w", stackPath, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close reader: %v\n", err)
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
	obj := s.bucket.Object(s.parentFolder + "/" + stackPath + "/lock.json")
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
	obj := s.bucket.Object(s.parentFolder + "/" + stackPath + "/lock.json")
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
