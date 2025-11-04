package repository

import (
	"context"

	"deepseek-trader/models"

	_ "embed"

	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/trade/create.sql
	createTradeSQL string

	//go:embed sql/trade/select_all_by_limit.sql
	listTradesSQL string
)

type TradeRepository struct {
	db *sqlx.DB
}

func (r *TradeRepository) Create(ctx context.Context, t *models.Trade) error {
	return r.db.
		QueryRowxContext(ctx, createTradeSQL, t.Symbol, t.Side, t.Qty, t.Price, t.PnL).
		Scan(&t.ID, &t.CreatedAt)
}

func (r *TradeRepository) List(ctx context.Context, limit int) ([]models.Trade, error) {
	var items []models.Trade

	if err := r.db.SelectContext(ctx, &items, listTradesSQL, limit); err != nil {
		return nil, err
	}
	return items, nil
}
