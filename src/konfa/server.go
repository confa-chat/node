package konfa

import (
	"context"

	"github.com/royalcat/konfa-server/pkg/uuid"
	"github.com/royalcat/konfa-server/src/store"
)

func (c *Service) CreateServer(ctx context.Context, name string) (uuid.UUID, error) {
	server := store.Server{
		ID:   uuid.New(),
		Name: name,
	}

	var idrow store.IDRow
	_, err := c.db.NewInsert().Model(&server).Returning("id").Exec(ctx, &idrow)
	return idrow.ID, err
}

func (c *Service) GetServer(ctx context.Context, serverID uuid.UUID) ([]store.Server, error) {
	var servers []store.Server
	err := c.db.NewSelect().
		Model(&servers).
		Where("server_id = ?", serverID).
		Scan(ctx)
	return servers, err
}

func (c *Service) ListServers(ctx context.Context) ([]store.Server, error) {
	var servers []store.Server
	err := c.db.NewSelect().
		Model(&servers).
		Scan(ctx)
	return servers, err
}
