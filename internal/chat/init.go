package chat

import (
	"github.com/go-jimu/components/mediator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/httpsrv"
	"github.com/jacexh/chatgpt-bot/internal/chat/application"
	"github.com/jacexh/chatgpt-bot/internal/chat/application/transport"
	"github.com/jacexh/chatgpt-bot/internal/chat/infrastructure"
	"github.com/jmoiron/sqlx"
)

func Init(db *sqlx.DB, http httpsrv.HTTPServer, mediator mediator.Mediator, bot *tgbotapi.BotAPI, opt infrastructure.ChatGPTOption) {
	repo := infrastructure.NewRepository(db)
	gpt := infrastructure.NewChatGTPServer(opt)
	app := application.NewApplication(repo, mediator, gpt)
	controller := transport.NewController(bot, app)
	http.With(controller)
}
