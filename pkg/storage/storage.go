package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
	"go.uber.org/zap"
)

type storage struct {
	logger *zap.Logger
	pool   *pgxpool.Pool
}

var _ core.Storage = (*storage)(nil)

const (
	maxOpenConnections = 20
)

func New(logger *zap.Logger, postgresURI string) (*storage, error) {
	pgxConfig, err := pgxpool.ParseConfig(postgresURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgresURI %v: %v", postgresURI, err)
	}
	pgxConfig.MaxConns = maxOpenConnections
	pool, err := pgxpool.ConnectConfig(context.TODO(), pgxConfig)
	if err != nil {
		return nil, err
	}
	return &storage{logger: logger, pool: pool}, nil
}

func (s *storage) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *storage) SubscribeToAccountEvents(ctx context.Context, userID telegram.UserID, addr ton.Address) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO twa.subscriptions (telegram_user_id, account) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, addr.ID.ToRaw())
	return err
}

func (s *storage) GetAccountEventsSubscriptions(ctx context.Context) ([]core.AccountEventsSubscription, error) {
	rows, err := s.pool.Query(ctx, "SELECT telegram_user_id, account FROM twa.subscriptions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []core.AccountEventsSubscription
	for rows.Next() {
		var sub core.AccountEventsSubscription
		var accountID string
		if err := rows.Scan(&sub.TelegramUserID, &accountID); err != nil {
			return nil, err
		}
		account, err := ton.ParseAccountID(accountID)
		if err != nil {
			return nil, err
		}
		sub.Account = account
		result = append(result, sub)
	}
	return result, nil
}

func (s *storage) UnsubscribeAccountEvents(ctx context.Context, userID telegram.UserID) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM twa.subscriptions WHERE telegram_user_id = $1", userID)
	return err
}
