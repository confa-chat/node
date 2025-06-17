package store

import (
	"context"
	"fmt"

	"github.com/confa-chat/node/pkg/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ReturnIdClause = clause.Returning{Columns: []clause.Column{{Name: "id"}}}

func PgxExecInsertReturningID(ctx context.Context, dbpool *pgxpool.Pool, sql string, args []interface{}) (uuid.UUID, error) {
	rows, err := dbpool.Query(ctx, sql, args...)
	if err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return id, nil
}

type IDRow struct {
	ID uuid.UUID `gorm:"column:id" bun:"id,pk"`
}

func ScanReturnID(tx *gorm.DB) (uuid.UUID, error) {
	if tx.Error != nil {
		return uuid.Nil, fmt.Errorf("error in transaction: %w", tx.Error)
	}

	out := IDRow{
		ID: uuid.Nil,
	}
	tx = tx.Scan(&out)
	if tx.Error != nil {
		return uuid.Nil, tx.Error
	}

	return out.ID, nil
}
