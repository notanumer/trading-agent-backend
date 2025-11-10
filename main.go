package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"deepseek-trader/api"
	"deepseek-trader/api/handlers"
	"deepseek-trader/bot"
	"deepseek-trader/config"
	"deepseek-trader/db"
	"deepseek-trader/hyperliquid"
	"deepseek-trader/logger"
	"deepseek-trader/repository"
	"deepseek-trader/services"
)

func main() {
	log := logger.New()
	defer func() {
		err := log.Sync()
		if err != nil {
			log.Sugar().Fatalw("failed to sync log", "error", err)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Sugar().Fatalw("failed to load config", "error", err)
	}

	mainCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	dbConn, err := db.NewPostgres(mainCtx, cfg.DBURL)
	if err != nil {
		log.Sugar().Fatalw("failed to connect postgres", "error", err)
	}
	if err := db.Migrate(dbConn.DB, log); err != nil {
		log.Sugar().Fatalw("failed to migrate schema", "error", err)
	}

	repos := repository.NewRepositories(dbConn)

	hlClient := hyperliquid.NewClient(cfg)

	walletSvc := services.NewWalletService(repos.Wallets, hlClient, cfg)
	tradesSvc := services.NewTradesService(repos.Trades, hlClient)
	statsSvc := services.NewStatsService(repos.Stats, repos.Trades)
	botSvc := bot.NewService(hlClient, tradesSvc, statsSvc, cfg, log)
	authSvc := services.NewAuthService(repos.Users, cfg)
	handlers := handlers.New(walletSvc, botSvc, statsSvc, tradesSvc, authSvc, hlClient)

	router := api.NewRouter(handlers, cfg)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Sugar().Fatalw("server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	botSvc.Stop()
}
