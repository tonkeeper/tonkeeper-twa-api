package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tonapiClient "github.com/tonkeeper/opentonapi/client"
	"github.com/tonkeeper/tongo"
)

func Test_formatMessages(t *testing.T) {
	tests := []struct {
		name      string
		accountID tongo.AccountID
		eventID   string
		want      []string
	}{
		{
			name:      "received nft",
			accountID: tongo.MustParseAddress("EQDdYTAOAGD4AjM2OztKDzsnrQOxnMS-xux5iqsLPkeeuorE").ID,
			eventID:   "e68822c836dd87a99e2b4172693add64e85594aeaf62623eca92ac0e189f8bcf",
			want: []string{
				"Received NFT",
			},
		},
		{
			name:      "received ton",
			accountID: tongo.MustParseAddress("EQDdYTAOAGD4AjM2OztKDzsnrQOxnMS-xux5iqsLPkeeuorE").ID,
			eventID:   "fea6cac8da213b0e7a05dcd5eb5daa55550a6c1d2ff8d6e75b309e649333fe73",
			want: []string{
				"Received 0.1 TON",
			},
		},
		{
			name:      "sent nft",
			accountID: tongo.MustParseAddress("EQDdYTAOAGD4AjM2OztKDzsnrQOxnMS-xux5iqsLPkeeuorE").ID,
			eventID:   "a24664bf86739c912d9ed94d68a3409d7414adc272be8a773fffbf8b5f8f5841",
			want:      nil,
		},
		{
			name:      "received jetton",
			accountID: tongo.MustParseAddress("EQDdYTAOAGD4AjM2OztKDzsnrQOxnMS-xux5iqsLPkeeuorE").ID,
			eventID:   "4e3697c01dcee652ead6b7625bf8efa7a35efbd0fffa0dd7b58ff13f1a230e35",
			want: []string{
				"Received 0.05 jUSDC",
			},
		},
		{
			name:      "swap jettons",
			accountID: tongo.MustParseAddress("EQBszTJahYw3lpP64ryqscKQaDGk4QpsO7RO6LYVvKHSINS0").ID,
			eventID:   "43de17434a703a3d20f1e9c016c1ab14c6ec4a1eebedcc7f5615fd87ed2a1612",
			want: []string{
				"Swapping 0.1 WTON for 0.000078233399479493 oETH",
			},
		},
		{
			name:      "nft purchase",
			accountID: tongo.MustParseAddress("EQBszTJahYw3lpP64ryqscKQaDGk4QpsO7RO6LYVvKHSINS0").ID,
			eventID:   "d2f0bec210c5adf20929c12680359006c5ab56e440b58120a785f05b3e5f02a2",
			want: []string{
				"Received NFT",
			},
		},
		{
			name:      "nft mint",
			accountID: tongo.MustParseAddress("EQBszTJahYw3lpP64ryqscKQaDGk4QpsO7RO6LYVvKHSINS0").ID,
			eventID:   "27ac37b27880386337ab1c80b533e99c35a3756d23458c36171fe1a09bd678af",
			want: []string{
				"Received NFT",
			},
		},
	}
	for _, tt := range tests {
		time.Sleep(3 * time.Second)
		t.Run(tt.name, func(t *testing.T) {
			cli, err := tonapiClient.NewClient("https://tonapi.io")
			require.Nil(t, err)

			params := tonapiClient.GetAccountEventParams{
				AccountID:   tt.accountID.ToRaw(),
				EventID:     tt.eventID,
				SubjectOnly: tonapiClient.OptBool{Value: true, Set: true},
			}

			event, err := cli.GetAccountEvent(context.Background(), params)
			require.Nil(t, err)

			messages := formatMessages(tt.accountID, event)
			fmt.Printf("%v\n", messages)
			require.Equal(t, tt.want, messages)
		})
	}
}
