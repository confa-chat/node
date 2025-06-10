package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/konfa-chat/hub/pkg/uuid"
)

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client     *s3.Client
	bucketName string
}

// NewS3Storage creates a new S3 storage with the specified client and bucket
func NewS3Storage(client *s3.Client, bucketName string) *S3Storage {
	return &S3Storage{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload saves the attachment to S3
func (s *S3Storage) Upload(ctx context.Context, name string, data io.Reader) (AttachmentInfo, error) {
	// Read all data into buffer (we need to know the size)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, data); err != nil {
		return AttachmentInfo{}, fmt.Errorf("failed to read attachment data: %w", err)
	}

	// Generate a unique ID for this attachment
	id := uuid.New()
	key := id.String()

	// Create the metadata for the object
	metadata := map[string]string{
		"filename": name,
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/octet-stream"),
		Metadata:    metadata,
	})
	if err != nil {
		return AttachmentInfo{}, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return AttachmentInfo{
		ID:       id,
		Filename: name,
	}, nil
}

// Get retrieves an attachment from S3
func (s *S3Storage) Get(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	key := id.String()

	// Get the object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, nil
}

// GetInfo retrieves metadata about an attachment
func (s *S3Storage) GetInfo(ctx context.Context, id uuid.UUID) (AttachmentInfo, error) {
	key := id.String()

	// Get the object metadata from S3
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if fmt.Sprintf("%T", err) == fmt.Sprintf("%T", noSuchKey) {
			return AttachmentInfo{}, fmt.Errorf("attachment not found: %s", id)
		}
		return AttachmentInfo{}, fmt.Errorf("failed to get object metadata from S3: %w", err)
	}

	// Extract filename from metadata
	filename := ""
	if result.Metadata != nil {
		if filenameVal, ok := result.Metadata["filename"]; ok {
			filename = filenameVal
		}
	}

	return AttachmentInfo{
		ID:       id,
		Filename: filename,
	}, nil
}

// Delete removes an attachment from S3
func (s *S3Storage) Delete(ctx context.Context, id uuid.UUID) error {
	key := id.String()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
