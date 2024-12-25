package konfa

import (
	"context"

	"github.com/royalcat/konfa-server/pkg/uuid"
	"github.com/royalcat/konfa-server/src/store"
)

func (c *Service) CreateTextChannel(ctx context.Context, serverID uuid.UUID, name string) (uuid.UUID, error) {
	channel := store.TextChannel{
		ID:       uuid.New(),
		ServerID: serverID,
		Name:     name,
	}

	var idrow store.IDRow
	_, err := c.db.NewInsert().Model(&channel).Returning("id").Exec(ctx, &idrow)
	if err != nil {
		return idrow.ID, err
	}

	return idrow.ID, nil
}

func (c *Service) GetChannel(ctx context.Context, serverID uuid.UUID, channelID uuid.UUID) (store.TextChannel, error) {
	var channel store.TextChannel
	err := c.db.NewSelect().
		Model(&channel).
		// Where("server_id = ?", serverID).
		Where("id = ?", channelID).
		Scan(ctx)

	return channel, err
}

func (c *Service) ListTextChannelsOnServer(ctx context.Context, serverID uuid.UUID) ([]store.TextChannel, error) {
	var channels []store.TextChannel
	err := c.db.NewSelect().
		Model(&channels).
		Where("server_id = ?", serverID).
		Scan(ctx)
	return channels, err
}
