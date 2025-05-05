package store

import (
	"time"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:user"`

	ID       uuid.UUID `bun:"id,pk"`
	Username string    `bun:"username"`
}

type ExternalLogin struct {
	bun.BaseModel `bun:"table:external_login"`

	ID        uuid.UUID `bun:"id,pk"`
	UserID    uuid.UUID `bun:"user_id"`
	Issuer    string    `bun:"issuer"`
	Subject   string    `bun:"subject"`
	CreatedAt time.Time `bun:"created_at"`
}
