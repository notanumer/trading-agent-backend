package bot

import (
	"context"
	"sync"
	"time"

	"deepseek-trader/agent"
	"deepseek-trader/config"
	"deepseek-trader/hyperliquid"
	"deepseek-trader/models"
	"deepseek-trader/services"
)

type DeepSeekAgent interface {
	Decide(ctx context.Context, snap agent.Snapshot) (agent.Decision, error)
}

type Service struct {
	mx     sync.RWMutex
	on     bool
	cancel context.CancelFunc

	hl        *hyperliquid.Client
	tradesSvc *services.TradesService
	statsSvc  *services.StatsService
	cfg       *config.Settings
	agent     DeepSeekAgent
}

func NewService(hl *hyperliquid.Client, tradesSvc *services.TradesService, statsSvc *services.StatsService, cfg *config.Settings) *Service {
	ag := agent.NewDeepseekAgent(cfg)
	return &Service{hl: hl, tradesSvc: tradesSvc, statsSvc: statsSvc, cfg: cfg, agent: ag}
}

func (s *Service) Start() {
	s.mx.Lock()
	if s.on {
		s.mx.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.on = true
	s.mx.Unlock()

	go s.loop(ctx)
}

func (s *Service) Stop() {
	s.mx.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.on = false
	s.mx.Unlock()
}

func (s *Service) IsOn() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.on
}

func (s *Service) loop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := s.hl.GetLiveStats(ctx)
			if err != nil {
				continue
			}

			coinsMids, err := s.hl.CoinsMids(ctx)
			if err != nil {
				continue
			}

			filtered := agent.FilterCoinsMids(coinsMids)
			snap := agent.Snapshot{Balance: stats.Balance, PnL: stats.PnL, ROE: stats.ROE, CoinsMids: filtered}

			hist, err := s.hl.HistoricalOrders(ctx)
			if err != nil {
				continue
			}

			decisions, err := s.tradesSvc.LatestDecisions(ctx, 10)
			if err != nil {
				continue
			}

			for _, d := range decisions {
				snap.Decisions = append(snap.Decisions, d)
			}
			for _, t := range hist {
				snap.Trades = append(snap.Trades, t)
			}

			snap.Balance += 10000
			snap.PnL += 10000
			snap.ROE += 10000
			dec, err := s.agent.Decide(ctx, snap)
			d := models.Decision{
				Action:     dec.Action,
				Symbol:     dec.Symbol,
				Size:       dec.Size,
				OrderType:  dec.Order,
				LimitPrice: dec.LimitPrice,
				TP1:        dec.Targets.TP1,
				TP2:        dec.Targets.TP2,
				TP3:        dec.Targets.TP3,
				SL:         dec.Targets.SL,
			}
			_, _ = s.tradesSvc.RecordDecision(ctx, d)
			if err != nil || dec.Action == "none" {
				continue
			}

			_, _ = s.tradesSvc.Record(ctx, dec.Symbol, dec.Action, dec.Size, dec.LimitPrice)
		}
	}
}
