package konfa

import (
	"github.com/cskr/pubsub/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/konfa-chat/hub/pkg/uuid"
	"github.com/uptrace/bun"
)

type Service struct {
	db     *bun.DB
	dbpool *pgxpool.Pool

	msgBroker *pubsub.PubSub[uuid.UUID, uuid.UUID]
}

func NewService(db *bun.DB, dbpool *pgxpool.Pool) *Service {
	return &Service{
		db:        db,
		dbpool:    dbpool,
		msgBroker: pubsub.New[uuid.UUID, uuid.UUID](10),
	}
}
