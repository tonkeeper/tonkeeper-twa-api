package core

import (
	"context"

	"github.com/tonkeeper/tongo/ton"
)

type Subscription struct {
	TelegramUserID TelegramUserID
	Account        ton.AccountID
}

type Storage interface {
	Subscribe(ctx context.Context, userID TelegramUserID, account ton.Address) error
	GetSubscriptions(ctx context.Context) ([]Subscription, error)
	Unsubscribe(ctx context.Context, userID TelegramUserID) error
}
