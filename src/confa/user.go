package confa

import (
	"context"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/store"
)

func (u *Service) GetUser(ctx context.Context, id uuid.UUID) (store.User, error) {
	log := u.log.With("id", id)

	var user store.User
	err := u.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		log.Error("failed to get user", "error", err)
		return user, err
	}

	return user, err
}

func (u *Service) ListServerUser(ctx context.Context, serverID uuid.UUID) ([]store.User, error) {
	log := u.log.With("server_id", serverID)

	var users []store.User
	err := u.db.NewSelect().
		Model(&users).
		Scan(ctx)

	if err != nil {
		log.Error("failed to list server users", "error", err)
		return users, err
	}

	return users, err
}
