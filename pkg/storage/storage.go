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

	maxSubscriptionsPerUser int
}

var _ core.Storage = (*storage)(nil)

const (
	maxOpenConnections      = 20
	maxSubscriptionsPerUser = 1_000
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
	return &storage{logger: logger, pool: pool, maxSubscriptionsPerUser: maxSubscriptionsPerUser}, nil
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

func (s *storage) SubscribeToBridgeEvents(ctx context.Context, userID telegram.UserID, clientID core.ClientID, origin string) error {
	// TODO: it'd be nice to have a transaction here
	_, err := s.pool.Exec(ctx, "DELETE FROM twa.bridge_subscriptions WHERE telegram_user_id = $1 AND client_id = $2", userID, clientID)
	if err != nil {
		return err
	}
	var subscriptionsCount int
	err = s.pool.QueryRow(ctx, "SELECT count(*) FROM twa.bridge_subscriptions WHERE telegram_user_id = $1", userID).Scan(&subscriptionsCount)
	if err != nil {
		return err
	}
	if subscriptionsCount >= s.maxSubscriptionsPerUser {
		return fmt.Errorf("max subscriptions per user reached")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO twa.bridge_subscriptions (telegram_user_id, client_id, origin) VALUES ($1, $2, $3) 
		ON CONFLICT (telegram_user_id, origin) 
		DO UPDATE set client_id = $2`, userID, clientID, origin)
	return err
}

func (s *storage) UnsubscribeFromBridgeEvents(ctx context.Context, userID telegram.UserID, clientID *core.ClientID) error {
	if clientID == nil {
		_, err := s.pool.Exec(ctx, "DELETE FROM twa.bridge_subscriptions WHERE telegram_user_id = $1", userID)
		return err
	}
	_, err := s.pool.Exec(ctx, "DELETE FROM twa.bridge_subscriptions WHERE telegram_user_id = $1 AND client_id = $2", userID, *clientID)
	return err
}

func (s *storage) GetBridgeSubscriptions(ctx context.Context) ([]core.BridgeSubscription, error) {
	rows, err := s.pool.Query(ctx, "SELECT telegram_user_id, client_id, origin FROM twa.bridge_subscriptions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []core.BridgeSubscription
	for rows.Next() {
		var sub core.BridgeSubscription
		if err := rows.Scan(&sub.TelegramUserID, &sub.ClientID, &sub.Origin); err != nil {
			return nil, err
		}
		result = append(result, sub)
	}
	return result, nil
}
