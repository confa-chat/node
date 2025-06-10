package konfa

import (
	"context"
	"io"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/store/attachment"
)

// UploadAttachment handles storing an attachment and returns the attachment info
func (s *Service) UploadAttachment(ctx context.Context, filename string, data io.Reader) (attachment.AttachmentInfo, error) {
	return s.attachStorage.Upload(ctx, filename, data)
}

// GetAttachment retrieves an attachment by its ID
func (s *Service) GetAttachment(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	return s.attachStorage.Get(ctx, id)
}

// GetAttachmentInfo retrieves attachment metadata by its ID
func (s *Service) GetAttachmentInfo(ctx context.Context, id uuid.UUID) (attachment.AttachmentInfo, error) {
	return s.attachStorage.GetInfo(ctx, id)
}

// DeleteAttachment removes an attachment by its ID
func (s *Service) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	return s.attachStorage.Delete(ctx, id)
}
