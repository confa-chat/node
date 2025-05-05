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

// CreateChannel implements serverv1.ServerServiceServer.
func (s *ServerService) CreateChannel(ctx context.Context, req *serverv1.CreateChannelRequest) (*serverv1.CreateChannelResponse, error) {
	serverID, err := uuid.FromString(req.ServerId)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	var channelID uuid.UUID
	var channel *channelv1.Channel

	// Create either a text or voice channel based on the type
	switch req.Type {
	case serverv1.CreateChannelRequest_TEXT:
		// Create a text channel
		channelID, err = s.srv.CreateTextChannel(ctx, serverID, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to create text channel: %w", err)
		}

		// Create the response with a text channel
		channel = &channelv1.Channel{
			Channel: &channelv1.Channel_TextChannel{
				TextChannel: &channelv1.TextChannel{
					ServerId:  serverID.String(),
					ChannelId: channelID.String(),
					Name:      req.Name,
				},
			},
		}

	case serverv1.CreateChannelRequest_VOICE:
		// Create a voice channel using the service method
		channelID, err = s.srv.CreateVoiceChannel(ctx, serverID, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to create voice channel: %w", err)
		}

		// Create the response with a voice channel
		channel = &channelv1.Channel{
			Channel: &channelv1.Channel_VoiceChannel{
				VoiceChannel: &channelv1.VoiceChannel{
					ServerId:     serverID.String(),
					ChannelId:    channelID.String(),
					Name:         req.Name,
					VoiceRelayId: "", // Empty for now
				},
			},
		}

	default:
		return nil, fmt.Errorf("unknown channel type: %v", req.Type)
	}

	return &serverv1.CreateChannelResponse{
		Channel: channel,
	}, nil
}
