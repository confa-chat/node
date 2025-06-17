package proto

import (
	"context"
	"fmt"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/auth"
	"github.com/confa-chat/node/src/confa"
	hubv1 "github.com/confa-chat/node/src/proto/confa/hub/v1"

	"github.com/Masterminds/semver/v3"
)

func NewHubService(srv *confa.Service) *HubService {
	return &HubService{srv: srv}
}

type HubService struct {
	srv *confa.Service
}

var _ hubv1.HubServiceServer = (*HubService)(nil)

var MinVersion = semver.MustParse("0.0.1-beta.1")

// SupportedClientVersions implements hubv1.HubServiceServer.
func (h *HubService) SupportedClientVersions(ctx context.Context, req *hubv1.SupportedClientVersionsRequest) (*hubv1.SupportedClientVersionsResponse, error) {
	clientVer, err := semver.NewVersion(req.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	return &hubv1.SupportedClientVersionsResponse{
		Supported:  clientVer.GreaterThanEqual(MinVersion),
		MinVersion: MinVersion.String(),
	}, nil
}

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

// GetUser implements serverv1.ServerServiceServer.
func (s *HubService) GetUser(ctx context.Context, req *hubv1.GetUserRequest) (*hubv1.GetUserResponse, error) {
	userID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, err
	}

	users, err := s.srv.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &hubv1.GetUserResponse{
		User: mapUser(users),
	}, nil
}

// CurrentUser implements serverv1.ServerServiceServer.
func (s *HubService) CurrentUser(ctx context.Context, req *hubv1.CurrentUserRequest) (*hubv1.CurrentUserResponse, error) {
	user := auth.CtxGetUser(ctx)
	if user == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	return &hubv1.CurrentUserResponse{
		User: mapUser(*user),
	}, nil
}
