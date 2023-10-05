package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/r3labs/sse/v2"
	tonapiClient "github.com/tonkeeper/opentonapi/client"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

var (
	accountEventsSubscribers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "twa_api_account_event_subscribers",
		Help: "Number of account-events subscribers",
	})
)

type AccountEventsNotificator struct {
	logger    *zap.Logger
	storage   Storage
	tonapiKey string

	client *tonapiClient.Client

	mu               sync.RWMutex
	subsPerUserID    map[telegram.UserID]map[ton.AccountID]struct{}
	subsPerAccountID map[ton.AccountID]map[telegram.UserID]struct{}
}

func NewNotificator(logger *zap.Logger, storage Storage, tonapiKey string) (*AccountEventsNotificator, error) {
	if len(tonapiKey) > 0 {
		req, err := http.NewRequest(http.MethodGet, "https://tonapi.io/v2/me", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tonapiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("invalid tonapi key")
		}
		logger.Info("tonapi.io key is valid")
	} else {
		logger.Warn("tonapi.io key is not set")
	}
	cli, err := tonapiClient.NewClient("https://tonapi.io", tonapiClient.WithTonApiKey(tonapiKey))
	if err != nil {
		return nil, err
	}
	subscriptions, err := storage.GetAccountEventsSubscriptions(context.TODO())
	if err != nil {
		return nil, err
	}
	subsPerUserID := make(map[telegram.UserID]map[ton.AccountID]struct{})
	subsPerAccountID := make(map[ton.AccountID]map[telegram.UserID]struct{})

	for _, sub := range subscriptions {
		if _, ok := subsPerUserID[sub.TelegramUserID]; !ok {
			subsPerUserID[sub.TelegramUserID] = make(map[ton.AccountID]struct{})
		}
		if _, ok := subsPerAccountID[sub.Account]; !ok {
			subsPerAccountID[sub.Account] = make(map[telegram.UserID]struct{})
		}
		subsPerUserID[sub.TelegramUserID][sub.Account] = struct{}{}
		subsPerAccountID[sub.Account][sub.TelegramUserID] = struct{}{}
	}

	accountEventsSubscribers.Set(float64(len(subsPerUserID)))

	return &AccountEventsNotificator{
		tonapiKey:        tonapiKey,
		logger:           logger,
		client:           cli,
		storage:          storage,
		subsPerAccountID: subsPerAccountID,
		subsPerUserID:    subsPerUserID,
	}, nil
}

func (n *AccountEventsNotificator) Subscribe(userID telegram.UserID, account ton.Address) error {
	if err := n.storage.SubscribeToAccountEvents(context.TODO(), userID, account); err != nil {
		return err
	}
	n.subscribe(userID, account)
	n.updateMetrics()
	return nil
}

func (n *AccountEventsNotificator) subscribe(userID telegram.UserID, account ton.Address) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.subsPerAccountID[account.ID]; !ok {
		n.subsPerAccountID[account.ID] = make(map[telegram.UserID]struct{})
	}
	n.subsPerAccountID[account.ID][userID] = struct{}{}

	if _, ok := n.subsPerUserID[userID]; !ok {
		n.subsPerUserID[userID] = make(map[ton.AccountID]struct{})
	}
	n.subsPerUserID[userID][account.ID] = struct{}{}
}

type TraceEventData struct {
	AccountIDs []tongo.AccountID `json:"accounts"`
	Hash       string            `json:"hash"`
}

func (n *AccountEventsNotificator) IsSubscribed(userID telegram.UserID, account ton.AccountID) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	usersIDs, ok := n.subsPerAccountID[account]
	if !ok {
		return false
	}
	_, ok = usersIDs[userID]
	return ok
}

func (n *AccountEventsNotificator) subscribedAccounts(accountIDs []ton.AccountID) []ton.AccountID {
	n.mu.RLock()
	defer n.mu.RUnlock()
	var accounts []ton.AccountID
	for _, accountID := range accountIDs {
		if _, ok := n.subsPerAccountID[accountID]; ok {
			accounts = append(accounts, accountID)
		}
	}
	return accounts
}

