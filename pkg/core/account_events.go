package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	ht "github.com/ogen-go/ogen/http"
	"github.com/r3labs/sse/v2"
	tonapiClient "github.com/tonkeeper/opentonapi/client"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type AccountEventsNotificator struct {
	logger  *zap.Logger
	storage Storage

	client *tonapiClient.Client

	mu               sync.RWMutex
	subsPerUserID    map[telegram.UserID]map[ton.AccountID]struct{}
	subsPerAccountID map[ton.AccountID]map[telegram.UserID]struct{}
	events           map[string]struct{}
}

type client struct {
	tonapiKey string
}

func (c client) Do(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tonapiKey))
	return http.DefaultClient.Do(r)
}

var _ ht.Client = &client{}

func NewNotificator(logger *zap.Logger, storage Storage, tonapiKey string) (*AccountEventsNotificator, error) {
	cli, err := tonapiClient.NewClient("https://tonapi.io", tonapiClient.WithClient(client{tonapiKey: tonapiKey}))
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
	return &AccountEventsNotificator{
		logger:           logger,
		client:           cli,
		storage:          storage,
		subsPerAccountID: subsPerAccountID,
		subsPerUserID:    subsPerUserID,
		events:           make(map[string]struct{}),
	}, nil
}

func (n *AccountEventsNotificator) Subscribe(userID telegram.UserID, account ton.Address) error {
	if err := n.storage.SubscribeToAccountEvents(context.TODO(), userID, account); err != nil {
		return err
	}
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

	return nil
}

// TransactionEventData represents the data part of a new-transaction event.
type TransactionEventData struct {
	AccountID tongo.AccountID `json:"account_id"`
	Lt        uint64          `json:"lt"`
	TxHash    string          `json:"tx_hash"`
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

func (n *AccountEventsNotificator) isSubscribed(account ton.AccountID) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	_, ok := n.subsPerAccountID[account]
	return ok
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

func (n *AccountEventsNotificator) startEventProcessing(eventID string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.events[eventID]; ok {
		return false
	}
	n.events[eventID] = struct{}{}
	return true
}

func (n *AccountEventsNotificator) stopEventProcessing(eventID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.events, eventID)
}

func (n *AccountEventsNotificator) notify(hash string, messageCh chan<- telegram.Message) {
	event, err := n.client.GetEvent(context.TODO(), tonapiClient.GetEventParams{EventID: hash})
	if err != nil {
		panic(err)
	}
	if !n.startEventProcessing(event.EventID) {
		return
	}
	eventID := event.EventID
	defer n.stopEventProcessing(eventID)

	for event.InProgress {
		time.Sleep(10 * time.Second)
		event, err = n.client.GetEvent(context.TODO(), tonapiClient.GetEventParams{EventID: hash})
		if err != nil {
			panic(err)
		}
	}
	for _, action := range event.Actions {
		for _, account := range action.SimplePreview.Accounts {
			addr, err := tongo.ParseAddress(account.Address)
			if err != nil {
				panic(err)
			}
			subscribers := n.accountSubscribers(addr.ID)
			for _, userID := range subscribers {
				messageCh <- telegram.Message{
					UserID: userID,
					Text:   fmt.Sprintf("%v", action.SimplePreview.Description),
				}
			}
		}
	}
}

func (n *AccountEventsNotificator) Run(ctx context.Context, messageCh chan<- telegram.Message) {
	for {
		sseClient := sse.NewClient("https://tonapi.io/v2/sse/accounts/transactions?accounts=ALL")
		err := sseClient.SubscribeWithContext(ctx, "", func(msg *sse.Event) {
			switch string(msg.Event) {
			case "heartbeat":
				return
			case "message":
				data := TransactionEventData{}
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					n.logger.Error("json.Unmarshal() failed",
						zap.Error(err),
						zap.String("data", string(msg.Data)))
					return
				}
				if n.isSubscribed(data.AccountID) {
					fmt.Printf("accountID: %v, lt: %v, tx hash: %x\n", data.AccountID.ToRaw(), data.Lt, data.TxHash)
					go n.notify(data.TxHash, messageCh)
				}
			}
		})
		if err != nil {
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
	return nil
}
