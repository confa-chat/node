package confa

import (
	"context"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/store"
)

func (c *Service) CreateTextChannel(ctx context.Context, serverID uuid.UUID, name string) (uuid.UUID, error) {
	log := c.log.With("server_id", serverID, "name", name)

	channel := store.TextChannel{
		ID:       uuid.New(),
		ServerID: serverID,
		Name:     name,
	}

	var idrow store.IDRow
	_, err := c.db.NewInsert().Model(&channel).Returning("id").Exec(ctx, &idrow)
	if err != nil {
		log.Error("failed to create text channel", "error", err)
		return idrow.ID, err
	}

	return idrow.ID, nil
}

// CreateVoiceChannel creates a new voice channel in the specified server
func (c *Service) CreateVoiceChannel(ctx context.Context, serverID uuid.UUID, name string) (uuid.UUID, error) {
	log := c.log.With("server_id", serverID, "name", name)

	channel := store.VoiceChannel{
		ID:       uuid.New(),
		ServerID: serverID,
		Name:     name,
	}

	var idrow store.IDRow
	_, err := c.db.NewInsert().Model(&channel).Returning("id").Exec(ctx, &idrow)
	if err != nil {
		log.Error("failed to create voice channel", "error", err)
		return idrow.ID, err
	}

	return idrow.ID, nil
}

func (c *Service) GetChannel(ctx context.Context, serverID uuid.UUID, channelID uuid.UUID) (store.TextChannel, error) {
	log := c.log.With("server_id", serverID, "channel_id", channelID)

	var channel store.TextChannel
	err := c.db.NewSelect().
		Model(&channel).
		// Where("server_id = ?", serverID).
		Where("id = ?", channelID).
		Scan(ctx)

	if err != nil {
		log.Error("failed to get channel", "error", err)
		return channel, err
	}

	return channel, err
}

func (c *Service) ListTextChannelsOnServer(ctx context.Context, serverID uuid.UUID) ([]store.TextChannel, error) {
	log := c.log.With("server_id", serverID)

	var channels []store.TextChannel
	err := c.db.NewSelect().
		Model(&channels).
		Where("server_id = ?", serverID).
		Scan(ctx)

	if err != nil {
		log.Error("failed to list text channels on server", "error", err)
		return channels, err
	}

	return channels, err
}

// UpdateTextChannel updates an existing text channel
func (c *Service) UpdateTextChannel(ctx context.Context, channelID uuid.UUID, name string) error {
	log := c.log.With("channel_id", channelID, "name", name)

	_, err := c.db.NewUpdate().
		Model((*store.TextChannel)(nil)).
		Set("name = ?", name).
		Where("id = ?", channelID).
		Exec(ctx)

	if err != nil {
		log.Error("failed to update text channel", "channel_id", channelID, "name", name, "error", err)
		return err
	}

	return err
}

// UpdateVoiceChannel updates an existing voice channel
func (c *Service) UpdateVoiceChannel(ctx context.Context, channelID uuid.UUID, name string) error {
	log := c.log.With("channel_id", channelID, "name", name)

	_, err := c.db.NewUpdate().
		Model((*store.VoiceChannel)(nil)).
		Set("name = ?", name).
		Where("id = ?", channelID).
		Exec(ctx)

	if err != nil {
		log.Error("failed to update voice channel", "channel_id", channelID, "name", name, "error", err)
		return err
	}

	return err
}
