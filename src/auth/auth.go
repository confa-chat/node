package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/store"
	"github.com/uptrace/bun"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

type AuthenticatorConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
}

type Authenticator struct {
	skipAuthMethods []string

	provider rs.ResourceServer
	db       *bun.DB

	logger *slog.Logger
}

func NewAuthenticator(ctx context.Context, db *bun.DB, acfg AuthenticatorConfig, skipAuthMethods []string) (*Authenticator, error) {

	provider, err := rs.NewResourceServerClientCredentials(ctx, acfg.Issuer, acfg.ClientID, acfg.ClientSecret)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		skipAuthMethods: skipAuthMethods,
		provider:        provider,
		db:              db,
		logger:          slog.With("component", "authenticator"),
	}, nil
}

func (a *Authenticator) authorize(ctx context.Context, token string) (store.User, error) {
	// var claims oidc.AccessTokenClaims
	// _, err := oidc.ParseToken(token, &claims)
	// if err != nil {
	// 	return nil, err
	// }

	resp, err := rs.Introspect[*oidc.IntrospectionResponse](ctx, a.provider, token)
	if err != nil {
		return store.User{}, err
	}

	user, err := a.loginWithExternal(ctx, resp)
	if err != nil {
		return store.User{}, err
	}

	return user, nil
}

func (a *Authenticator) loginWithExternal(ctx context.Context, resp *oidc.IntrospectionResponse) (store.User, error) {
	var user store.User
	err := a.db.NewSelect().
		Model(&user).
		Join("JOIN external_login ON \"user\".\"id\" = \"external_login\".\"user_id\"").
		Where("issuer = ?", resp.Issuer).
		Where("subject = ?", resp.Subject).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return a.createUserFromExternal(ctx, resp)
		}

		return store.User{}, err
	}

	return user, nil
}

func (a *Authenticator) createUserFromExternal(ctx context.Context, resp *oidc.IntrospectionResponse) (store.User, error) {
	user := store.User{
		ID:       uuid.New(),
		Username: resp.Username,
	}

	err := a.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().
			Model(&user).
			Exec(ctx)
		if err != nil {
			return err
		}

		_, err = tx.NewInsert().
			Model(&store.ExternalLogin{
				ID:      uuid.New(),
				UserID:  user.ID,
				Issuer:  resp.Issuer,
				Subject: resp.Subject,
			}).
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return store.User{}, err
	}

	return user, nil
}
