package transport

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-jimu/components/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/httpsrv"
	"github.com/jacexh/chatgpt-bot/internal/chat/application"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
)

type controller struct {
	bot *tgbotapi.BotAPI
	app *application.Application
}

var _ httpsrv.Controller = (*controller)(nil)

func NewController(bot *tgbotapi.BotAPI, app *application.Application) httpsrv.Controller {
	return &controller{bot: bot, app: app}
}

func (tg *controller) Slug() string {
	return "/api/v1"
}

func (tg *controller) APIs() []httpsrv.API {
	return []httpsrv.API{
		{
			Method:  http.MethodPost,
			Pattern: "/telegram/callback",
			Func:    tg.Webhook,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/chats/{chatID}",
			Func:    tg.Query,
		},
	}
}

func (tg *controller) Middlewares() []httpsrv.Middleware {
	return []httpsrv.Middleware{}
}

func (tg *controller) Query(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	dto, err := tg.app.GetByChatID(r.Context(), logger.FromContext(r.Context()), chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(dto)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (tg *controller) Webhook(w http.ResponseWriter, r *http.Request) {
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
		Channel:       domain.ChannelTelegram,
		ChannelUserID: domain.ChannelUserID(fmt.Sprintf("%d", update.Message.From.ID)),
	}

	log := logger.With(helper,
		"telegram_user_id", update.Message.From.ID,
		"telegram_chat_id", update.Message.Chat.ID,
		"telegram_message_id", update.Message.MessageID,
	)

	if update.Message != nil {
		var chattable tgbotapi.Chattable

		switch update.Message.Text {
		case "/start":
			if err = tg.app.NewChat(r.Context(), log, from); err != nil {
				chattable = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("[ERR] %s", err.Error()))
			} else {
				chattable = tgbotapi.NewMessage(update.Message.Chat.ID, "开始新的会话")
			}

		case "/end":
			tg.app.End(r.Context(), log, from)
			chattable = tgbotapi.NewMessage(update.Message.Chat.ID, "原会话已经结束")

		case "/current":
			details, err := tg.app.Get(r.Context(), log, from)
			var text string
			if err != nil {
				text = err.Error()
			} else {
				data, _ := json.Marshal(details)
				text = string(data)
			}
			chattable = tgbotapi.NewMessage(update.Message.Chat.ID, text)

		default:
			msgID := fmt.Sprintf("%d@%d", update.Message.MessageID, update.Message.Chat.ID)
			if err = tg.app.Prompt(r.Context(), log, from, update.Message.Text, domain.ChannelMessageID(msgID)); err != nil {
				chattable = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("[ERR] %s", err.Error()))
			}
		}

		go func(msg tgbotapi.Chattable, log logger.Logger) {
			if msg != nil {
				if _, err = tg.bot.Send(msg); err != nil {
					logger.NewHelper(log).WithContext(r.Context()).Error("failed to send chat details", "error", err.Error())
				}
			}
		}(chattable, helper)
	}
}
