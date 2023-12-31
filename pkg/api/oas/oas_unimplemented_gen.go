// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"context"

	ht "github.com/ogen-go/ogen/http"
)

// UnimplementedHandler is no-op Handler which returns http.ErrNotImplemented.
type UnimplementedHandler struct{}

var _ Handler = UnimplementedHandler{}

// AccountEventsSubscriptionStatus implements accountEventsSubscriptionStatus operation.
//
// Get a status of an account-events subscription.
//
// POST /account-events/subscription-status
func (UnimplementedHandler) AccountEventsSubscriptionStatus(ctx context.Context, req *AccountEventsSubscriptionStatusReq) (r *AccountEventsSubscriptionStatusOK, _ error) {
	return r, ht.ErrNotImplemented
}

// BridgeWebhook implements bridgeWebhook operation.
//
// Webhook called by the HTTP Bridge when an event occurs.
//
// POST /bridge/webhook/{client_id}
func (UnimplementedHandler) BridgeWebhook(ctx context.Context, req *BridgeWebhookReq, params BridgeWebhookParams) error {
	return ht.ErrNotImplemented
}

// GetTonConnectPayload implements getTonConnectPayload operation.
//
// Get a challenge for TON Connect.
//
// GET /tonconnect/payload
func (UnimplementedHandler) GetTonConnectPayload(ctx context.Context) (r *GetTonConnectPayloadOK, _ error) {
	return r, ht.ErrNotImplemented
}

// SubscribeToAccountEvents implements subscribeToAccountEvents operation.
//
// Subscribe to notifications about events in the TON blockchain for a specific address.
//
// POST /account-events/subscribe
func (UnimplementedHandler) SubscribeToAccountEvents(ctx context.Context, req *SubscribeToAccountEventsReq) error {
	return ht.ErrNotImplemented
}

// SubscribeToBridgeEvents implements subscribeToBridgeEvents operation.
//
// Subscribe to notifications from the HTTP Bridge regarding a specific smart contract or wallet.
//
// POST /bridge/subscribe
func (UnimplementedHandler) SubscribeToBridgeEvents(ctx context.Context, req *SubscribeToBridgeEventsReq) error {
	return ht.ErrNotImplemented
}

// UnsubscribeFromAccountEvents implements unsubscribeFromAccountEvents operation.
//
// Unsubscribe from notifications about events in the TON blockchain for a specific address.
//
// POST /account-events/unsubscribe
func (UnimplementedHandler) UnsubscribeFromAccountEvents(ctx context.Context, req *UnsubscribeFromAccountEventsReq) error {
	return ht.ErrNotImplemented
}

// UnsubscribeFromBridgeEvents implements unsubscribeFromBridgeEvents operation.
//
// Unsubscribe from bridge notifications.
//
// POST /bridge/unsubscribe
func (UnimplementedHandler) UnsubscribeFromBridgeEvents(ctx context.Context, req *UnsubscribeFromBridgeEventsReq) error {
	return ht.ErrNotImplemented
}

// NewError creates *ErrorStatusCode from error returned by handler.
//
// Used for common default response.
func (UnimplementedHandler) NewError(ctx context.Context, err error) (r *ErrorStatusCode) {
	r = new(ErrorStatusCode)
	return r
}
