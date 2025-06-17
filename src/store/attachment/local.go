package attachment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/confa-chat/node/pkg/uuid"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local filesystem storage at the specified directory
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create attachments directory: %w", err)
	}

	// Create metadata directory
	metadataDir := filepath.Join(basePath, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// Upload saves the attachment to the local filesystem
func (s *LocalStorage) Upload(ctx context.Context, name string, data io.Reader) (AttachmentInfo, error) {
	// Generate a unique ID for this attachment
	id := uuid.New()

	// Create the file path
	filePath := s.getPath(id)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return AttachmentInfo{}, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the data to the file
	if _, err := io.Copy(file, data); err != nil {
		return AttachmentInfo{}, fmt.Errorf("failed to write data: %w", err)
	}

	// Create and save metadata
	info := AttachmentInfo{
		ID:       id,
		Filename: name,
	}

	if err := s.saveMetadata(info); err != nil {
		return AttachmentInfo{}, err
	}

	return info, nil
}

// Get retrieves an attachment from the local filesystem
func (s *LocalStorage) Get(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	filePath := s.getPath(id)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("attachment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to open attachment: %w", err)
	}

	return file, nil
}

// GetInfo retrieves metadata about an attachment
func (s *LocalStorage) GetInfo(ctx context.Context, id uuid.UUID) (AttachmentInfo, error) {
	metadataPath := s.getMetadataPath(id)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return AttachmentInfo{}, fmt.Errorf("attachment metadata not found: %s", id)
		}
		return AttachmentInfo{}, fmt.Errorf("failed to read attachment metadata: %w", err)
	}

	var info AttachmentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return AttachmentInfo{}, fmt.Errorf("failed to parse attachment metadata: %w", err)
	}

	return info, nil
}

// Delete removes an attachment from the local filesystem
func (s *LocalStorage) Delete(ctx context.Context, id uuid.UUID) error {
	filePath := s.getPath(id)
	metadataPath := s.getMetadataPath(id)

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete attachment: %w", err)
		}
	}

	// Delete the metadata
	if err := os.Remove(metadataPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete attachment metadata: %w", err)
		}
	}

	return nil
}

// getPath returns the full path for an attachment ID
func (s *LocalStorage) getPath(id uuid.UUID) string {
	return filepath.Join(s.basePath, id.String())
}

// getMetadataPath returns the path to the metadata file for an attachment
func (s *LocalStorage) getMetadataPath(id uuid.UUID) string {
	return filepath.Join(s.basePath, "metadata", id.String()+".json")
}

// saveMetadata saves attachment metadata to a JSON file
func (s *LocalStorage) saveMetadata(info AttachmentInfo) error {
	metadataPath := s.getMetadataPath(info.ID)

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal attachment metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write attachment metadata: %w", err)
	}

	return nil
}
