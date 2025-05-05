package store

import (
	"time"

	"github.com/konfa-chat/hub/pkg/uuid"
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

type Message struct {
	bun.BaseModel `bun:"table:message"`

	ID uuid.UUID `bun:"id,pk"`
	// ServerID  uuid.UUID `bun:"server_id"`
	ChannelID uuid.UUID `bun:"channel_id"`
	SenderID  uuid.UUID `bun:"sender_id"`
	Content   string    `bun:"content"`
	Timestamp time.Time `bun:"timestamp"`
}

type VoiceChannel struct {
	bun.BaseModel `bun:"table:voice_channel"`

	ID       uuid.UUID `bun:"id,pk"`
	ServerID uuid.UUID `bun:"server_id"`
	Name     string    `bun:"name"`
}
