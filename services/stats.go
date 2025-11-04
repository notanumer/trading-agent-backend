package services

import (
	"context"
	"database/sql"

	"deepseek-trader/models"
	"deepseek-trader/repository"
)

type StatsService struct {
	statsRepo  *repository.StatsRepository
	tradesRepo *repository.TradeRepository
}

func NewStatsService(statsRepo *repository.StatsRepository, tradesRepo *repository.TradeRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo, tradesRepo: tradesRepo}
}

func (s *StatsService) Latest(ctx context.Context) (models.Stats, error) {
	st, err := s.statsRepo.Latest(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			zero := models.Stats{Balance: 0, PnL: 0, ROE: 0}
			if err := s.statsRepo.Create(ctx, &zero); err != nil {
				return models.Stats{}, err
			}
			return zero, nil
		}
		return models.Stats{}, err
	}
	return st, nil
}

func (s *StatsService) Record(ctx context.Context, balance, pnl, roe float64) (models.Stats, error) {
	st := models.Stats{Balance: balance, PnL: pnl, ROE: roe}
	if err := s.statsRepo.Create(ctx, &st); err != nil {
		return models.Stats{}, err
	}
	return st, nil
}
