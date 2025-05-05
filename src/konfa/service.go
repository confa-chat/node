package konfa

import (
	"github.com/cskr/pubsub/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/konfa-chat/hub/src/config"
	"github.com/uptrace/bun"
)

type Service struct {
	db     *bun.DB
	dbpool *pgxpool.Pool

	msgBroker *pubsub.PubSub[uuid.UUID, uuid.UUID]

	Config *config.Config
}

func NewService(db *bun.DB, dbpool *pgxpool.Pool, cfg *config.Config) *Service {
	return &Service{
		db:        db,
		dbpool:    dbpool,
		msgBroker: pubsub.New[uuid.UUID, uuid.UUID](10),
		Config:    cfg,
	}
}
