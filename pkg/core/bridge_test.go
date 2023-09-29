package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

func TestBridge_Subscribe(t *testing.T) {
	tests := []struct {
		name                 string
		userID               telegram.UserID
		clientID             ClientID
		origin               string
		wantSubsPerClientID  map[ClientID]bridgeSubscription
		wantClientIDsPerUser map[telegram.UserID]map[ClientID]struct{}
	}{
		{
			name:     "all good",
			userID:   3,
			clientID: "3003",
			origin:   "cex.com",
			wantClientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
				1: {"1000": {}, "1001": {}, "1002": {}},
				2: {"2002": {}},
				3: {"3003": {}},
			},
			wantSubsPerClientID: map[ClientID]bridgeSubscription{
				"1001": {Origin: "ton.org", UserID: 1},
				"1002": {Origin: "dex.ton", UserID: 1},
				"2002": {Origin: "dns.ton.org", UserID: 2},
				"3003": {Origin: "cex.com", UserID: 3},
			},
		},
		{
			name:     "overwriting existing subscription",
			userID:   2,
			clientID: "2002",
			origin:   "cex.com",
			wantClientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
				1: {"1000": {}, "1001": {}, "1002": {}},
				2: {"2002": {}},
			},
			wantSubsPerClientID: map[ClientID]bridgeSubscription{
				"1001": {Origin: "ton.org", UserID: 1},
				"1002": {Origin: "dex.ton", UserID: 1},
				"2002": {Origin: "cex.com", UserID: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mockStorage{
				OnSubscribeToBridgeEvents: func(ctx context.Context, userID telegram.UserID, clientID ClientID, origin string) error {
					require.Equal(t, tt.userID, userID)
					require.Equal(t, tt.clientID, clientID)
					require.Equal(t, tt.origin, origin)
					return nil
				},
			}
			b := &Bridge{
				logger:  zap.L(),
				storage: s,
				subsPerClientID: map[ClientID]bridgeSubscription{
					"1001": {Origin: "ton.org", UserID: 1},
					"1002": {Origin: "dex.ton", UserID: 1},
					"2002": {Origin: "dns.ton.org", UserID: 2},
				},
				clientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
					1: {"1000": {}, "1001": {}, "1002": {}},
					2: {"2002": {}},
				},
			}
			err := b.Subscribe(tt.userID, tt.clientID, tt.origin)
			require.Nil(t, err)
			require.Equal(t, tt.wantSubsPerClientID, b.subsPerClientID)
			require.Equal(t, tt.wantClientIDsPerUser, b.clientIDsPerUser)
		})
	}
}

func clientIDPtr(clientID ClientID) *ClientID {
	return &clientID
}

func TestBridge_Unsubscribe(t *testing.T) {
	tests := []struct {
		name                 string
		userID               telegram.UserID
		clientID             *ClientID
		wantSubsPerClientID  map[ClientID]bridgeSubscription
		wantClientIDsPerUser map[telegram.UserID]map[ClientID]struct{}
	}{
		{
			name:     "remove single client_id",
			userID:   1,
			clientID: clientIDPtr("1001"),
			wantClientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
				1: {"1000": {}, "1002": {}},
				2: {"2002": {}},
			},
			wantSubsPerClientID: map[ClientID]bridgeSubscription{
				"1002": {Origin: "dex.ton", UserID: 1},
				"2002": {Origin: "dns.ton.org", UserID: 2},
			},
		},
		{
			name:     "remove all",
			userID:   1,
			clientID: nil,
			wantClientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
				2: {"2002": {}},
			},
			wantSubsPerClientID: map[ClientID]bridgeSubscription{
				"2002": {Origin: "dns.ton.org", UserID: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &mockStorage{
				OnUnsubscribeFromBridgeEvents: func(ctx context.Context, userID telegram.UserID, clientID *ClientID) error {
					require.Equal(t, tt.userID, userID)
					require.Equal(t, tt.clientID, clientID)
					return nil
				},
			}
			b := &Bridge{
				logger:  zap.L(),
				storage: s,
				subsPerClientID: map[ClientID]bridgeSubscription{
					"1001": {Origin: "ton.org", UserID: 1},
					"1002": {Origin: "dex.ton", UserID: 1},
					"2002": {Origin: "dns.ton.org", UserID: 2},
				},
				clientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
					1: {"1000": {}, "1001": {}, "1002": {}},
					2: {"2002": {}},
				},
			}
			err := b.Unsubscribe(tt.userID, tt.clientID)
			require.Nil(t, err)
		})
	}
}

func TestBridge_HandleWebhook(t *testing.T) {
	tests := []struct {
		name     string
		clientID ClientID
		topic    string
		wantMsgs []telegram.Message
	}{
		{
			name:     "sendTransaction",
			clientID: "1001",
			topic:    "sendTransaction",
			wantMsgs: []telegram.Message{{UserID: 1, Text: "Transaction for ton.org"}},
		},
		{
			name:     "signData",
			clientID: "2002",
			topic:    "signData",
			wantMsgs: []telegram.Message{{UserID: 2, Text: "Data signature request dns.ton.org"}},
		},
		{
			name:     "no client_id -> no message",
			clientID: "3005",
			topic:    "signData",
		},
		{
			name:     "unknown topic -> no message",
			clientID: "2002",
			topic:    "unknownTopic",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageCh := make(chan telegram.Message, 1)
			b := &Bridge{
				logger: zap.L(),
				subsPerClientID: map[ClientID]bridgeSubscription{
					"1001": {Origin: "ton.org", UserID: 1},
					"1002": {Origin: "dex.ton", UserID: 1},
					"2002": {Origin: "dns.ton.org", UserID: 2},
				},
				clientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{
					1: {"1000": {}, "1001": {}, "1002": {}},
					2: {"2002": {}},
				},
				messageCh: messageCh,
			}
			b.HandleWebhook(tt.clientID, tt.topic)
			close(messageCh)

			var msgs []telegram.Message
			for msg := range messageCh {
				msgs = append(msgs, msg)
			}
			require.Equal(t, tt.wantMsgs, msgs)
		})
	}
}

func TestNewBridge(t *testing.T) {
	s := &mockStorage{
		OnGetBridgeSubscriptions: func(ctx context.Context) ([]BridgeSubscription, error) {
			subscriptions := []BridgeSubscription{
				{TelegramUserID: 2, ClientID: "2002", Origin: "dns.ton.org"},
				{TelegramUserID: 3, ClientID: "3000", Origin: "ton.org"},
				{TelegramUserID: 3, ClientID: "3001", Origin: "ton.org"},
				{TelegramUserID: 3, ClientID: "3002", Origin: "dex.ton"},
			}
			return subscriptions, nil
		},
	}
	bridge, err := NewBridge(zap.L(), s, nil)
	require.Nil(t, err)
	expectedSubsPerClientID := map[ClientID]bridgeSubscription{
		"2002": {Origin: "dns.ton.org", UserID: 2},
		"3000": {Origin: "ton.org", UserID: 3},
		"3001": {Origin: "ton.org", UserID: 3},
		"3002": {Origin: "dex.ton", UserID: 3},
	}
	require.Equal(t, expectedSubsPerClientID, bridge.subsPerClientID)
	expectedClientIDsPerUser := map[telegram.UserID]map[ClientID]struct{}{
		2: {"2002": {}},
		3: {"3000": {}, "3001": {}, "3002": {}},
	}
	require.Equal(t, expectedClientIDsPerUser, bridge.clientIDsPerUser)
}
