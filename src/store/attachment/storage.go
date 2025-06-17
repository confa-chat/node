package attachment

import (
	"context"
	"io"

	"github.com/confa-chat/node/pkg/uuid"
)

// AttachmentInfo holds metadata about an attachment
type AttachmentInfo struct {
	ID       uuid.UUID
	Filename string
}

// Storage defines a universal interface for storing attachments
type Storage interface {
	// Upload saves the attachment data and returns a unique identifier
	Upload(ctx context.Context, name string, data io.Reader) (AttachmentInfo, error)

	// Get retrieves the attachment data by its ID
	Get(ctx context.Context, id uuid.UUID) (io.ReadCloser, error)

	// GetInfo retrieves metadata about an attachment
	GetInfo(ctx context.Context, id uuid.UUID) (AttachmentInfo, error)

	// Delete removes an attachment by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}
