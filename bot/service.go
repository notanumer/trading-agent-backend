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

	"go.uber.org/zap"
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
	log       *zap.Logger
}

func NewService(hl *hyperliquid.Client, tradesSvc *services.TradesService, statsSvc *services.StatsService, cfg *config.Settings, log *zap.Logger) *Service {
	ag := agent.NewDeepseekAgent(cfg)
	return &Service{
		hl:        hl,
		tradesSvc: tradesSvc,
		statsSvc:  statsSvc,
		cfg:       cfg,
		agent:     ag,
		log:       log,
	}
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
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			endTime := unixMilli(now)
			startTime := unixMilli(now.Add(-3 * time.Hour))
			stats, err := s.hl.GetLiveStats(ctx)
			if err != nil {
				s.log.Sugar().Errorw("failed to get stats", "error", err)
				continue
			}

			coinsMids, err := s.hl.CoinsMids(ctx)
			if err != nil {
				s.log.Sugar().Errorw("failed to get coin mids", "error", err)
				continue
			}

			meta, err := s.hl.Meta(ctx)
			if err != nil {
				s.log.Sugar().Errorw("failed to get meta", "error", err)
				continue
			}

			var orderBooks []hyperliquid.OrderBookSnapshot
			candleSnapshots := make(map[string][]hyperliquid.Candle, 0)
			for _, coin := range agent.Coins {
				l2Book, err := s.hl.L2Book(ctx, coin)
				if err != nil {
					s.log.Sugar().Errorw("failed to get l2book", "error", err)
					continue
				}
				orderBooks = append(orderBooks, l2Book)

				candleSnapshot, err := s.hl.CandleSnapshot(ctx, coin, startTime, endTime)
				if err != nil {
					s.log.Sugar().Errorw("failed to get candle snapshot", "error", err)
					continue
				}

				candleSnapshots[coin] = candleSnapshot
			}

			filtered := agent.FilterCoinsMids(coinsMids)
			snap := agent.Snapshot{
				Balance:         stats.Balance,
				PnL:             stats.PnL,
				ROE:             stats.ROE,
				CoinsMids:       filtered,
				Meta:            meta,
				OrderBooks:      orderBooks,
				CandleSnapshots: candleSnapshots,
			}

			hist, err := s.hl.HistoricalOrders(ctx)
			if err != nil {
				s.log.Sugar().Errorw("failed to get orders", "error", err)
				continue
			}

			decisions, err := s.tradesSvc.LatestDecisions(ctx, 10)
			if err != nil {
				s.log.Sugar().Errorw("failed to get lates decisions", "error", err)
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
			s.log.Sugar().Infow("start agent", "snapshot", snap)
			dec, err := s.agent.Decide(ctx, snap)
			if err != nil {
				s.log.Sugar().Errorw("failed to get decision", "error", err)
				continue
			}

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

			_, err = s.tradesSvc.RecordDecision(ctx, d)
			if err != nil || dec.Action == "none" {
				continue
			}

			_, err = s.tradesSvc.Record(ctx, dec.Symbol, dec.Action, dec.Size, dec.LimitPrice)
			if err != nil {
				s.log.Sugar().Errorw("failed to record trade", "error", err)
				continue
			}
		}
	}
}

func unixMilli(t time.Time) int64 {
	return t.UnixNano() / 1_000_000
}
