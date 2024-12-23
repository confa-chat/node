package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/royalcat/konfa-server/pkg/uuid"
	"github.com/royalcat/konfa-server/src/konfa"
	"github.com/royalcat/konfa-server/src/proto"
	chatv1 "github.com/royalcat/konfa-server/src/proto/konfa/chat/v1"
	serverv1 "github.com/royalcat/konfa-server/src/proto/konfa/server/v1"
	"github.com/royalcat/konfa-server/src/store"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	db, dbpool, err := store.ConnectPostgres(ctx, "postgres://localhost:5432/konfa?sslmode=disable&user=konfa&password=konfa")
	if err != nil {
		panic(err)
	}

	srv := konfa.NewService(db, dbpool)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(proto.Authenticate))
	chatv1.RegisterChatServiceServer(grpcServer, proto.NewChatService(srv))
	serverv1.RegisterServerServiceServer(grpcServer, proto.NewServerService(srv))

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

	channels, err := srv.ListChannelsOnServer(ctx, serverID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	for _, channel := range channels {
		if channel.Name == "general" {
			chanID = channel.ID
		}
	}
	if chanID == uuid.Nil {
		chanID, err = srv.CreateChannel(ctx, serverID, "general")
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("failed to create channel: %w", err)
		}
	}

	return serverID, chanID, nil
}
