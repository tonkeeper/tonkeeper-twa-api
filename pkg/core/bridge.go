package core

import (
	"fmt"
	"sync"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type ClientID string

type Bridge struct {
	logger *zap.Logger

	messageCh chan<- telegram.Message

	mu              sync.RWMutex
	subsPerClientID map[ClientID]map[telegram.UserID]struct{}
	subsPerUserID   map[telegram.UserID]map[ClientID]struct{}
}

func NewBridge(logger *zap.Logger, messageCh chan<- telegram.Message) *Bridge {
	return &Bridge{
		logger:          logger,
		messageCh:       messageCh,
		subsPerUserID:   map[telegram.UserID]map[ClientID]struct{}{},
		subsPerClientID: map[ClientID]map[telegram.UserID]struct{}{},
	}
}

func (r *Bridge) Subscribe(userID telegram.UserID, clientID ClientID) error {
	// TODO: save to the database
	r.subscribe(userID, clientID)
	return nil
}

func (r *Bridge) subscribe(userID telegram.UserID, clientID ClientID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subsPerUserID[userID]; !ok {
		r.subsPerUserID[userID] = make(map[ClientID]struct{}, 1)
	}
	if _, ok := r.subsPerClientID[clientID]; !ok {
		r.subsPerClientID[clientID] = make(map[telegram.UserID]struct{}, 1)
	}
	r.subsPerUserID[userID][clientID] = struct{}{}
	r.subsPerClientID[clientID][userID] = struct{}{}
}

func (r *Bridge) unsubscribe(userID telegram.UserID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	clientIDs, ok := r.subsPerUserID[userID]
	if !ok {
		return
	}
	for clientID := range clientIDs {
		delete(r.subsPerClientID[clientID], userID)
	}
	delete(r.subsPerUserID, userID)
}

func (r *Bridge) subscriptions(clientID ClientID) []telegram.UserID {
	r.mu.RLock()
	defer r.mu.RUnlock()
	userIDs, ok := r.subsPerClientID[clientID]
	if !ok {
		return nil
	}
	return maps.Keys(userIDs)
}

func (r *Bridge) HandleWebhook(clientID ClientID, topic string, hash string) {
	userIDs := r.subscriptions(clientID)
	for _, userID := range userIDs {
		r.messageCh <- telegram.Message{
			UserID: userID,
			Text:   fmt.Sprintf("New event: %s", topic),
		}
	}
}

func (r *Bridge) Unsubscribe(userID telegram.UserID) error {
	// TODO: remove from the database
	r.unsubscribe(userID)
	return nil
}
