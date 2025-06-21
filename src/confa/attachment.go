package confa

import (
	"context"
	"io"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/store/attachment"
)

// UploadAttachment handles storing an attachment and returns the attachment info
func (s *Service) UploadAttachment(ctx context.Context, filename string, data io.Reader) (attachment.AttachmentInfo, error) {
	log := s.log.With("filename", filename)
	info, err := s.attachStorage.Upload(ctx, filename, data)
	if err != nil {
		log.Error("failed to upload attachment", "error", err)
		return attachment.AttachmentInfo{}, err
	}
	return info, nil
}

// GetAttachment retrieves an attachment by its ID
func (s *Service) GetAttachment(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	log := s.log.With("id", id)
	r, err := s.attachStorage.Get(ctx, id)
	if err != nil {
		log.Error("failed to get attachment", "error", err)
		return nil, err
	}
	return r, nil
}

// GetAttachmentInfo retrieves attachment metadata by its ID
func (s *Service) GetAttachmentInfo(ctx context.Context, id uuid.UUID) (attachment.AttachmentInfo, error) {
	log := s.log.With("id", id)
	info, err := s.attachStorage.GetInfo(ctx, id)
	if err != nil {
		log.Error("failed to get attachment info", "error", err)
		return attachment.AttachmentInfo{}, err
	}
	return info, nil
}

// DeleteAttachment removes an attachment by its ID
func (s *Service) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	log := s.log.With("id", id)
	err := s.attachStorage.Delete(ctx, id)
	if err != nil {
		log.Error("failed to delete attachment", "error", err)
		return err
	}
	return nil
}
