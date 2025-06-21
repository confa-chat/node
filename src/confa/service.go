package confa

import (
	"log/slog"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/confa-chat/node/src/config"
	"github.com/confa-chat/node/src/store/attachment"
	"github.com/cskr/pubsub/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uptrace/bun"
)

type Service struct {
	db            *bun.DB
	dbpool        *pgxpool.Pool
	msgBroker     *pubsub.PubSub[uuid.UUID, uuid.UUID]
	Config        *config.Config
	attachStorage attachment.Storage

	log *slog.Logger
}

func NewService(db *bun.DB, dbpool *pgxpool.Pool, cfg *config.Config, attachStorage attachment.Storage) *Service {
	return &Service{
		db:            db,
		dbpool:        dbpool,
		msgBroker:     pubsub.New[uuid.UUID, uuid.UUID](10),
		Config:        cfg,
		attachStorage: attachStorage,

		log: slog.Default().With(slog.String("service", "confa")),
	}
}
