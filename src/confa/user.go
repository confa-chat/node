package confa

import (
	"context"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/store"
)

func (u *Service) GetUser(ctx context.Context, id uuid.UUID) (store.User, error) {
	var user store.User
	err := u.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)
	return user, err
}

func (u *Service) ListServerUser(ctx context.Context, serverID uuid.UUID) ([]store.User, error) {
	var users []store.User
	err := u.db.NewSelect().
		Model(&users).
		Scan(ctx)
	return users, err
}
