package proto

import (
	"context"
	"fmt"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/auth"
	"github.com/confa-chat/node/src/confa"
	nodev1 "github.com/confa-chat/node/src/proto/confa/node/v1"

	"github.com/Masterminds/semver/v3"
)

func NewHubService(srv *confa.Service) *NodeService {
	return &NodeService{srv: srv}
}

type NodeService struct {
	srv *confa.Service
}

var _ nodev1.NodeServiceServer = (*NodeService)(nil)

var MinVersion = semver.MustParse("0.0.1-beta.1")

// SupportedClientVersions implements nodev1.NodeServiceServer.
func (h *NodeService) SupportedClientVersions(ctx context.Context, req *nodev1.SupportedClientVersionsRequest) (*nodev1.SupportedClientVersionsResponse, error) {
	clientVer, err := semver.NewVersion(req.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	return &nodev1.SupportedClientVersionsResponse{
		Supported:  clientVer.GreaterThanEqual(MinVersion),
		MinVersion: MinVersion.String(),
	}, nil
}

// ListAuthProviders implements nodev1.HubServiceServer.
func (h *NodeService) ListAuthProviders(context.Context, *nodev1.ListAuthProvidersRequest) (*nodev1.ListAuthProvidersResponse, error) {
	// Return the auth providers from the config instead of directly from the service
	return &nodev1.ListAuthProvidersResponse{
		AuthProviders: h.srv.Config.GetHubAuthProviders(),
	}, nil
}

// ListVoiceRelays implements nodev1.HubServiceServer.
func (h *NodeService) ListVoiceRelays(context.Context, *nodev1.ListVoiceRelaysRequest) (*nodev1.ListVoiceRelaysResponse, error) {
	// Return the voice relays from the config instead of hardcoded values
	return &nodev1.ListVoiceRelaysResponse{
		VoiceRelays: h.srv.Config.GetHubVoiceRelays(),
	}, nil
}

// ListServers implements nodev1.HubServiceServer.
func (h *NodeService) ListServerIDs(ctx context.Context, req *nodev1.ListServersRequest) (*nodev1.ListServersResponse, error) {
	servers, err := h.srv.ListServers(ctx)
	if err != nil {
		return nil, err
	}
	serverIds := make([]string, 0, len(servers))
	for _, server := range servers {
		serverIds = append(serverIds, server.ID.String())
	}
	return &nodev1.ListServersResponse{
		ServerIds: serverIds,
	}, nil
}

// GetUser implements serverv1.ServerServiceServer.
func (s *NodeService) GetUser(ctx context.Context, req *nodev1.GetUserRequest) (*nodev1.GetUserResponse, error) {
	userID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, err
	}

	users, err := s.srv.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &nodev1.GetUserResponse{
		User: mapUser(users),
	}, nil
}

// CurrentUser implements serverv1.ServerServiceServer.
func (s *NodeService) CurrentUser(ctx context.Context, req *nodev1.CurrentUserRequest) (*nodev1.CurrentUserResponse, error) {
	user := auth.CtxGetUser(ctx)
	if user == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	return &nodev1.CurrentUserResponse{
		User: mapUser(*user),
	}, nil
}
