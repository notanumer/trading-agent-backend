package repository

import (
	"context"

	"deepseek-trader/models"

	_ "embed"

	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/stats/create.sql
	createStatsSQL string

	//go:embed sql/stats/find_latest.sql
	latestStatsSQL string
)

type StatsRepository struct {
	db *sqlx.DB
}

func (r *StatsRepository) Create(ctx context.Context, s *models.Stats) error {
	return r.db.QueryRowxContext(ctx, createStatsSQL, s.Balance, s.PnL, s.ROE).Scan(&s.ID, &s.CreatedAt)
}

func (r *StatsRepository) Latest(ctx context.Context) (models.Stats, error) {
	var s models.Stats

	if err := r.db.GetContext(ctx, &s, latestStatsSQL); err != nil {
		return models.Stats{}, err
	}

	return s, nil
}
