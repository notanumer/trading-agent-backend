package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //nolint:gci
)

func NewPostgres(ctx context.Context, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		err = db.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close database: %w", err)
		}

		return nil, err
	}

	return db, nil
}
