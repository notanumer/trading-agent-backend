package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq" //nolint:gci
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(db *sql.DB, log *zap.Logger) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	return nil
}
