package core

import (
	"context"

	"github.com/tonkeeper/tongo/ton"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

type mockStorage struct {
	OnUnsubscribeFromBridgeEvents func(ctx context.Context, userID telegram.UserID, clientID *ClientID) error
	OnSubscribeToBridgeEvents     func(ctx context.Context, userID telegram.UserID, clientID ClientID, origin string) error
	OnGetBridgeSubscriptions      func(ctx context.Context) ([]BridgeSubscription, error)
}

func (m *mockStorage) SubscribeToAccountEvents(ctx context.Context, userID telegram.UserID, account ton.Address) error {
	return nil
}

func (m *mockStorage) GetAccountEventsSubscriptions(ctx context.Context) ([]AccountEventsSubscription, error) {
	return nil, nil
}

func (m *mockStorage) UnsubscribeAccountEvents(ctx context.Context, userID telegram.UserID) error {
	return nil
}

func (m *mockStorage) SubscribeToBridgeEvents(ctx context.Context, userID telegram.UserID, clientID ClientID, origin string) error {
	return m.OnSubscribeToBridgeEvents(ctx, userID, clientID, origin)
}

func (m *mockStorage) UnsubscribeFromBridgeEvents(ctx context.Context, userID telegram.UserID, clientID *ClientID) error {
	return m.OnUnsubscribeFromBridgeEvents(ctx, userID, clientID)
}

func (m *mockStorage) GetBridgeSubscriptions(ctx context.Context) ([]BridgeSubscription, error) {
	return m.OnGetBridgeSubscriptions(ctx)
}

var _ Storage = (*mockStorage)(nil)
