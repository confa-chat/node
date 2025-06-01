package auth

import (
	"context"

	"github.com/konfa-chat/hub/src/store"
)

type ctxKey string

const ctxUserKey ctxKey = "user"

func CtxGetUser(ctx context.Context) *store.User {
	user := ctx.Value(ctxUserKey)
	if user == nil {
		return nil
	}
	return user.(*store.User)
}
