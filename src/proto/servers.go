package proto

import (
	"context"
	"fmt"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/auth"
	"github.com/konfa-chat/hub/src/konfa"
	channelv1 "github.com/konfa-chat/hub/src/proto/konfa/channel/v1"
	serverv1 "github.com/konfa-chat/hub/src/proto/konfa/server/v1"
	"github.com/konfa-chat/hub/src/store"
)

func NewServerService(srv *konfa.Service) *ServerService {
	return &ServerService{srv: srv}
}

type ServerService struct {
	srv *konfa.Service
}

var _ serverv1.ServerServiceServer = (*ServerService)(nil)

// ListServerChannels implements serverv1.ServerServiceServer.
func (s *ServerService) ListServerChannels(ctx context.Context, req *serverv1.ListServerChannelsRequest) (*serverv1.ListServerChannelsResponse, error) {
	serverID, err := uuid.FromString(req.ServerId)
	if err != nil {
		return nil, err
	}

	textChannels, err := s.srv.ListTextChannelsOnServer(ctx, serverID)
	if err != nil {
		return nil, err
	}

	voiceChannels := []store.VoiceChannel{{
		ID:       serverID,
		ServerID: serverID,
		Name:     "general",
	}}

	channels := make([]*channelv1.Channel, 0, len(textChannels)+len(voiceChannels))
	channels = append(channels, apply(textChannels, mapTextChannelToChannel)...)
	channels = append(channels, apply(voiceChannels, mapVoiceChannelToChannel)...)

	return &serverv1.ListServerChannelsResponse{
		Channels: channels,
	}, nil
}

// ListServerUsers implements serverv1.ServerServiceServer.
func (s *ServerService) ListServerUsers(ctx context.Context, req *serverv1.ListServerUsersRequest) (*serverv1.ListServerUsersResponse, error) {
	serverID, err := uuid.FromString(req.ServerId)
	if err != nil {
		return nil, err
	}

	users, err := s.srv.ListServerUser(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return &serverv1.ListServerUsersResponse{
		Users: apply(users, mapUser),
	}, nil
}

// GetUser implements serverv1.ServerServiceServer.
func (s *ServerService) GetUser(ctx context.Context, req *serverv1.GetUserRequest) (*serverv1.GetUserResponse, error) {
	userID, err := uuid.FromString(req.UserId)
	if err != nil {
		return nil, err
	}

	users, err := s.srv.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &serverv1.GetUserResponse{
		User: mapUser(users),
	}, nil
}

// CurrentUser implements serverv1.ServerServiceServer.
func (s *ServerService) CurrentUser(ctx context.Context, req *serverv1.CurrentUserRequest) (*serverv1.CurrentUserResponse, error) {
	user := auth.CtxGetUser(ctx)
	if user == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	return &serverv1.CurrentUserResponse{
		User: mapUser(*user),
	}, nil
}
