package auth

import (
	"context"

	"github.com/royalcat/konfa-server/src/store"
)

type ctxKey string

const ctxUserKey ctxKey = "user"

func CtxGetUser(ctx context.Context) *store.User {
	return ctx.Value(ctxUserKey).(*store.User)
}
