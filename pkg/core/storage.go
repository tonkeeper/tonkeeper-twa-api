package core

import (
	"context"

	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

type Subscription struct {
	TelegramUserID telegram.UserID
	Account        ton.AccountID
}

type Storage interface {
	SubscribeToAccountEvents(ctx context.Context, userID telegram.UserID, account ton.Address) error
	GetAccountEventsSubscriptions(ctx context.Context) ([]Subscription, error)
	UnsubscribeAccountEvents(ctx context.Context, userID telegram.UserID) error
}
