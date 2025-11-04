package services

import (
	"context"

	"deepseek-trader/hyperliquid"
	"deepseek-trader/models"
	"deepseek-trader/repository"
)

type TradesService struct {
	repo *repository.TradeRepository
	hl   hyperliquid.Client
}

func NewTradesService(repo *repository.TradeRepository, hl hyperliquid.Client) *TradesService {
	return &TradesService{repo: repo, hl: hl}
}

func (s *TradesService) Place(ctx context.Context, symbol, side string, qty, price float64) (models.Trade, error) {
	_, _ = s.hl.PlaceOrder(ctx, symbol, side, qty, price)
	t := models.Trade{Symbol: symbol, Side: side, Qty: qty, Price: price, PnL: 0}
	if err := s.repo.Create(ctx, &t); err != nil {
		return models.Trade{}, err
	}
	return t, nil
}

func (s *TradesService) History(ctx context.Context, limit int) ([]models.Trade, error) {
	return s.repo.List(ctx, limit)
}
