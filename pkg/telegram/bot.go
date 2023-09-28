package telegram

import (
	"context"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type UserID int64

type bot struct {
	logger *zap.Logger
	bot    *tgbotapi.BotAPI
}

func NewBot(logger *zap.Logger, botSecretKey string) (*bot, error) {
	tgbot, err := tgbotapi.NewBotAPI(botSecretKey)
	if err != nil {
		return nil, err
	}
	return &bot{logger: logger, bot: tgbot}, nil
}

type Message struct {
	UserID UserID
	Text   string
}

func (b *bot) Run(ctx context.Context) chan Message {
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
