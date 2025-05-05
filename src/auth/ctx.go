package auth

import (
	"context"

	"github.com/konfa-chat/hub/src/store"
)

type ctxKey string

const ctxUserKey ctxKey = "user"

func CtxGetUser(ctx context.Context) *store.User {
	return ctx.Value(ctxUserKey).(*store.User)
}
