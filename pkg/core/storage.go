package core

import (
	"context"

	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

type AccountEventsSubscription struct {
	TelegramUserID telegram.UserID
	Account        ton.AccountID
}

type BridgeSubscription struct {
	TelegramUserID telegram.UserID
	ClientID       ClientID
	Origin         string
}

type Storage interface {
	SubscribeToAccountEvents(ctx context.Context, userID telegram.UserID, account ton.Address) error
	GetAccountEventsSubscriptions(ctx context.Context) ([]AccountEventsSubscription, error)
	UnsubscribeAccountEvents(ctx context.Context, userID telegram.UserID) error

	SubscribeToBridgeEvents(ctx context.Context, userID telegram.UserID, clientID ClientID, origin string) error
	UnsubscribeFromBridgeEvents(ctx context.Context, userID telegram.UserID, clientID *ClientID) error

	GetBridgeSubscriptions(ctx context.Context) ([]BridgeSubscription, error)
}
