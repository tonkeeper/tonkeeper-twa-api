package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	initData "github.com/Telegram-Web-Apps/init-data-golang"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tonconnect"
	"go.uber.org/zap"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/api/oas"
	"github.com/tonkeeper/tonkeeper-twa-api/pkg/core"
)

type Handler struct {
	logger *zap.Logger

	telegramSecret string
	tonConnect     *tonconnect.Server
	notificator    *core.Notificator
}

type Config struct {
	TonConnectSecret  string
	TelegramBotSecret string
}

var _ oas.Handler = (*Handler)(nil)

func NewHandler(logger *zap.Logger, notificator *core.Notificator, config Config) (*Handler, error) {
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

func extractUserIDFromInitData(data string, telegramSecret string) (core.TelegramUserID, error) {
	// TODO: use right duration
	twaInitData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return 0, fmt.Errorf("failed to decode init data")
	}
	if err := initData.Validate(string(twaInitData), telegramSecret, time.Duration(0)); err != nil {
		return 0, fmt.Errorf("failed to validate init data")
	}
	parsedData, err := initData.Parse(string(twaInitData))
	if err != nil {
		return 0, fmt.Errorf("failed to parse init data")
	}
	if parsedData.User == nil {
		return 0, fmt.Errorf("user not found in init data")
	}
	return core.TelegramUserID(parsedData.User.ID), nil
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
	userID, err := extractUserIDFromInitData(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.notificator.Subscribe(userID, account); err != nil {
		return InternalError(err)
	}
	return nil
}

func (h *Handler) UnsubscribeFromAccountEvents(ctx context.Context, req *oas.UnsubscribeFromAccountEventsReq) error {
	userID, err := extractUserIDFromInitData(req.TwaInitData, h.telegramSecret)
	if err != nil {
		return BadRequest(err.Error())
	}
	if err := h.notificator.Unsubscribe(userID); err != nil {
		return InternalError(err)
	}
	return nil
}
