package chat

import (
	"github.com/go-jimu/components/logger"
	"github.com/go-jimu/components/mediator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/httpsrv"
	"github.com/jacexh/chatgpt-bot/internal/chat/application"
	"github.com/jacexh/chatgpt-bot/internal/chat/application/transport"
	"github.com/jacexh/chatgpt-bot/internal/chat/infrastructure"
	"github.com/jmoiron/sqlx"
	"github.com/sashabaranov/go-openai"
	"github.com/silenceper/wechat/v2/officialaccount"
)

func Init(
	log logger.Logger,
	db *sqlx.DB,
	http httpsrv.HTTPServer,
	mediator mediator.Mediator,
	bot *tgbotapi.BotAPI,
	gpt *openai.Client,
	wc *officialaccount.OfficialAccount) {
	repo := infrastructure.NewRepository(db)
	gptSrv := infrastructure.NewChatGTPServer(gpt)
	app := application.NewApplication(repo, mediator, gptSrv)
	controller := transport.NewController(app, bot, wc)
	http.With(controller)

	handler := application.NewTelegramEventHandler(log, bot)
	mediator.Subscribe(handler)

	handler = application.NewWechatEventHandler(log, wc)
	mediator.Subscribe(handler)
}
