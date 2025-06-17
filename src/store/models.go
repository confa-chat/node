package store

import (
	"time"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/uptrace/bun"
)

type Server struct {
	bun.BaseModel `bun:"table:server"`

	ID   uuid.UUID `bun:"id,pk"`
	Name string    `bun:"name"`
}

type TextChannel struct {
	bun.BaseModel `bun:"table:text_channel"`

	ID       uuid.UUID `bun:"id,pk"`
	ServerID uuid.UUID `bun:"server_id"`
	Name     string    `bun:"name"`
}

type MessageAttachment struct {
	bun.BaseModel `bun:"table:message_attachment"`

	ID           uuid.UUID `bun:"id,pk"`
	MessageID    uuid.UUID `bun:"message_id"`
	Name         string    `bun:"name"`
	AttachmentID uuid.UUID `bun:"attachment_id"`
}

type Message struct {
	bun.BaseModel `bun:"table:message"`

	ID uuid.UUID `bun:"id,pk"`
	// ServerID  uuid.UUID `bun:"server_id"`
	ChannelID uuid.UUID `bun:"channel_id"`
	SenderID  uuid.UUID `bun:"sender_id"`
	Content   string    `bun:"content"`
	Timestamp time.Time `bun:"timestamp"`

	Attachments []MessageAttachment `bun:"rel:has-many,join:id=message_id"`
}

type VoiceChannel struct {
	bun.BaseModel `bun:"table:voice_channel"`

	ID       uuid.UUID `bun:"id,pk"`
	ServerID uuid.UUID `bun:"server_id"`
	Name     string    `bun:"name"`
	RelayID  string    `bun:"relay_id"`
}
