package telegram

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Telegram-Web-Apps/init-data-golang"
)

func ExtractUserIDFromInitData(data string, telegramSecret string) (UserID, error) {
	// TODO: use right duration
	twaInitData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return 0, fmt.Errorf("failed to decode init data")
	}
	if err := initdata.Validate(string(twaInitData), telegramSecret, time.Duration(0)); err != nil {
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
