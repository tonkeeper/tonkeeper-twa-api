package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/storage"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/storage/migrations"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

func createLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if level != "" {
		lvl, err := zapcore.ParseLevel(level)
		if err != nil {
			return nil, err
		}
		cfg.Level.SetLevel(lvl)
	}
	return cfg.Build()
}

func main() {
	cfg := Load()
	logger, err := createLogger(cfg.App.LogLevel)
	if err != nil {
		logger.Fatal("createLogger() failed", zap.Error(err))
	}
	if err := migrations.MigrateDb(cfg.App.PostgresURI); err != nil {
		logger.Fatal("migrateDb() failed", zap.Error(err))
	}
	config := api.Config{
		TonConnectSecret:  cfg.TonConnect.Secret,
		TelegramBotSecret: cfg.Telegram.BotSecretKey,
	}
	s, err := storage.New(logger, cfg.App.PostgresURI)
	if err != nil {
		logger.Fatal("storage.New() failed", zap.Error(err))
	}
	notificator, err := core.NewNotificator(logger, s, cfg.TonAPI.ApiKey)
	if err != nil {
		logger.Fatal("core.NewNotificator() failed", zap.Error(err))
	}
	bot, err := telegram.NewBot(logger, cfg.Telegram.BotSecretKey)
	if err != nil {
		logger.Fatal("telegram.NewBot() failed", zap.Error(err))
	}
	messageCh := bot.Run(context.TODO())

	go notificator.Run(context.TODO(), messageCh)

	bridge, err := core.NewBridge(logger, s, messageCh)
	if err != nil {
		logger.Fatal("core.NewBridge() failed", zap.Error(err))
	}

	handler, err := api.NewHandler(logger, notificator, bridge, config)
	if err != nil {
		logger.Fatal("api.NewHandler() failed", zap.Error(err))
	}
	server, err := api.NewServer(logger, s.Pool(), handler, fmt.Sprintf(":%v", cfg.API.Port))
	if err != nil {
		logger.Fatal("api.NewServer() failed", zap.Error(err))
	}
	metricServer := http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.API.MetricsPort),
		Handler: promhttp.Handler(),
	}
	go func() {
		if err := metricServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen and serve", zap.Error(err))
		}
	}()

	fmt.Printf("running server :%v\n", cfg.API.Port)
	server.Run()
}
