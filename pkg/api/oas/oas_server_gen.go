// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"context"
)

// Handler handles operations described by OpenAPI v3 specification.
type Handler interface {
	// AccountEventsSubscriptionStatus implements accountEventsSubscriptionStatus operation.
	//
	// Get a status of an account-events subscription.
	//
	// POST /account-events/subscription-status
	AccountEventsSubscriptionStatus(ctx context.Context, req *AccountEventsSubscriptionStatusReq) (*AccountEventsSubscriptionStatusOK, error)
	// BridgeWebhook implements bridgeWebhook operation.
	//
	// Webhook called by the HTTP Bridge when an event occurs.
	//
	// POST /bridge/webhook/{client_id}
	BridgeWebhook(ctx context.Context, req *BridgeWebhookReq, params BridgeWebhookParams) error
	// GetTonConnectPayload implements getTonConnectPayload operation.
	//
	// Get a challenge for TON Connect.
	//
	// GET /tonconnect/payload
	GetTonConnectPayload(ctx context.Context) (*GetTonConnectPayloadOK, error)
	// SubscribeToAccountEvents implements subscribeToAccountEvents operation.
	//
	// Subscribe to notifications about events in the TON blockchain for a specific address.
	//
	// POST /account-events/subscribe
	SubscribeToAccountEvents(ctx context.Context, req *SubscribeToAccountEventsReq) error
	// SubscribeToBridgeEvents implements subscribeToBridgeEvents operation.
	//
	// Subscribe to notifications from the HTTP Bridge regarding a specific smart contract or wallet.
	//
	// POST /bridge/subscribe
	SubscribeToBridgeEvents(ctx context.Context, req *SubscribeToBridgeEventsReq) error
	// UnsubscribeFromAccountEvents implements unsubscribeFromAccountEvents operation.
	//
	// Unsubscribe from notifications about events in the TON blockchain for a specific address.
	//
	// POST /account-events/unsubscribe
	UnsubscribeFromAccountEvents(ctx context.Context, req *UnsubscribeFromAccountEventsReq) error
	// UnsubscribeFromBridgeEvents implements unsubscribeFromBridgeEvents operation.
	//
	// Unsubscribe from bridge notifications.
	//
	// POST /bridge/unsubscribe
	UnsubscribeFromBridgeEvents(ctx context.Context, req *UnsubscribeFromBridgeEventsReq) error
	// NewError creates *ErrorStatusCode from error returned by handler.
	//
	// Used for common default response.
	NewError(ctx context.Context, err error) *ErrorStatusCode
}

// Server implements http server based on OpenAPI v3 specification and
// calls Handler to handle requests.
type Server struct {
	h Handler
	baseServer
}

// NewServer creates new Server.
func NewServer(h Handler, opts ...ServerOption) (*Server, error) {
	s, err := newServerConfig(opts...).baseServer()
	if err != nil {
		return nil, err
	}
	return &Server{
		h:          h,
		baseServer: s,
	}, nil
}
