package telegram

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Telegram-Web-Apps/init-data-golang"
)

const (
	maxInitDataLifetime = 1 * time.Hour
)

// ExtractUserIDFromInitData extracts user ID from twa init data.
// See more details at https://docs.twa.dev/docs/libraries/init-data-golang.
func ExtractUserIDFromInitData(data string, telegramSecret string) (UserID, error) {
	twaInitData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return 0, fmt.Errorf("failed to decode init data")
	}
	if err := initdata.Validate(string(twaInitData), telegramSecret, maxInitDataLifetime); err != nil {
		return 0, fmt.Errorf("failed to validate init data")
	}
	parsedData, err := initdata.Parse(string(twaInitData))
	if err != nil {
		return 0, fmt.Errorf("failed to parse init data")
	}
	if parsedData.User == nil {
		return 0, fmt.Errorf("user not found in init data")
	}
	return UserID(parsedData.User.ID), nil
}
