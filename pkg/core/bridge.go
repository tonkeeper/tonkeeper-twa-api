package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

// ClientID is a unique identifier of a client.
// See more details at https://github.com/ton-blockchain/ton-connect/blob/main/bridge.md#http-bridge.
type ClientID string

type bridgeSubscription struct {
	Origin string
	UserID telegram.UserID
}

// Bridge receives notifications from the HTTP Bridge and sends them to telegram users.
type Bridge struct {
	logger *zap.Logger

	storage   Storage
	messageCh chan<- telegram.Message

	mu               sync.RWMutex
	subsPerClientID  map[ClientID]bridgeSubscription
	clientIDsPerUser map[telegram.UserID]map[ClientID]struct{}
}

var (
	bridgeSubscribers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "twa_api_bridge_subscribers",
		Help: "Number of bridge subscribers",
	})
)

func NewBridge(logger *zap.Logger, storage Storage, messageCh chan<- telegram.Message) (*Bridge, error) {
	subscriptions, err := storage.GetBridgeSubscriptions(context.TODO())
	if err != nil {
		return nil, err
	}
	subsPerClientID := make(map[ClientID]bridgeSubscription)
	clientIDsPerUser := make(map[telegram.UserID]map[ClientID]struct{})
	for _, sub := range subscriptions {
		if _, ok := clientIDsPerUser[sub.TelegramUserID]; !ok {
			clientIDsPerUser[sub.TelegramUserID] = make(map[ClientID]struct{})
		}
		clientIDsPerUser[sub.TelegramUserID][sub.ClientID] = struct{}{}
		subsPerClientID[sub.ClientID] = bridgeSubscription{
			Origin: sub.Origin,
			UserID: sub.TelegramUserID,
		}
	}
	bridgeSubscribers.Set(float64(len(clientIDsPerUser)))
	return &Bridge{
		logger:           logger,
		storage:          storage,
		messageCh:        messageCh,
		subsPerClientID:  subsPerClientID,
		clientIDsPerUser: clientIDsPerUser,
	}, nil
}

func formatMessage(topic string, origin string) (string, error) {
	switch topic {
	case "sendTransaction":
		return fmt.Sprintf("Transaction for %v", origin), nil
	case "signData":
		return fmt.Sprintf("Data signature request %v", origin), nil
	default:
		return "", fmt.Errorf("unknown topic")
	}
}

// HandleWebhook is called by the HTTP Bridge when it receives a new event.
func (b *Bridge) HandleWebhook(clientID ClientID, topic string) {
	subscription, ok := b.subscription(clientID)
	if !ok {
		return
	}
	msg, err := formatMessage(topic, subscription.Origin)
	if err != nil {
		b.logger.Error("failed to format message", zap.Error(err))
		return
	}
	b.messageCh <- telegram.Message{
		UserID: subscription.UserID,
		Text:   msg,
	}
}

// Subscribe subscribes a telegram user to the HTTP Bridge events.
func (b *Bridge) Subscribe(userID telegram.UserID, clientID ClientID, origin string) error {
	if err := b.storage.SubscribeToBridgeEvents(context.TODO(), userID, clientID, origin); err != nil {
		return err
	}
	b.subscribe(userID, clientID, origin)
	b.updateMetrics()
	return nil
}

// Unsubscribe unsubscribes a telegram user from the HTTP Bridge events.
func (b *Bridge) Unsubscribe(userID telegram.UserID, clientID *ClientID) error {
	if err := b.storage.UnsubscribeFromBridgeEvents(context.TODO(), userID, clientID); err != nil {
		return err
	}
	if clientID == nil {
		b.cancelUserSubscriptions(userID)
	} else {
		b.cancelSpecificSubscription(userID, *clientID)
	}
	b.updateMetrics()
	return nil
}

func (b *Bridge) subscribe(userID telegram.UserID, clientID ClientID, origin string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.clientIDsPerUser[userID]; !ok {
		b.clientIDsPerUser[userID] = make(map[ClientID]struct{}, 1)
	}
	b.clientIDsPerUser[userID][clientID] = struct{}{}

	b.subsPerClientID[clientID] = bridgeSubscription{
		Origin: origin,
		UserID: userID,
	}
}

func (b *Bridge) cancelUserSubscriptions(userID telegram.UserID) {
	b.mu.Lock()
	defer b.mu.Unlock()

	clientIDs, ok := b.clientIDsPerUser[userID]
	if !ok {
		return
	}
	for clientID := range clientIDs {
		delete(b.subsPerClientID, clientID)
	}
	delete(b.clientIDsPerUser, userID)
}

func (b *Bridge) cancelSpecificSubscription(userID telegram.UserID, clientID ClientID) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.subsPerClientID, clientID)

	if _, ok := b.clientIDsPerUser[userID]; !ok {
		return
	}
	delete(b.clientIDsPerUser[userID], clientID)
	if len(b.clientIDsPerUser[userID]) == 0 {
		delete(b.clientIDsPerUser, userID)
	}
}

func (b *Bridge) subscription(clientID ClientID) (bridgeSubscription, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	sub, ok := b.subsPerClientID[clientID]
	return sub, ok
}

func (b *Bridge) updateMetrics() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	bridgeSubscribers.Set(float64(len(b.clientIDsPerUser)))
}
