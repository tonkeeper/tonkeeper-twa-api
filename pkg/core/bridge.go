package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
	"go.uber.org/zap"
)

type ClientID string

type bridgeSubscription struct {
	Origin string
	UserID telegram.UserID
}

type Bridge struct {
	logger *zap.Logger

	storage   Storage
	messageCh chan<- telegram.Message

	mu               sync.RWMutex
	subsPerClientID  map[ClientID]bridgeSubscription
	clientIDsPerUser map[telegram.UserID]map[ClientID]struct{}
}

func NewBridge(logger *zap.Logger, storage Storage, messageCh chan<- telegram.Message) *Bridge {
	return &Bridge{
		logger:           logger,
		storage:          storage,
		messageCh:        messageCh,
		subsPerClientID:  map[ClientID]bridgeSubscription{},
		clientIDsPerUser: map[telegram.UserID]map[ClientID]struct{}{},
	}
}

// HandleWebhook is called by the HTTP Bridge when it receives a new event.
func (b *Bridge) HandleWebhook(clientID ClientID, topic string, hash string) {
	subscription, ok := b.subscription(clientID)
	if !ok {
		return
	}
	b.messageCh <- telegram.Message{
		UserID: subscription.UserID,
		Text:   fmt.Sprintf("New event: %s from %v", topic, subscription.Origin),
	}
}

func (b *Bridge) Subscribe(userID telegram.UserID, clientID ClientID, origin string) error {
	if err := b.storage.SubscribeToBridgeEvents(context.TODO(), userID, clientID, origin); err != nil {
		return err
	}
	b.subscribe(userID, clientID, origin)
	return nil
}

func (b *Bridge) Unsubscribe(userID telegram.UserID, clientID *ClientID) error {
	if err := b.storage.UnsubscribeFromBridgeEvents(context.TODO(), userID, clientID); err != nil {
		return err
	}
	if clientID == nil {
		b.cancelUserSubscriptions(userID)
	} else {
		b.cancelSpecificSubscription(userID, *clientID)
	}
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