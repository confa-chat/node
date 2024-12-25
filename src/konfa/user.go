package konfa

import (
	"context"

	"github.com/royalcat/konfa-server/pkg/uuid"
	"github.com/royalcat/konfa-server/src/store"
	"github.com/uptrace/bun"
)

type Users struct {
	db *bun.DB
}

func NewUsers(db *bun.DB) *Users {
	return &Users{db: db}
}

func (u *Users) GetUser(ctx context.Context, id uuid.UUID) (store.User, error) {
	var user store.User
	err := u.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)
	return user, err
}
