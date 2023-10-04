package storage

import (
	"context"
	"sort"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

func initDatabase(pool *pgxpool.Pool, t *testing.T) {
	_, err := pool.Exec(context.Background(), "INSERT INTO twa.bridge_subscriptions (telegram_user_id, client_id, origin) VALUES (1, '1000', 'dns.ton.org')")
	require.Nil(t, err)
	_, err = pool.Exec(context.Background(), "INSERT INTO twa.bridge_subscriptions (telegram_user_id, client_id, origin) VALUES (1, '1001', 'ton.org')")
	require.Nil(t, err)
	_, err = pool.Exec(context.Background(), "INSERT INTO twa.bridge_subscriptions (telegram_user_id, client_id, origin) VALUES (2, '2002', 'dns.ton.org')")
	require.Nil(t, err)

	_, err = pool.Exec(context.Background(), "INSERT INTO twa.subscriptions (telegram_user_id, account) VALUES (1, '0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1000')")
	require.Nil(t, err)
	_, err = pool.Exec(context.Background(), "INSERT INTO twa.subscriptions (telegram_user_id, account) VALUES (1, '0:6ccd325a858c379693fae2bcaab1c2906831a4e10a6c3bb44ee8b615bca1d220')")
	require.Nil(t, err)
}

type bridgeSubscriptionData struct {
	TelegramUserID telegram.UserID
	ClientID       core.ClientID
	Origin         string
}

func Test_storage_SubscribeToBridgeEvents(t *testing.T) {
	tests := []struct {
		name             string
		userID           telegram.UserID
		clientID         core.ClientID
		origin           string
		maxSubscriptions int
		wantData         []bridgeSubscriptionData
		wantErr          string
	}{
		{
			name:             "all good",
			userID:           1,
			clientID:         "1002",
			origin:           "dex.ton",
			maxSubscriptions: maxBridgeSubscriptionsPerUser,
			wantData: []bridgeSubscriptionData{
				{TelegramUserID: 1, ClientID: "1000", Origin: "dns.ton.org"},
				{TelegramUserID: 1, ClientID: "1001", Origin: "ton.org"},
				{TelegramUserID: 1, ClientID: "1002", Origin: "dex.ton"},
				{TelegramUserID: 2, ClientID: "2002", Origin: "dns.ton.org"},
			},
		},
		{
			name:             "userID and origin are already in db - update client_id",
			userID:           1,
			clientID:         "1005",
			origin:           "dns.ton.org",
			maxSubscriptions: maxBridgeSubscriptionsPerUser,
			wantData: []bridgeSubscriptionData{
				{TelegramUserID: 1, ClientID: "1001", Origin: "ton.org"},
				{TelegramUserID: 1, ClientID: "1005", Origin: "dns.ton.org"},
				{TelegramUserID: 2, ClientID: "2002", Origin: "dns.ton.org"},
			},
		},
		{
			name:             "another subscription with the same client_id - previous subscription should be removed",
			userID:           2,
			clientID:         "2002",
			origin:           "ton.org",
			maxSubscriptions: maxBridgeSubscriptionsPerUser,
			wantData: []bridgeSubscriptionData{
				{TelegramUserID: 1, ClientID: "1000", Origin: "dns.ton.org"},
				{TelegramUserID: 1, ClientID: "1001", Origin: "ton.org"},
				{TelegramUserID: 2, ClientID: "2002", Origin: "ton.org"},
			},
		},
		{
			name:             "max subscriptions per user reached",
			userID:           1,
			clientID:         "1002",
			origin:           "dex.org",
			maxSubscriptions: 2,
			wantErr:          `max subscriptions per user reached`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := createDB(t)
			initDatabase(pool, t)
			s := &storage{
				logger:                        zap.L(),
				pool:                          pool,
				maxBridgeSubscriptionsPerUser: tt.maxSubscriptions,
			}
			err := s.SubscribeToBridgeEvents(context.Background(), tt.userID, tt.clientID, tt.origin)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.Nil(t, err)

			rows, err := pool.Query(context.Background(), "SELECT telegram_user_id, client_id, origin FROM twa.bridge_subscriptions")
			require.Nil(t, err)
			defer rows.Close()

			data := make([]bridgeSubscriptionData, 0)
			for rows.Next() {
				var sub bridgeSubscriptionData
				if err := rows.Scan(&sub.TelegramUserID, &sub.ClientID, &sub.Origin); err != nil {
					require.Nil(t, err)
				}
				data = append(data, sub)
			}
			sort.Slice(data, func(i, j int) bool {
				return data[i].ClientID < data[j].ClientID
			})
			require.Equal(t, tt.wantData, data)
		})
	}
}

