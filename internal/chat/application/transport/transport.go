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
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/message"
)

type controller struct {
	tgBot  *tgbotapi.BotAPI
	wechat *officialaccount.OfficialAccount
	app    *application.Application
}

var _ httpsrv.Controller = (*controller)(nil)

func NewController(app *application.Application, bot *tgbotapi.BotAPI, wc *officialaccount.OfficialAccount) httpsrv.Controller {
	return &controller{tgBot: bot, app: app, wechat: wc}
}

func (ctrl *controller) Slug() string {
	return "/api/v1"
}

func (ctrl *controller) APIs() []httpsrv.API {
	return []httpsrv.API{
		{
			Method:  http.MethodPost,
			Pattern: "/telegram/callback",
			Func:    ctrl.TelegramWebhook,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/chats/{chatID}",
			Func:    ctrl.Query,
		},
	}
}

func (ctrl *controller) Middlewares() []httpsrv.Middleware {
	return []httpsrv.Middleware{}
}

func (ctrl *controller) Query(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	dto, err := ctrl.app.GetByChatID(r.Context(), logger.FromContext(r.Context()), chatID)
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

func (ctrl *controller) TelegramWebhook(w http.ResponseWriter, r *http.Request) {
	helper := logger.FromContextAsHelper(r.Context()).WithContext(r.Context())
	update, err := ctrl.tgBot.HandleUpdate(r)
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
			if err = ctrl.app.NewChat(r.Context(), log, from); err != nil {
				chattable = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("[ERR] %s", err.Error()))
			} else {
				chattable = tgbotapi.NewMessage(update.Message.Chat.ID, "开始新的会话")
			}

		case "/end":
			ctrl.app.End(r.Context(), log, from)
			chattable = tgbotapi.NewMessage(update.Message.Chat.ID, "已结束当前会话")

		case "/current":
			details, err := ctrl.app.Get(r.Context(), log, from)
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
			if err = ctrl.app.Prompt(r.Context(), log, from, update.Message.Text, domain.ChannelMessageID(msgID)); err != nil {
				chattable = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("[ERR] %s", err.Error()))
			}
		}

		go func(msg tgbotapi.Chattable, log logger.Logger) {
			if msg != nil {
				if _, err = ctrl.tgBot.Send(msg); err != nil {
					logger.NewHelper(log).WithContext(r.Context()).Error("failed to send chat details", "error", err.Error())
				}
			}
		}(chattable, helper)
	}
}

func (ctrl *controller) WechatWebhook(w http.ResponseWriter, r *http.Request) {
	helper := logger.FromContextAsHelper(r.Context())
	server := ctrl.wechat.GetServer(r, w)

	server.SetMessageHandler(func(mm *message.MixMessage) *message.Reply {
		log := logger.With(logger.FromContext(r.Context()), "wechat_open_id", mm.FromUserName, "wechat_message_id", mm.MsgID)

		from := domain.From{
			Channel:       domain.ChannelWechat,
			ChannelUserID: domain.ChannelUserID(mm.FromUserName),
		}

		var text *message.Text

		switch mm.Content {
		case "/start":
			if err := ctrl.app.NewChat(r.Context(), log, from); err != nil {
				text = message.NewText("[ERR] " + err.Error())
			} else {
				text = message.NewText("开始新的会话")
			}

		case "/end":
			ctrl.app.End(r.Context(), log, from)
			text = message.NewText("已结束当前对话")

		case "/current":
			details, err := ctrl.app.Get(r.Context(), log, from)
			if err != nil {
				text = message.NewText("[ERR] " + err.Error())
			} else {
				data, _ := json.Marshal(details)
				text = message.NewText(string(data))
			}

		case "":
			return nil

		default:
			if err := ctrl.app.Prompt(r.Context(), log, from, mm.Content, domain.ChannelMessageID(mm.FromUserName)); err != nil {
				text = message.NewText("[ERR] " + err.Error())
			}
		}
		if text == nil {
			return nil
		}
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
	})

	if err := server.Serve(); err != nil {
		helper.WithContext(r.Context()).Error("failed to handle message", "error", err.Error())
		return
	}
	if err := server.Send(); err != nil {
		helper.WithContext(r.Context()).Error("failed to reply message", "error", err.Error())
	}
}
