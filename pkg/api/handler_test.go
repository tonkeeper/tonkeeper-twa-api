package api

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api/oas"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
	"go.uber.org/zap"
)

type MockStorage struct {
}

func (m *MockStorage) SubscribeToAccountEvents(ctx context.Context, userID telegram.UserID, account ton.Address) error {
	return nil
}

func (m *MockStorage) GetAccountEventsSubscriptions(ctx context.Context) ([]core.AccountEventsSubscription, error) {
	return nil, nil
}

func (m *MockStorage) UnsubscribeAccountEvents(ctx context.Context, userID telegram.UserID) error {
	return nil
}

func (m *MockStorage) SubscribeToBridgeEvents(ctx context.Context, userID telegram.UserID, clientID core.ClientID, origin string) error {
	return nil
}

func (m *MockStorage) UnsubscribeFromBridgeEvents(ctx context.Context, userID telegram.UserID, clientID *core.ClientID) error {
	return nil
}

func (m *MockStorage) GetBridgeSubscriptions(ctx context.Context) ([]core.BridgeSubscription, error) {
	return nil, nil
}

var _ core.Storage = (*MockStorage)(nil)

func TestHandler_AccountEventsSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name           string
		request        *oas.AccountEventsSubscriptionStatusReq
		wantSubscribed bool
		wantErr        string
	}{
		{
			name: "subscribed = true",
			request: &oas.AccountEventsSubscriptionStatusReq{
				TwaInitData: "1",
				Address:     "0:dd61300e0060f80233363b3b4a0f3b27ad03b19cc4bec6ec798aab0b3e479eba",
			},
			wantSubscribed: true,
		},
		{
			name: "subscribed = false",
			request: &oas.AccountEventsSubscriptionStatusReq{
				TwaInitData: "2",
				Address:     "0:dd61300e0060f80233363b3b4a0f3b27ad03b19cc4bec6ec798aab0b3e479eba",
			},
			wantSubscribed: false,
		},
		{
			name: "bad init data - error",
			request: &oas.AccountEventsSubscriptionStatusReq{
				TwaInitData: "broken_data",
				Address:     "0:dd61300e0060f80233363b3b4a0f3b27ad03b19cc4bec6ec798aab0b3e479eba",
			},
			wantSubscribed: false,
			wantErr:        `code 400: {Error:twa init data err}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MockStorage{}
			notificator, err := core.NewNotificator(zap.L(), s, "")
			require.Nil(t, err)
			addr, err := tongo.ParseAddress("0:dd61300e0060f80233363b3b4a0f3b27ad03b19cc4bec6ec798aab0b3e479eba")
			require.Nil(t, err)
			err = notificator.Subscribe(1, addr)
			require.Nil(t, err)

			h := &Handler{
				logger: zap.L(),
				extractUserFn: func(data string, telegramSecret string) (telegram.UserID, error) {
					require.Equal(t, "secret", telegramSecret)
					value, err := strconv.Atoi(data)
					if err != nil {
						return 0, errors.New("twa init data err")
					}
					return telegram.UserID(value), nil
				},
				telegramSecret: "secret",
				notificator:    notificator,
			}

			status, err := h.AccountEventsSubscriptionStatus(context.Background(), tt.request)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.Nil(t, err)
			require.Equal(t, tt.wantSubscribed, status.Subscribed)
		})
	}
}
