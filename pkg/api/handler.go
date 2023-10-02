package api

import (
	"context"
	"fmt"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tonconnect"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api/oas"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
)

// Handler handles operations described by OpenAPI v3 specification of this service.
// It implements oas.Handler interface and every API operation is implemented as a method on Handler.
type Handler struct {
	logger *zap.Logger

	telegramSecret string
	tonConnect     *tonconnect.Server
	notificator    *core.AccountEventsNotificator
	bridge         *core.Bridge

	// extractUserFn is an indirection for testing.
	extractUserFn extractUserFromTwaInitDataFn
}

// extractUserFromTwaInitDataFn extracts a telegram user ID from TWA init data.
//
// For more details see
// https://docs.twa.dev/docs/launch-params/init-data#authorization-and-authentication
type extractUserFromTwaInitDataFn func(data string, telegramSecret string) (telegram.UserID, error)

type Config struct {
	TonConnectSecret  string
	TelegramBotSecret string
}

var _ oas.Handler = (*Handler)(nil)

func NewHandler(logger *zap.Logger, notificator *core.AccountEventsNotificator, bridge *core.Bridge, config Config) (*Handler, error) {
	cli, err := liteapi.NewClient(liteapi.Mainnet(), liteapi.FromEnvs())
	if err != nil {
		return nil, err
	}
	tonConnect, err := tonconnect.NewTonConnect(cli, config.TonConnectSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init tonconnect: %w", err)
	}
	return &Handler{
		logger:         logger,
		bridge:         bridge,
		tonConnect:     tonConnect,
		notificator:    notificator,
		telegramSecret: config.TelegramBotSecret,
		extractUserFn:  telegram.ExtractUserIDFromInitData,
	}, nil
}

func (h *Handler) NewError(ctx context.Context, err error) *oas.ErrorStatusCode {
	switch x := err.(type) {
	case *oas.ErrorStatusCode:
		return x
	default:
		return InternalError(x)
	}
}

// GetTonConnectPayload returns a challenge for TON Connect.
func (h *Handler) GetTonConnectPayload(ctx context.Context) (*oas.GetTonConnectPayloadOK, error) {
	payload, err := h.tonConnect.GeneratePayload()
	if err != nil {
		return nil, InternalError(err)
	}
	return &oas.GetTonConnectPayloadOK{Payload: payload}, nil
}

// SubscribeToAccountEvents subscribes to notifications about events in the TON blockchain for a specific address.
func (h *Handler) SubscribeToAccountEvents(ctx context.Context, req *oas.SubscribeToAccountEventsReq) error {
	proof := tonconnect.Proof{
		Address: req.Address,
		Proof: tonconnect.ProofData{
			Timestamp: req.Proof.Timestamp,
			Domain:    req.Proof.Domain.Value,
			Signature: req.Proof.Signature,
			Payload:   req.Proof.Payload,
			StateInit: req.Proof.StateInit.Value,
		},
	}
	account, err := tongo.ParseAddress(req.Address)
	if err != nil {
		return BadRequest(err.Error())
	}
	verified, _, err := h.tonConnect.CheckProof(ctx, &proof)
	if err != nil {
		return BadRequest(fmt.Sprintf("failed to check proof: %v", err))
	}
	if !verified {
		return BadRequest("failed to verify proof")
	}
	userID, err := h.extractUserFn(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.notificator.Subscribe(userID, account); err != nil {
		return InternalError(err)
	}
	return nil
}

// AccountEventsSubscriptionStatus returns a status of an account-events subscription.
func (h *Handler) AccountEventsSubscriptionStatus(ctx context.Context, req *oas.AccountEventsSubscriptionStatusReq) (*oas.AccountEventsSubscriptionStatusOK, error) {
	userID, err := h.extractUserFn(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return nil, BadRequest(err.Error())
	}
	accountID, err := tongo.ParseAccountID(req.Address)
	if err != nil {
		return nil, BadRequest(err.Error())
	}
	subscribed := h.notificator.IsSubscribed(userID, accountID)
	return &oas.AccountEventsSubscriptionStatusOK{Subscribed: subscribed}, nil
}

// UnsubscribeFromAccountEvents unsubscribes from notifications about events in the TON blockchain for a specific address.
func (h *Handler) UnsubscribeFromAccountEvents(ctx context.Context, req *oas.UnsubscribeFromAccountEventsReq) error {
	userID, err := h.extractUserFn(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.notificator.Unsubscribe(userID); err != nil {
		return InternalError(err)
	}
	return nil
}

// SubscribeToBridgeEvents subscribes to notifications from the HTTP Bridge regarding a specific smart contract or wallet.
func (h *Handler) SubscribeToBridgeEvents(ctx context.Context, req *oas.SubscribeToBridgeEventsReq) error {
	userID, err := h.extractUserFn(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.bridge.Subscribe(userID, core.ClientID(req.ClientID), req.Origin); err != nil {
		return InternalError(err)
	}
	return nil
}

// BridgeWebhook is called by the HTTP Bridge when an event occurs.
func (h *Handler) BridgeWebhook(ctx context.Context, req *oas.BridgeWebhookReq, params oas.BridgeWebhookParams) error {
	h.bridge.HandleWebhook(core.ClientID(params.ClientID), req.Topic)
	return nil
}

// UnsubscribeFromBridgeEvents unsubscribes from bridge notifications.
func (h *Handler) UnsubscribeFromBridgeEvents(ctx context.Context, req *oas.UnsubscribeFromBridgeEventsReq) error {
	userID, err := h.extractUserFn(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	var clientID *core.ClientID
	if req.ClientID.IsSet() {
		id := core.ClientID(req.ClientID.Value)
		clientID = &id
	}
	if err := h.bridge.Unsubscribe(userID, clientID); err != nil {
		return InternalError(err)
	}
	return nil
}
