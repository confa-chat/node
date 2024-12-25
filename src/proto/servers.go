package proto

import (
	"context"

	"github.com/royalcat/konfa-server/pkg/uuid"
	"github.com/royalcat/konfa-server/src/konfa"
	channelv1 "github.com/royalcat/konfa-server/src/proto/konfa/channel/v1"
	serverv1 "github.com/royalcat/konfa-server/src/proto/konfa/server/v1"
	"github.com/royalcat/konfa-server/src/store"
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
