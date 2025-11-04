package handlers

import (
	"deepseek-trader/bot"
	"deepseek-trader/hyperliquid"
	"deepseek-trader/services"
)

type Handler struct {
	wallet  *services.WalletService
	botSvc  *bot.Service
	stats   *services.StatsService
	trades  *services.TradesService
	authSvc *services.AuthService
	hl      hyperliquid.Client
}

func New(
	wallet *services.WalletService,
	botSvc *bot.Service, stats *services.StatsService,
	trades *services.TradesService, authSvc *services.AuthService,
	hl hyperliquid.Client,
) *Handler {
	return &Handler{
		wallet:  wallet,
		botSvc:  botSvc,
		stats:   stats,
		trades:  trades,
		authSvc: authSvc,
		hl:      hl,
	}
}
