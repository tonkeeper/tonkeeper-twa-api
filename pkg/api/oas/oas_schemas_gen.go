// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"fmt"
)

func (s *ErrorStatusCode) Error() string {
	return fmt.Sprintf("code %d: %+v", s.StatusCode, s.Response)
}

type AccountEventsSubscriptionStatusOK struct {
	Subscribed bool `json:"subscribed"`
}

// GetSubscribed returns the value of Subscribed.
func (s *AccountEventsSubscriptionStatusOK) GetSubscribed() bool {
	return s.Subscribed
}

// SetSubscribed sets the value of Subscribed.
func (s *AccountEventsSubscriptionStatusOK) SetSubscribed(val bool) {
	s.Subscribed = val
}

type AccountEventsSubscriptionStatusReq struct {
	// Base64 encoded twa init data.
	TwaInitData string `json:"twa_init_data"`
	// Wallet or smart contract address.
	Address string `json:"address"`
}

// GetTwaInitData returns the value of TwaInitData.
func (s *AccountEventsSubscriptionStatusReq) GetTwaInitData() string {
	return s.TwaInitData
}

// GetAddress returns the value of Address.
func (s *AccountEventsSubscriptionStatusReq) GetAddress() string {
	return s.Address
}

// SetTwaInitData sets the value of TwaInitData.
func (s *AccountEventsSubscriptionStatusReq) SetTwaInitData(val string) {
	s.TwaInitData = val
}

// SetAddress sets the value of Address.
func (s *AccountEventsSubscriptionStatusReq) SetAddress(val string) {
	s.Address = val
}

// BridgeWebhookOK is response for BridgeWebhook operation.
type BridgeWebhookOK struct{}

type BridgeWebhookReq struct {
	Topic string `json:"topic"`
	Hash  string `json:"hash"`
}

// GetTopic returns the value of Topic.
func (s *BridgeWebhookReq) GetTopic() string {
	return s.Topic
}

// GetHash returns the value of Hash.
func (s *BridgeWebhookReq) GetHash() string {
	return s.Hash
}

// SetTopic sets the value of Topic.
func (s *BridgeWebhookReq) SetTopic(val string) {
	s.Topic = val
}

// SetHash sets the value of Hash.
func (s *BridgeWebhookReq) SetHash(val string) {
	s.Hash = val
}

type Error struct {
	Error string `json:"error"`
}

// GetError returns the value of Error.
func (s *Error) GetError() string {
	return s.Error
}

// SetError sets the value of Error.
func (s *Error) SetError(val string) {
	s.Error = val
}

// ErrorStatusCode wraps Error with StatusCode.
type ErrorStatusCode struct {
	StatusCode int
	Response   Error
}

// GetStatusCode returns the value of StatusCode.
func (s *ErrorStatusCode) GetStatusCode() int {
	return s.StatusCode
}

// GetResponse returns the value of Response.
func (s *ErrorStatusCode) GetResponse() Error {
	return s.Response
}

// SetStatusCode sets the value of StatusCode.
func (s *ErrorStatusCode) SetStatusCode(val int) {
	s.StatusCode = val
}

// SetResponse sets the value of Response.
func (s *ErrorStatusCode) SetResponse(val Error) {
	s.Response = val
}

type GetTonConnectPayloadOK struct {
	Payload string `json:"payload"`
}

// GetPayload returns the value of Payload.
func (s *GetTonConnectPayloadOK) GetPayload() string {
	return s.Payload
}

// SetPayload sets the value of Payload.
func (s *GetTonConnectPayloadOK) SetPayload(val string) {
	s.Payload = val
}

// NewOptString returns new OptString with value set to v.
func NewOptString(v string) OptString {
	return OptString{
		Value: v,
		Set:   true,
	}
}

// OptString is optional string.
type OptString struct {
	Value string
	Set   bool
}

