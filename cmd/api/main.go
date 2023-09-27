package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/storage"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/storage/migrations"
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
	notificator, err := core.NewNotificator(logger, s, cfg.Telegram.BotSecretKey, cfg.TonAPI.ApiKey)
	if err != nil {
		logger.Fatal("core.NewNotificator() failed", zap.Error(err))
	}
	go notificator.Run(context.TODO())

	handler, err := api.NewHandler(logger, notificator, config)
	if err != nil {
		logger.Fatal("api.NewHandler() failed", zap.Error(err))
	}
	server, err := api.NewServer(logger, s.Pool(), handler, fmt.Sprintf(":%v", cfg.API.Port))
	if err != nil {
		logger.Fatal("api.NewServer() failed", zap.Error(err))
	}
	fmt.Printf("running server :%v\n", cfg.API.Port)
	server.Run()
}
