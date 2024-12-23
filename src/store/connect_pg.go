package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"

	sq "github.com/Masterminds/squirrel"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed pg_migrations/*.sql
var pg_migrations embed.FS

func ConnectPostgres(ctx context.Context, databaseURL string) (*bun.DB, *pgxpool.Pool, error) {
	sq.StatementBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dbconfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	// dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
	// 	pgxuuid.Register(conn.TypeMap())
	// 	return nil
	// }

	if err := postgresMigrate(ctx, dbconfig.ConnConfig.Copy()); err != nil {
		return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, dbconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// db, err := gorm.Open(
	// 	postgres.New(postgres.Config{
	// 		Conn: stdlib.OpenDBFromPool(dbpool),
	// 	}),
	// 	&gorm.Config{
	// 		Logger: logger.New(
	// 			slog.NewWriter(),
	// 			logger.Config{
	// 				SlowThreshold: time.Millisecond,
	// 				LogLevel:      logger.Warn,
	// 				Colorful:      false,
	// 			},
	// 		),
	// 		NamingStrategy: schema.NamingStrategy{
	// 			SingularTable: true,
	// 		},
	// 	})
	// if err != nil {
	// 	return nil, nil, err
	// }

	// err = db.Use(tracing.NewPlugin())
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to use tracing plugin: %w", err)
	// }

	sqldb := stdlib.OpenDBFromPool(dbpool)
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName("mydb")))
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true), // log everything
	))

	return db, dbpool, nil
}

func postgresMigrate(ctx context.Context, dbconfig *pgx.ConnConfig) error {
	goose.SetBaseFS(pg_migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	db, err := sql.Open("pgx", dbconfig.ConnString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	err = goose.UpContext(ctx, db, "pg_migrations")
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
