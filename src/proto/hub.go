package proto

import (
	"context"

	"github.com/konfa-chat/hub/src/konfa"
	hubv1 "github.com/konfa-chat/hub/src/proto/konfa/hub/v1"
)

func NewHubService(srv *konfa.Service) *HubService {
	return &HubService{srv: srv}
}

type HubService struct {
	srv *konfa.Service
}

var _ hubv1.HubServiceServer = (*HubService)(nil)

// ListAuthProviders implements hubv1.HubServiceServer.
func (h *HubService) ListAuthProviders(context.Context, *hubv1.ListAuthProvidersRequest) (*hubv1.ListAuthProvidersResponse, error) {
	// Return the auth providers from the config instead of directly from the service
	return &hubv1.ListAuthProvidersResponse{
		AuthProviders: h.srv.Config.GetHubAuthProviders(),
	}, nil
}

// ListVoiceRelays implements hubv1.HubServiceServer.
func (h *HubService) ListVoiceRelays(context.Context, *hubv1.ListVoiceRelaysRequest) (*hubv1.ListVoiceRelaysResponse, error) {
	// Return the voice relays from the config instead of hardcoded values
	return &hubv1.ListVoiceRelaysResponse{
		VoiceRelays: h.srv.Config.GetHubVoiceRelays(),
	}, nil
}

// ListServers implements hubv1.HubServiceServer.
func (h *HubService) ListServerIDs(ctx context.Context, req *hubv1.ListServersRequest) (*hubv1.ListServersResponse, error) {
	servers, err := h.srv.ListServers(ctx)
	if err != nil {
		return nil, err
	}
	serverIds := make([]string, 0, len(servers))
	for _, server := range servers {
		serverIds = append(serverIds, server.ID.String())
	}
	return &hubv1.ListServersResponse{
		ServerIds: serverIds,
	}, nil
}
