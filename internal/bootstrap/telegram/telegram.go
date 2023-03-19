package telegram

import (
	"fmt"

	"github.com/go-jimu/components/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Option struct {
	WebhookLink string `json:"webhook_link" yaml:"webhook_link"`
	AccessToken string `json:"access_token" yaml:"access_token"`
}

func NewBotAPI(opt Option, log logger.Logger) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(opt.AccessToken)
	if err != nil {
		panic(err)
	}

	wc, err := tgbotapi.NewWebhook(opt.WebhookLink)
	if err != nil {
		panic(err)
	}

	_, err = bot.Request(wc)
	if err != nil {
		panic(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		panic(err)
	}
	logger.NewHelper(log).Info(fmt.Sprintf("telegram webhook link: %s", info.URL))
	return bot
}
