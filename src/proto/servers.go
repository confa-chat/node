package proto

import (
	"context"
	"fmt"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/confa"
	channelv1 "github.com/confa-chat/node/src/proto/confa/channel/v1"
	serverv1 "github.com/confa-chat/node/src/proto/confa/server/v1"
	"github.com/confa-chat/node/src/store"
)

func NewServerService(srv *confa.Service) *ServerService {
	return &ServerService{srv: srv}
}

type ServerService struct {
	srv *confa.Service
}

var _ serverv1.ServerServiceServer = (*ServerService)(nil)

// ListChannels implements serverv1.ServerServiceServer.
func (s *ServerService) ListChannels(ctx context.Context, req *serverv1.ListChannelsRequest) (*serverv1.ListChannelsResponse, error) {
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
		RelayID:  s.srv.Config.VoiceRelays[0].ID,
	}}

	channels := make([]*channelv1.Channel, 0, len(textChannels)+len(voiceChannels))
	channels = append(channels, apply(textChannels, mapTextChannelToChannel)...)
	channels = append(channels, apply(voiceChannels, mapVoiceChannelToChannel)...)

	return &serverv1.ListChannelsResponse{
		Channels: channels,
	}, nil
}

// ListUsers implements serverv1.ServerServiceServer.
func (s *ServerService) ListUsers(ctx context.Context, req *serverv1.ListUsersRequest) (*serverv1.ListUsersResponse, error) {
	serverID, err := uuid.FromString(req.ServerId)
	if err != nil {
		return nil, err
	}

	users, err := s.srv.ListServerUser(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return &serverv1.ListUsersResponse{
		Users: apply(users, mapUser),
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

		channel = &channelv1.Channel{
			Channel: &channelv1.Channel_VoiceChannel{
				VoiceChannel: &channelv1.VoiceChannel{
					ServerId:     serverID.String(),
					ChannelId:    channelID.String(),
					Name:         req.Name,
					VoiceRelayId: []string{s.srv.Config.VoiceRelays[0].ID},
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
