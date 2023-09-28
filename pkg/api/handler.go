package api

import (
	"context"
	"fmt"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tonconnect"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/telegram"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api/oas"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
)

type Handler struct {
	logger *zap.Logger

	telegramSecret string
	tonConnect     *tonconnect.Server
	notificator    *core.AccountEventsNotificator
	bridge         *core.Bridge
}

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

func (h *Handler) GetTonConnectPayload(ctx context.Context) (*oas.GetTonConnectPayloadOK, error) {
	payload, err := h.tonConnect.GeneratePayload()
	if err != nil {
		return nil, InternalError(err)
	}
	return &oas.GetTonConnectPayloadOK{Payload: payload}, nil
}

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
		return BadRequest(err.Error())
	}
	if !verified {
		return BadRequest("failed to verify proof")
	}
	userID, err := telegram.ExtractUserIDFromInitData(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.notificator.Subscribe(userID, account); err != nil {
		return InternalError(err)
	}
	return nil
}

func (h *Handler) UnsubscribeFromAccountEvents(ctx context.Context, req *oas.UnsubscribeFromAccountEventsReq) error {
	userID, err := telegram.ExtractUserIDFromInitData(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.notificator.Unsubscribe(userID); err != nil {
		return InternalError(err)
	}
	return nil
}

func (h *Handler) SubscribeToBridgeEvents(ctx context.Context, req *oas.SubscribeToBridgeEventsReq) error {
	userID, err := telegram.ExtractUserIDFromInitData(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.bridge.Subscribe(userID, core.ClientID(req.ClientID), req.Origin); err != nil {
		return InternalError(err)
	}
	return nil
}

func (h *Handler) BridgeWebhook(ctx context.Context, req *oas.BridgeWebhookReq, params oas.BridgeWebhookParams) error {
	h.bridge.HandleWebhook(core.ClientID(params.ClientID), req.Topic, req.Hash)
	return nil
}

func (h *Handler) UnsubscribeFromBridgeEvents(ctx context.Context, req *oas.UnsubscribeFromBridgeEventsReq) error {
	userID, err := telegram.ExtractUserIDFromInitData(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.bridge.Unsubscribe(userID); err != nil {
		return InternalError(err)
	}
	return nil
}
