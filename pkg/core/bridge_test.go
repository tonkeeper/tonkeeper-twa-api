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
