package confa

import (
	"context"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/store"
)

func (c *Service) CreateServer(ctx context.Context, name string) (uuid.UUID, error) {
	log := c.log.With("name", name)

	server := store.Server{
		ID:   uuid.New(),
		Name: name,
	}

	var idrow store.IDRow
	_, err := c.db.NewInsert().Model(&server).Returning("id").Exec(ctx, &idrow)
	if err != nil {
		log.Error("failed to create server", "error", err)
		return uuid.Nil, err
	}

	return idrow.ID, err
}

func (c *Service) GetServer(ctx context.Context, serverID uuid.UUID) ([]store.Server, error) {
	log := c.log.With("serverID", serverID)

	var servers []store.Server
	err := c.db.NewSelect().
		Model(&servers).
		Where("server_id = ?", serverID).
		Scan(ctx)
	if err != nil {
		log.Error("failed to get server", "error", err)
		return nil, err
	}

	return servers, err
}

func (c *Service) ListServers(ctx context.Context) ([]store.Server, error) {
	log := c.log.With()

	var servers []store.Server
	err := c.db.NewSelect().
		Model(&servers).
		Scan(ctx)

	if err != nil {
		log.Error("failed to list servers", "error", err)
		return nil, err
	}

	return servers, err
}
