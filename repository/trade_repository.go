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

	//go:embed sql/trade/create_decision.sql
	createDecisionSQL string

	//go:embed sql/trade/latest_dicisions.sql
	latestDecisionsSQL string
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

func (r *TradeRepository) CreateDecision(ctx context.Context, d *models.Decision) error {
	return r.db.
		QueryRowxContext(ctx, createDecisionSQL, d.Action, d.Symbol, d.Size, d.OrderType, d.LimitPrice, d.TP1, d.TP2, d.TP3, d.SL).
		Scan(&d.ID, &d.CreatedAt)
}

func (r *TradeRepository) LatestDecisions(ctx context.Context, limit int) ([]models.Decision, error) {
	var items []models.Decision

	if err := r.db.SelectContext(ctx, &items, latestDecisionsSQL, limit); err != nil {
		return nil, err
	}
	return items, nil
}
