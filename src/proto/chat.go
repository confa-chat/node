package proto

import (
	"context"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/auth"
	"github.com/konfa-chat/hub/src/konfa"
	chatv1 "github.com/konfa-chat/hub/src/proto/konfa/chat/v1"
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
	for {
		req, err := req.Recv()
		if err != nil {
			return err
		}
		_ = req
	}
}