func clientIDPtr(clientID core.ClientID) *core.ClientID {
	return &clientID
}

func Test_storage_UnsubscribeFromBridgeEvents(t *testing.T) {

	tests := []struct {
		name     string
		userID   telegram.UserID
		clientID *core.ClientID
		wantData []bridgeSubscriptionData
	}{
		{
			name:   "remove all subscriptions for user",
			userID: 1,
			wantData: []bridgeSubscriptionData{
				{TelegramUserID: 2, ClientID: "2002", Origin: "dns.ton.org"},
			},
		},
		{
			name:     "remove specific subscription for user",
			userID:   1,
			clientID: clientIDPtr("1000"),
			wantData: []bridgeSubscriptionData{
				{TelegramUserID: 1, ClientID: "1001", Origin: "ton.org"},
				{TelegramUserID: 2, ClientID: "2002", Origin: "dns.ton.org"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := createDB(t)
			initDatabase(pool, t)
			s := &storage{logger: zap.L(), pool: pool}
			err := s.UnsubscribeFromBridgeEvents(context.Background(), tt.userID, tt.clientID)
			require.Nil(t, err)
		})
	}
}

func Test_storage_SubscribeToAccountEvents(t *testing.T) {
	tests := []struct {
		name              string
		userID            telegram.UserID
		addr              ton.Address
		maxWallets        int
		wantSubscriptions []core.AccountEventsSubscription
		wantErr           string
	}{
		{
			name:       "all good",
			userID:     2,
			addr:       ton.Address{ID: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1999")},
			maxWallets: 10,
			wantSubscriptions: []core.AccountEventsSubscription{
				{TelegramUserID: 1, Account: ton.MustParseAccountID("0:6ccd325a858c379693fae2bcaab1c2906831a4e10a6c3bb44ee8b615bca1d220")},
				{TelegramUserID: 1, Account: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1000")},
				{TelegramUserID: 2, Account: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1999")},
			},
		},
		{
			name:       "add already existing subscription",
			userID:     1,
			addr:       ton.Address{ID: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1000")},
			maxWallets: 10,
			wantSubscriptions: []core.AccountEventsSubscription{
				{TelegramUserID: 1, Account: ton.MustParseAccountID("0:6ccd325a858c379693fae2bcaab1c2906831a4e10a6c3bb44ee8b615bca1d220")},
				{TelegramUserID: 1, Account: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1000")},
			},
		},
		{
			name:       "max wallets per user reached",
			userID:     1,
			addr:       ton.Address{ID: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1999")},
			maxWallets: 2,
			wantErr:    `max wallets per user reached`,
			wantSubscriptions: []core.AccountEventsSubscription{
				{TelegramUserID: 1, Account: ton.MustParseAccountID("0:6ccd325a858c379693fae2bcaab1c2906831a4e10a6c3bb44ee8b615bca1d220")},
				{TelegramUserID: 1, Account: ton.MustParseAccountID("0:bdf3fa8098d129b54b4f73b5bac5d1e1fd91eb054169c3916dfc8ccd536d1000")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := createDB(t)
			initDatabase(pool, t)
			s := &storage{logger: zap.L(), pool: pool, maxWalletsPerUser: tt.maxWallets}
			err := s.SubscribeToAccountEvents(context.Background(), tt.userID, tt.addr)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.Nil(t, err)
			}

			subs, err := s.GetAccountEventsSubscriptions(context.Background())
			require.Nil(t, err)

			sort.Slice(subs, func(i, j int) bool {
				if subs[i].TelegramUserID != subs[j].TelegramUserID {
					return subs[i].TelegramUserID < subs[j].TelegramUserID
				}
				return subs[i].Account.String() < subs[j].Account.String()
			})

			require.Equal(t, tt.wantSubscriptions, subs)

		})
	}
}
