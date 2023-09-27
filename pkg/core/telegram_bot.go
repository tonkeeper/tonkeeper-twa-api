package core

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type telegramBot struct {
	logger *zap.Logger
	bot    *tgbotapi.BotAPI
}

func newBot(logger *zap.Logger, botSecretKey string) (*telegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(botSecretKey)
	if err != nil {
		return nil, err
	}
	return &telegramBot{logger: logger, bot: bot}, nil
}

type Message struct {
	UserID TelegramUserID
	Text   string
}

func (b *telegramBot) Run(ctx context.Context) chan Message {
	ch := make(chan Message)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				message := tgbotapi.NewMessage(int64(msg.UserID), msg.Text)
				if _, err := b.bot.Send(message); err != nil {
					b.logger.Error("failed to send message", zap.Error(err))
				}
			}
		}
	}()
	return ch
}
