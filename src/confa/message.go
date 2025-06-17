package confa

import (
	"context"
	"fmt"
	"time"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/store"
	"github.com/cskr/pubsub/v2"
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
		Relation("Attachments").
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
		Relation("Attachments").
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

// SendMessageWithAttachments creates a new message with the specified attachments
func (c *Service) SendMessageWithAttachments(ctx context.Context, senderID, serverID, channelID uuid.UUID, content string, attachmentIDs []uuid.UUID, attachmentNames []string) (uuid.UUID, error) {
	// Create transaction
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	// Create the message
	msgID := uuid.New()
	msg := store.Message{
		ID:        msgID,
		Timestamp: time.Now(),
		ChannelID: channelID,
		SenderID:  senderID,
		Content:   content,
	}

	// Insert the message
	_, err = tx.NewInsert().Model(&msg).Exec(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	// Add attachments if any
	if len(attachmentIDs) > 0 {
		// Ensure we have the same number of names as IDs
		if len(attachmentIDs) != len(attachmentNames) {
			return uuid.Nil, fmt.Errorf("number of attachment IDs (%d) doesn't match number of names (%d)", len(attachmentIDs), len(attachmentNames))
		}

		// Create attachment records
		attachments := make([]store.MessageAttachment, len(attachmentIDs))
		for i := range attachmentIDs {
			attachments[i] = store.MessageAttachment{
				ID:           uuid.New(),
				MessageID:    msgID,
				Name:         attachmentNames[i],
				AttachmentID: attachmentIDs[i],
			}
		}

		// Insert all attachments
		_, err = tx.NewInsert().Model(&attachments).Exec(ctx)
		if err != nil {
			return uuid.Nil, err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}

	// Publish message to subscribers
	c.msgBroker.Pub(msgID, channelID)

	return msgID, nil
}
