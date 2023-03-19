package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-jimu/components/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jacexh/chatgpt-bot/internal/chat/application"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
)

type controller struct {
	bot *tgbotapi.BotAPI
	app *application.Application
}

func (tg *controller) Handle(w http.ResponseWriter, r *http.Request) {
	helper := logger.FromContextAsHelper(r.Context()).WithContext(r.Context())
	update, err := tg.bot.HandleUpdate(r)
	if err != nil {
		helper.Error("received invalid callback from telegram", "error", err.Error())
		errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(errMsg)
		return
	}

	from := domain.From{
		Channel:           domain.ChannelTelegram,
		ChannelUserID:     fmt.Sprintf("%d", update.Message.From.ID),
		ChannelInternalID: fmt.Sprintf("%d+%d", update.Message.Chat.ID, update.Message.MessageID),
	}

	log := logger.With(helper,
		"telegram_user_id", update.Message.From.ID,
		"telegram_chat_id", update.Message.Chat.ID,
		"telegram_message_id", update.Message.MessageID,
	)

	if update.Message != nil {
		switch update.Message.Text {
		case "/start":
			_ = tg.app.NewChat(r.Context(), log, from)

		case "/end":
			tg.app.End(r.Context(), log, from)
		default:
			_ = tg.app.Prompt(r.Context(), log, from, update.Message.Text)
		}
	}
}