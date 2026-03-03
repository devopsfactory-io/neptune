package lock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/log"
)

// Ensure S3Storage implements ObjectStorage.
var _ ObjectStorage = (*S3Storage)(nil)

// S3Storage is an S3-backed store for lock files (AWS S3 or S3-compatible e.g. MinIO).
type S3Storage struct {
	client       *s3.Client
	bucket       string
	prefix       string
	parentFolder string
}

// NewS3Storage creates an S3 storage client. bucketURL is e.g. s3://bucket-name or s3://bucket/prefix; parentFolder is sanitized and used as path segment.
// Credentials and endpoint come from the environment (e.g. AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION; for MinIO set AWS_ENDPOINT_URL_S3).
func NewS3Storage(ctx context.Context, bucketURL, parentFolder string) (*S3Storage, error) {
	if !strings.HasPrefix(bucketURL, "s3://") {
		return nil, fmt.Errorf("bucket URL must start with s3://")
	}
	pathPart := strings.TrimPrefix(bucketURL, "s3://")
	pathPart = strings.Trim(pathPart, "/")
	var bucket, prefix string
	if idx := strings.Index(pathPart, "/"); idx >= 0 {
		bucket = pathPart[:idx]
		prefix = pathPart[idx+1:]
	} else {
		bucket = pathPart
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	// Default region if unset (e.g. MinIO often used without AWS_REGION).
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	// Path-style is required for MinIO and other S3-compatible endpoints so the bucket is in the path (e.g. http://localhost:9000/bucket/...) not the host (http://bucket.localhost:9000/...).
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	parent := strings.ReplaceAll(parentFolder, "/", "-")
	parent = strings.ReplaceAll(parent, ":", "")
	parent = strings.ReplaceAll(parent, ".", "-")
	return &S3Storage{
		client:       client,
		bucket:       bucket,
		prefix:       prefix,
		parentFolder: parent,
	}, nil
}

func (s *S3Storage) objectKey(stackPath string) string {
	if s.prefix != "" {
		return s.prefix + "/" + s.parentFolder + "/" + stackPath + "/lock.json"
	}
	return s.parentFolder + "/" + stackPath + "/lock.json"
}

// GetLockFile returns the lock file for the given stack path, or nil if not found.
func (s *S3Storage) GetLockFile(ctx context.Context, stackPath string) (*domain.LockFile, error) {
	log.For("lock").Debug("Getting lock file for stack " + stackPath)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(stackPath)),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lock file for stack %s: %w", stackPath, err)
	}
	defer func() {
		if out.Body != nil {
			if err := out.Body.Close(); err != nil {
				log.Error("close S3 response body", "err", err)
			}
		}
	}()
	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file for stack %s: %w", stackPath, err)
	}
	var lf domain.LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("failed to decode lock file for stack %s: %w", stackPath, err)
	}
	return &lf, nil
}

// CreateOrUpdateLockFile writes the lock file for the given stack.
func (s *S3Storage) CreateOrUpdateLockFile(ctx context.Context, stackPath string, lockData *domain.LockFile) error {
	log.For("lock").Debug("Creating or updating lock file for stack " + stackPath)
	data, err := json.MarshalIndent(lockData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.objectKey(stackPath)),
		Body:        strings.NewReader(string(data)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to write lock file for stack %s: %w", stackPath, err)
	}
	return nil
}

// DeleteLockFile removes the lock file for the given stack.
func (s *S3Storage) DeleteLockFile(ctx context.Context, stackPath string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(stackPath)),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil
		}
		return fmt.Errorf("failed to delete lock file for stack %s: %w", stackPath, err)
	}
	return nil
}

// Close releases resources. The AWS S3 client does not require explicit close; this is a no-op for compatibility with ObjectStorage.
func (s *S3Storage) Close() error {
	return nil
}