// IsSet returns true if OptString was set.
func (o OptString) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptString) Reset() {
	var v string
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptString) SetTo(v string) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptString) Get() (v string, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptString) Or(d string) string {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptUint32 returns new OptUint32 with value set to v.
func NewOptUint32(v uint32) OptUint32 {
	return OptUint32{
		Value: v,
		Set:   true,
	}
}

// OptUint32 is optional uint32.
type OptUint32 struct {
	Value uint32
	Set   bool
}

// IsSet returns true if OptUint32 was set.
func (o OptUint32) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptUint32) Reset() {
	var v uint32
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptUint32) SetTo(v uint32) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptUint32) Get() (v uint32, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptUint32) Or(d uint32) uint32 {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// SubscribeToAccountEventsOK is response for SubscribeToAccountEvents operation.
type SubscribeToAccountEventsOK struct{}

type SubscribeToAccountEventsReq struct {
	// Base64 encoded twa init data.
	TwaInitData string `json:"twa_init_data"`
	// Wallet or smart contract address.
	Address string `json:"address"`
	// TON Connect proof of ownership of the address.
	Proof SubscribeToAccountEventsReqProof `json:"proof"`
}

// GetTwaInitData returns the value of TwaInitData.
func (s *SubscribeToAccountEventsReq) GetTwaInitData() string {
	return s.TwaInitData
}

// GetAddress returns the value of Address.
func (s *SubscribeToAccountEventsReq) GetAddress() string {
	return s.Address
}

// GetProof returns the value of Proof.
func (s *SubscribeToAccountEventsReq) GetProof() SubscribeToAccountEventsReqProof {
	return s.Proof
}

// SetTwaInitData sets the value of TwaInitData.
func (s *SubscribeToAccountEventsReq) SetTwaInitData(val string) {
	s.TwaInitData = val
}

// SetAddress sets the value of Address.
func (s *SubscribeToAccountEventsReq) SetAddress(val string) {
	s.Address = val
}

// SetProof sets the value of Proof.
func (s *SubscribeToAccountEventsReq) SetProof(val SubscribeToAccountEventsReqProof) {
	s.Proof = val
}

// TON Connect proof of ownership of the address.
type SubscribeToAccountEventsReqProof struct {
	Timestamp int64                                  `json:"timestamp"`
	Domain    SubscribeToAccountEventsReqProofDomain `json:"domain"`
	Signature string                                 `json:"signature"`
	Payload   string                                 `json:"payload"`
	StateInit OptString                              `json:"state_init"`
}

// GetTimestamp returns the value of Timestamp.
func (s *SubscribeToAccountEventsReqProof) GetTimestamp() int64 {
	return s.Timestamp
}

// GetDomain returns the value of Domain.
func (s *SubscribeToAccountEventsReqProof) GetDomain() SubscribeToAccountEventsReqProofDomain {
	return s.Domain
}

// GetSignature returns the value of Signature.
func (s *SubscribeToAccountEventsReqProof) GetSignature() string {
	return s.Signature
}

// GetPayload returns the value of Payload.
func (s *SubscribeToAccountEventsReqProof) GetPayload() string {
	return s.Payload
}

// GetStateInit returns the value of StateInit.
func (s *SubscribeToAccountEventsReqProof) GetStateInit() OptString {
	return s.StateInit
}

// SetTimestamp sets the value of Timestamp.
func (s *SubscribeToAccountEventsReqProof) SetTimestamp(val int64) {
	s.Timestamp = val
}

// SetDomain sets the value of Domain.
func (s *SubscribeToAccountEventsReqProof) SetDomain(val SubscribeToAccountEventsReqProofDomain) {
	s.Domain = val
}

// SetSignature sets the value of Signature.
func (s *SubscribeToAccountEventsReqProof) SetSignature(val string) {
	s.Signature = val
}

// SetPayload sets the value of Payload.
func (s *SubscribeToAccountEventsReqProof) SetPayload(val string) {
	s.Payload = val
}

// SetStateInit sets the value of StateInit.
func (s *SubscribeToAccountEventsReqProof) SetStateInit(val OptString) {
	s.StateInit = val
}

type SubscribeToAccountEventsReqProofDomain struct {
	LengthBytes OptUint32 `json:"length_bytes"`
	Value       string    `json:"value"`
}

// GetLengthBytes returns the value of LengthBytes.
func (s *SubscribeToAccountEventsReqProofDomain) GetLengthBytes() OptUint32 {
	return s.LengthBytes
}

// GetValue returns the value of Value.
func (s *SubscribeToAccountEventsReqProofDomain) GetValue() string {
	return s.Value
}

// SetLengthBytes sets the value of LengthBytes.
func (s *SubscribeToAccountEventsReqProofDomain) SetLengthBytes(val OptUint32) {
	s.LengthBytes = val
}

// SetValue sets the value of Value.
func (s *SubscribeToAccountEventsReqProofDomain) SetValue(val string) {
	s.Value = val
}

// SubscribeToBridgeEventsOK is response for SubscribeToBridgeEvents operation.
type SubscribeToBridgeEventsOK struct{}

type SubscribeToBridgeEventsReq struct {
	// Base64 encoded twa init data.
	TwaInitData string `json:"twa_init_data"`
	ClientID    string `json:"client_id"`
	Origin      string `json:"origin"`
}

// GetTwaInitData returns the value of TwaInitData.
func (s *SubscribeToBridgeEventsReq) GetTwaInitData() string {
	return s.TwaInitData
}

// GetClientID returns the value of ClientID.
func (s *SubscribeToBridgeEventsReq) GetClientID() string {
	return s.ClientID
}

// GetOrigin returns the value of Origin.
func (s *SubscribeToBridgeEventsReq) GetOrigin() string {
	return s.Origin
}

// SetTwaInitData sets the value of TwaInitData.
func (s *SubscribeToBridgeEventsReq) SetTwaInitData(val string) {
	s.TwaInitData = val
}

// SetClientID sets the value of ClientID.
func (s *SubscribeToBridgeEventsReq) SetClientID(val string) {
	s.ClientID = val
}

// SetOrigin sets the value of Origin.
func (s *SubscribeToBridgeEventsReq) SetOrigin(val string) {
	s.Origin = val
}

// UnsubscribeFromAccountEventsOK is response for UnsubscribeFromAccountEvents operation.
type UnsubscribeFromAccountEventsOK struct{}

type UnsubscribeFromAccountEventsReq struct {
	// Base64 encoded twa init data.
	TwaInitData string `json:"twa_init_data"`
}

// GetTwaInitData returns the value of TwaInitData.
func (s *UnsubscribeFromAccountEventsReq) GetTwaInitData() string {
	return s.TwaInitData
}

// SetTwaInitData sets the value of TwaInitData.
func (s *UnsubscribeFromAccountEventsReq) SetTwaInitData(val string) {
	s.TwaInitData = val
}

// UnsubscribeFromBridgeEventsOK is response for UnsubscribeFromBridgeEvents operation.
type UnsubscribeFromBridgeEventsOK struct{}

type UnsubscribeFromBridgeEventsReq struct {
	// Base64 encoded twa init data.
	TwaInitData string    `json:"twa_init_data"`
	ClientID    OptString `json:"client_id"`
}

// GetTwaInitData returns the value of TwaInitData.
func (s *UnsubscribeFromBridgeEventsReq) GetTwaInitData() string {
	return s.TwaInitData
}

// GetClientID returns the value of ClientID.
func (s *UnsubscribeFromBridgeEventsReq) GetClientID() OptString {
	return s.ClientID
}

// SetTwaInitData sets the value of TwaInitData.
func (s *UnsubscribeFromBridgeEventsReq) SetTwaInitData(val string) {
	s.TwaInitData = val
}

// SetClientID sets the value of ClientID.
func (s *UnsubscribeFromBridgeEventsReq) SetClientID(val OptString) {
	s.ClientID = val
}
