package telegram

import (
	"context"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	messageCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "message_counter",
		Help: "Number of telegram messages sent",
	})
)

// UserID is an identifier of a telegram user.
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
				go b.sendTextMessage(msg.UserID, msg.Text)
			}
		}
	}()
	return ch
}

func (b *bot) sendTextMessage(userID UserID, text string) {
	messageCounter.Inc()

	message := tgbotapi.NewMessage(int64(userID), text)
	if _, err := b.bot.Send(message); err != nil {
		// TODO: maybe we should retry sending a message?
		b.logger.Error("failed to send message", zap.Error(err))
	}
}
