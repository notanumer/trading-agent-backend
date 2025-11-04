package bot

import (
	"context"
	"sync"
	"time"

	"deepseek-trader/agent"
	"deepseek-trader/config"
	"deepseek-trader/hyperliquid"
	"deepseek-trader/services"
)

type Service struct {
	mx     sync.RWMutex
	on     bool
	cancel context.CancelFunc

	hl        hyperliquid.Client
	tradesSvc *services.TradesService
	statsSvc  *services.StatsService
	cfg       *config.Settings
	agent     agent.DecisionAgent
}

func NewService(hl hyperliquid.Client, tradesSvc *services.TradesService, statsSvc *services.StatsService, cfg *config.Settings) *Service {
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
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Skip if wallet not set
			stats, err := s.hl.GetLiveStats(ctx)
			if err != nil {
				continue
			}

			// history is optional snapshot context
			hist, _ := s.hl.HistoricalOrders(ctx, 20)
			snap := agent.Snapshot{Balance: stats.Balance, PnL: stats.PnL, ROE: stats.ROE}
			// lightweight trades slice
			for _, t := range hist {
				snap.Trades = append(snap.Trades, t)
			}
			dec, err := s.agent.Decide(ctx, snap)
			if err != nil || dec.Action == "none" {
				continue
			}

			// market order represented with price 0 in our TradesService
			trade, err := s.hl.PlaceOrder(ctx, dec.Symbol, dec.Action, dec.Size, dec.LimitPrice)
			if err != nil {
				continue
			}
			_ = trade
		}
	}
}
