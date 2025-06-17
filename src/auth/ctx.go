package auth

import (
	"context"

	"github.com/confa-chat/node/src/store"
)

type ctxKey string

const ctxUserKey ctxKey = "user"

func ctxWithUser(ctx context.Context, user store.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

func CtxGetUser(ctx context.Context) *store.User {
	userI := ctx.Value(ctxUserKey)
	if userI == nil {
		return nil
	}

	user, ok := userI.(store.User)
	if !ok {
		return nil
	}

	return &user
}