func (n *AccountEventsNotificator) unsubscribe(userID telegram.UserID) {
	n.mu.Lock()
	defer n.mu.Unlock()
	subs, ok := n.subsPerUserID[userID]
	if !ok {
		return
	}
	for account := range subs {
		delete(n.subsPerAccountID[account], userID)
		if len(n.subsPerAccountID[account]) == 0 {
			delete(n.subsPerAccountID, account)
		}
	}
	delete(n.subsPerUserID, userID)
}

func (n *AccountEventsNotificator) accountSubscribers(account ton.AccountID) []telegram.UserID {
	n.mu.RLock()
	defer n.mu.RUnlock()

	subs, ok := n.subsPerAccountID[account]
	if !ok {
		return nil
	}
	return maps.Keys(subs)
}

func (n *AccountEventsNotificator) notify(accounts []ton.AccountID, hash string, messageCh chan<- telegram.Message) {
	rawAccounts := make([]string, 0, len(accounts))
	for _, account := range accounts {
		rawAccounts = append(rawAccounts, account.ToRaw())
	}

	n.logger.Info("notify",
		zap.String("hash", hash),
		zap.Strings("accounts", rawAccounts))

	for _, account := range accounts {
		subscribers := n.accountSubscribers(account)
		if len(subscribers) == 0 {
			continue
		}
		var event *tonapiClient.AccountEvent
		err := retry.Do(func() error {
			params := tonapiClient.GetAccountEventParams{
				AccountID: account.ToRaw(),
				EventID:   hash,
				SubjectOnly: tonapiClient.OptBool{
					Value: true,
					Set:   true,
				},
			}
			e, err := n.client.GetAccountEvent(context.TODO(), params)
			if err != nil {
				return err
			}
			event = e
			return nil
		}, retry.Attempts(3), retry.Delay(1*time.Second))
		if err != nil {
			n.logger.Error("GetAccountEvent() failed", zap.Error(err))
			continue
		}
		msgs := formatMessages(account, event)
		n.logger.Info("send-notification",
			zap.String("hash", hash),
			zap.Int("#messages", len(msgs)),
			zap.Int("#subscribers", len(subscribers)))
		for _, userID := range subscribers {
			for _, msg := range msgs {
				messageCh <- telegram.Message{
					UserID: userID,
					Text:   msg,
				}

			}
		}
	}
}

func (n *AccountEventsNotificator) sseSubscribe(ctx context.Context, messageCh chan<- telegram.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sseClient := sse.NewClient("https://tonapi.io/v2/sse/accounts/traces?accounts=ALL")
	if len(n.tonapiKey) > 0 {
		sseClient.Headers["Authorization"] = fmt.Sprintf("Bearer %s", n.tonapiKey)
	}
	return sseClient.SubscribeWithContext(ctx, "", func(msg *sse.Event) {
		switch string(msg.Event) {
		case "heartbeat":
			n.logger.Info("sse heartbeat")
			return

		case "message":
			data := TraceEventData{}

			n.logger.Info("trace event",
				zap.String("event-id", string(msg.ID)))

			if err := json.Unmarshal(msg.Data, &data); err != nil {
				n.logger.Error("json.Unmarshal() failed",
					zap.Error(err),
					zap.String("data", string(msg.Data)))
				return
			}
			if accounts := n.subscribedAccounts(data.AccountIDs); len(accounts) > 0 {
				go n.notify(accounts, data.Hash, messageCh)
			}
		}
	})
}

func (n *AccountEventsNotificator) Run(ctx context.Context, messageCh chan<- telegram.Message) {
	for {
		if err := n.sseSubscribe(ctx, messageCh); err != nil {
			n.logger.Error("sseClient.Subscribe() failed", zap.Error(err))
			time.Sleep(10 * time.Second)
		}
	}
}

func (n *AccountEventsNotificator) Unsubscribe(userID telegram.UserID) error {
	if err := n.storage.UnsubscribeAccountEvents(context.TODO(), userID); err != nil {
		return err
	}
	n.unsubscribe(userID)
	n.updateMetrics()
	return nil
}

func (n *AccountEventsNotificator) updateMetrics() {
	n.mu.RLock()
	defer n.mu.RUnlock()
	accountEventsSubscribers.Set(float64(len(n.subsPerUserID)))
}
