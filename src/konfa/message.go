package konfa

import (
	"context"
	"time"

	"github.com/cskr/pubsub/v2"
	"github.com/royalcat/konfa-server/pkg/uuid"
	"github.com/royalcat/konfa-server/src/store"
)

type ChannelSubscription struct {
	ChannelID uuid.UUID
	Msgs      chan uuid.UUID

	msgBroker *pubsub.PubSub[uuid.UUID, uuid.UUID]
}

func (c *ChannelSubscription) Close() {
	c.msgBroker.Unsub(c.Msgs, c.ChannelID)
	// Drain the channel
	for range c.Msgs {
	}
}

func (c *Service) SubscribeNewMessages(ctx context.Context, channelID uuid.UUID) (*ChannelSubscription, error) {
	sub := c.msgBroker.Sub(channelID)

	return &ChannelSubscription{
		ChannelID: channelID,
		Msgs:      sub,
		msgBroker: c.msgBroker,
	}, nil
}

func (c *Service) GetMessagesHistory(ctx context.Context, serverID uuid.UUID, channelID uuid.UUID, from time.Time, count int) ([]store.Message, error) {
	var messages []store.Message
	err := c.db.NewSelect().
		Model(&messages).
		// Where("server_id = ?", serverID).
		Where("channel_id = ?", channelID).
		Where("timestamp < ?", from).
		Order("timestamp DESC").
		Limit(count).
		Scan(ctx)
	return messages, err
}

func (c *Service) GetMessage(ctx context.Context, serverID, channelID, messageID uuid.UUID) (store.Message, error) {
	var message store.Message
	err := c.db.NewSelect().
		Model(&message).
		// Where("server_id = ?", serverID).
		// Where("channel_id = ?", channelID).
		Where("id = ?", messageID).
		Order("timestamp DESC").
		Scan(ctx)
	return message, err
}

func (c *Service) SendMessage(ctx context.Context, senderID, serverID, channelID uuid.UUID, content string) (uuid.UUID, error) {
	msg := store.Message{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		ChannelID: channelID,
		SenderID:  senderID,
		Content:   content,
	}

	_, err := c.db.NewInsert().Model(&msg).Exec(ctx)

	c.msgBroker.Pub(msg.ID, channelID)

	return msg.ID, err

	// sql, args, err := sq.Insert("channel").
	// 	Columns("id", "server_id", "channel_id", "sender_id", "content").
	// 	Values(id, serverID, channelID, senderID, content).
	// 	Suffix("RETURNING id").
	// 	ToSql()
	// if err != nil {
	// 	return uuid.Nil, err
	// }

	// id, err = store.PgxExecInsertReturningID(ctx, c.dbpool, sql, args)
	// if err != nil {
	// 	return uuid.Nil, err
	// }

	// c.msgBroker.Pub(id, channelID)

	// return id, nil
}
