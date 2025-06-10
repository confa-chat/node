package proto

import (
	"context"
	"fmt"
	"io"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/auth"
	"github.com/konfa-chat/hub/src/konfa"
	chatv1 "github.com/konfa-chat/hub/src/proto/konfa/chat/v1"
	"github.com/konfa-chat/hub/src/store/attachment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatService struct {
	srv *konfa.Service
}

func NewChatService(srv *konfa.Service) *ChatService {
	return &ChatService{
		srv: srv,
	}
}

var ErrUnauthenticated = status.Error(codes.Unauthenticated, "unauthenticated")

var _ chatv1.ChatServiceServer = (*ChatService)(nil)

// SendMessage implements chatv1.ChatServiceServer.
func (c *ChatService) SendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	user := auth.CtxGetUser(ctx)
	if user == nil {
		return nil, ErrUnauthenticated
	}
	ref, err := parseChannelRef(req.Channel)
	if err != nil {
		return nil, err
	}

	var attachmentIDs []uuid.UUID
	var attachmentNames []string

	// Process attachments if any
	if len(req.AttachmentIds) > 0 {
		// Convert string IDs to UUID
		attachmentIDs = make([]uuid.UUID, len(req.AttachmentIds))
		attachmentNames = make([]string, len(req.AttachmentIds))

		for i, idStr := range req.AttachmentIds {
			id, err := uuid.FromString(idStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid attachment ID: %v", err)
			}
			attachmentIDs[i] = id

			// Get attachment info to retrieve the original filename
			info, err := c.srv.GetAttachmentInfo(ctx, id)
			if err != nil {
				return nil, status.Errorf(codes.NotFound, "attachment not found: %v", err)
			}
			attachmentNames[i] = info.Filename
		}

		// Send message with attachments
		id, err := c.srv.SendMessageWithAttachments(ctx, user.ID, ref.ServerID, ref.ChannelID, req.Content, attachmentIDs, attachmentNames)
		if err != nil {
			return nil, err
		}

		return &chatv1.SendMessageResponse{MessageId: id.String()}, nil
	}

	// Regular message without attachments
	id, err := c.srv.SendMessage(ctx, user.ID, ref.ServerID, ref.ChannelID, req.Content)
	if err != nil {
		return nil, err
	}

	return &chatv1.SendMessageResponse{MessageId: id.String()}, nil
}

// GetMessage implements chatv1.ChatServiceServer.
func (c *ChatService) GetMessage(ctx context.Context, req *chatv1.GetMessageRequest) (*chatv1.GetMessageResponse, error) {
	ref, err := parseChannelRef(req.Channel)
	if err != nil {
		return nil, err
	}

	messageID, err := uuid.FromString(req.MessageId)
	if err != nil {
		return nil, err
	}

	msg, err := c.srv.GetMessage(ctx, ref.ServerID, ref.ChannelID, messageID)
	if err != nil {
		return nil, err
	}

	return &chatv1.GetMessageResponse{
		Message: mapMessage(msg),
	}, nil
}

// GetMessageHistory implements chatv1.ChatServiceServer.
func (c *ChatService) GetMessageHistory(ctx context.Context, req *chatv1.GetMessageHistoryRequest) (*chatv1.GetMessageHistoryResponse, error) {
	ref, err := parseChannelRef(req.Channel)
	if err != nil {
		return nil, err
	}

	msgs, err := c.srv.GetMessagesHistory(ctx, ref.ServerID, ref.ChannelID, req.From.AsTime(), int(req.Count))
	if err != nil {
		return nil, err
	}

	return &chatv1.GetMessageHistoryResponse{
		Messages: apply(msgs, mapMessage),
	}, nil
}

// StreamNewMessages implements chatv1.ChatServiceServer.
func (c *ChatService) StreamNewMessages(req *chatv1.StreamNewMessagesRequest, out grpc.ServerStreamingServer[chatv1.StreamNewMessagesResponse]) error {
	channelID, err := uuid.FromString(req.Channel.ChannelId)
	if err != nil {
		return err
	}

	sub, err := c.srv.SubscribeNewMessages(out.Context(), channelID)
	if err != nil {
		return err
	}

	defer sub.Close()

	for {
		select {
		case <-out.Context().Done():
			return nil
		case msg := <-sub.Msgs:
			err := out.Send(&chatv1.StreamNewMessagesResponse{MessageId: msg.String()})
			if err != nil {
				return err
			}
		}
	}

}

// UploadAttachment implements chatv1.ChatServiceServer.
func (c *ChatService) UploadAttachment(req grpc.ClientStreamingServer[chatv1.UploadAttachmentRequest, chatv1.UploadAttachmentResponse]) error {
	user := auth.CtxGetUser(req.Context())
	if user == nil {
		return ErrUnauthenticated
	}

	// Create a pipe to stream data to the attachment storage
	pr, pw := io.Pipe()

	// Variables to hold metadata
	var filename string

	// Channel to communicate the result of the upload
	resultCh := make(chan struct {
		info attachment.AttachmentInfo
		err  error
	})

	// Start the upload process in a goroutine
	go func() {
		var info attachment.AttachmentInfo
		var err error

		defer func() {
			resultCh <- struct {
				info attachment.AttachmentInfo
				err  error
			}{info, err}
		}()

		info, err = c.srv.UploadAttachment(req.Context(), filename, pr)
	}()

	// Process the incoming stream of data
	first := true

	for {
		msg, err := req.Recv()
		if err == io.EOF {
			// End of stream, close the writer
			pw.Close()
			break
		}
		if err != nil {
			pw.CloseWithError(err)
			return err
		}

		// Handle the first message which should contain metadata
		if first {
			info := msg.GetInfo()
			if info == nil {
				pw.CloseWithError(fmt.Errorf("first message must contain attachment info"))
				return status.Error(codes.InvalidArgument, "first message must contain attachment info")
			}

			filename = info.Name
			first = false
			continue
		}

		// Process data chunks
		data := msg.GetData()
		if data == nil {
			continue // Skip messages with no data
		}

		// Write the data to the pipe
		if _, err := pw.Write(data); err != nil {
			return err
		}
	}

	// Wait for upload result
	result := <-resultCh
	if result.err != nil {
		return status.Errorf(codes.Internal, "failed to upload attachment: %v", result.err)
	}

	// Return the attachment ID
	return req.SendAndClose(&chatv1.UploadAttachmentResponse{
		AttachmentId: result.info.ID.String(),
	})
}
