package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/auth"
	"github.com/konfa-chat/hub/src/config"
	"github.com/konfa-chat/hub/src/konfa"
	"github.com/konfa-chat/hub/src/proto"
	chatv1 "github.com/konfa-chat/hub/src/proto/konfa/chat/v1"
	hubv1 "github.com/konfa-chat/hub/src/proto/konfa/hub/v1"
	serverv1 "github.com/konfa-chat/hub/src/proto/konfa/server/v1"
	"github.com/konfa-chat/hub/src/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Parse command line flags
	configFilePath := flag.String("config", "", "Path to YAML configuration file")
	flag.Parse()

	ctx := context.Background()

	// Load configuration passing the config file path directly
	cfg, err := config.Load(*configFilePath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, dbpool, err := store.ConnectPostgres(ctx, cfg.DB)
	if err != nil {
		panic(err)
	}

	// Pass the entire config object to the service
	srv := konfa.NewService(db, dbpool, cfg)

	// Use the first auth provider for the authenticator
	// In a more robust implementation, this might be configurable
	provider := cfg.AuthProviders[0]
	authen, err := auth.NewAuthenticator(ctx, db,
		auth.AuthenticatorConfig{
			Issuer:       provider.OpenIDConnect.Issuer,
			ClientID:     provider.OpenIDConnect.ClientID,
			ClientSecret: provider.OpenIDConnect.ClientSecret,
		},
		[]string{
			"/grpc.reflection.v1alpha.ServerReflection",
			"/" + hubv1.HubService_ServiceDesc.ServiceName,
		},
	)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authen.UnaryAuthenticate),
		grpc.StreamInterceptor(authen.StreamAuthenticate),
	)
	chatv1.RegisterChatServiceServer(grpcServer, proto.NewChatService(srv))
	serverv1.RegisterServerServiceServer(grpcServer, proto.NewServerService(srv))
	hubv1.RegisterHubServiceServer(grpcServer, proto.NewHubService(srv))

	reflection.Register(grpcServer)

	serverID, chanID, err := createKonfach(ctx, srv)
	if err != nil {
		panic(err)
	}

	println(serverID.String())
	println(chanID.String())

	port := 38100

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	println("Server is running on port", port)
	panic(grpcServer.Serve(lis))
}

func createKonfach(ctx context.Context, srv *konfa.Service) (uuid.UUID, uuid.UUID, error) {
	var serverID uuid.UUID

	servers, err := srv.ListServers(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to list servers: %w", err)
	}
	for _, serv := range servers {
		if serv.Name == "konfach" {
			serverID = serv.ID
		}
	}
	if serverID == uuid.Nil {
		serverID, err = srv.CreateServer(ctx, "konfach")
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create server: %w", err)
		}
	}

	var chanID uuid.UUID

	channels, err := srv.ListTextChannelsOnServer(ctx, serverID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	for _, channel := range channels {
		if channel.Name == "general" {
			chanID = channel.ID
		}
	}
	if chanID == uuid.Nil {
		chanID, err = srv.CreateTextChannel(ctx, serverID, "general")
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create channel: %w", err)
		}
	}

	return serverID, chanID, nil
}
