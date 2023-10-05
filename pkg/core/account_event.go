package core

import (
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
	tonapiClient "github.com/tonkeeper/opentonapi/client"
	"github.com/tonkeeper/tongo"
)

func scaleTons(amount int64) decimal.Decimal {
	value := big.NewInt(amount)
	return decimal.NewFromBigInt(value, int32(-9))
}

func scaleJettons(amount string, decimals int) decimal.Decimal {
	var value big.Int
	value.SetString(amount, 10)
	return decimal.NewFromBigInt(&value, int32(-decimals))
}

func formatTonTransfer(accountID tongo.AccountID, action tonapiClient.OptTonTransferAction) string {
	if action.Value.Recipient.Address == accountID.ToRaw() {
		return fmt.Sprintf("Received %v TON", scaleTons(action.Value.Amount))
	}
	return ""
}

func formatJettonTransfer(accountID tongo.AccountID, action tonapiClient.OptJettonTransferAction) string {
	if !action.Set {
		return ""
	}
	if !action.Value.Recipient.IsSet() {
		return ""
	}
	if action.Value.Recipient.Value.Address == accountID.ToRaw() {
		return fmt.Sprintf("Received %v %v", scaleJettons(action.Value.Amount, action.Value.Jetton.Decimals), action.Value.Jetton.Symbol)
	}
	return ""
}

func formatJettonMint(accountID tongo.AccountID, action tonapiClient.OptJettonMintAction) string {
	if action.Value.Recipient.Address == accountID.ToRaw() {
		return fmt.Sprintf("Received %v %v", scaleJettons(action.Value.Amount, action.Value.Jetton.Decimals), action.Value.Jetton.Symbol)
	}
	return ""
}

func formatNftTransfer(accountID tongo.AccountID, action tonapiClient.OptNftItemTransferAction) string {
	if !action.Value.Recipient.IsSet() {
		return ""
	}
	if action.Value.Recipient.Value.Address == accountID.ToRaw() {
		return fmt.Sprintf("Received NFT")
	}
	return ""
}

func formatNftPurchase(accountID tongo.AccountID, action tonapiClient.OptNftPurchaseAction) string {
	if action.Value.Buyer.Address == accountID.ToRaw() {
		return fmt.Sprintf("Received NFT")
	}
	return ""
}

func formatMessages(accountID tongo.AccountID, event *tonapiClient.AccountEvent) []string {
	var messages []string
	for _, action := range event.Actions {
		switch {
		case action.Type == tonapiClient.ActionTypeTonTransfer && action.TonTransfer.IsSet():
			if msg := formatTonTransfer(accountID, action.TonTransfer); len(msg) > 0 {
				messages = append(messages, msg)
			}
		case action.Type == tonapiClient.ActionTypeJettonTransfer && action.JettonTransfer.IsSet():
			if msg := formatJettonTransfer(accountID, action.JettonTransfer); len(msg) > 0 {
				messages = append(messages, msg)
			}
		case action.Type == tonapiClient.ActionTypeJettonMint && action.JettonMint.IsSet():
			if msg := formatJettonMint(accountID, action.JettonMint); len(msg) > 0 {
				messages = append(messages, msg)
			}
		case action.Type == tonapiClient.ActionTypeNftItemTransfer && action.NftItemTransfer.IsSet():
			if msg := formatNftTransfer(accountID, action.NftItemTransfer); len(msg) > 0 {
				messages = append(messages, msg)
			}
		case action.Type == tonapiClient.ActionTypeNftPurchase && action.NftPurchase.IsSet():
			if msg := formatNftPurchase(accountID, action.NftPurchase); len(msg) > 0 {
				messages = append(messages, msg)
			}
		case action.Type == tonapiClient.ActionTypeJettonSwap && action.JettonSwap.IsSet():
			if msg := action.SimplePreview.Description; len(msg) > 0 {
				messages = append(messages, msg)
			}
		}
	}
	return messages
}
