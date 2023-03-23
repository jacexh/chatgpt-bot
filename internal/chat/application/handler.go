package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-jimu/components/logger"
	"github.com/go-jimu/components/mediator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/message"
)

type TelegramEventHandler struct {
	bot *tgbotapi.BotAPI
	log logger.Logger
}

type WechatEventHandler struct {
	wechat *officialaccount.OfficialAccount
	log    logger.Logger
}

func NewTelegramEventHandler(log logger.Logger, bot *tgbotapi.BotAPI) mediator.EventHandler {
	return &TelegramEventHandler{log: log, bot: bot}
}

func (ev *TelegramEventHandler) Listening() []mediator.EventKind {
	return []mediator.EventKind{
		domain.KindConversationCreated,
		domain.KindCoversationInterrupted,
		domain.KindConversationReplied,
	}
}

func (ev *TelegramEventHandler) Handle(ctx context.Context, event mediator.Event) {
	e := event.(domain.MetaEvent)
	if e.Channel() != domain.ChannelTelegram {
		return
	}

	log := logger.With(ev.log, "chat_id", e.ChatID, "telegram_user_id", e.From.ChannelUserID, "message_id", e.Conversation.MessageID, "event_kind", event.Kind())
	helper := logger.NewHelper(log)

	slices := strings.Split(string(e.Conversation.MessageID), "@")
	if len(slices) != 2 {
		helper.Error("failed to parse telegram chat/message id from converstaion")
		return
	}
	msgID, err := strconv.ParseInt(slices[0], 10, 0)
	if err != nil {
		helper.Error("failed to parse message id", "error", err.Error())
		return
	}

	chatID, err := strconv.ParseInt(slices[1], 10, 0)
	if err != nil {
		helper.Error("failed to parse chat id", "error", err.Error())
		return
	}

	var chattable tgbotapi.Chattable
	switch e.Kind() {
	case domain.KindConversationCreated:
		chattable = tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
		if _, err := ev.bot.Request(chattable); err != nil {
			helper.Error("failed to set chat action", "error", err.Error())
		}
		return

	case domain.KindConversationReplied:
		chattable = tgbotapi.NewMessage(chatID, e.Conversation.Completion)
		msg := chattable.(tgbotapi.MessageConfig)
		msg.ReplyToMessageID = int(msgID)
		chattable = msg

	case domain.KindCoversationInterrupted:
		chattable = tgbotapi.NewMessage(chatID, fmt.Sprintf("[ERR] %s", e.Error.Error()))
		msg := chattable.(tgbotapi.MessageConfig)
		msg.ReplyToMessageID = int(msgID)
		chattable = msg
		helper.Error("current conversation was interrupted", "error", e.Error.Error())
	}

	if chattable != nil {
		if _, err := ev.bot.Send(chattable); err != nil {
			helper.Error("failed to send message to telegram", "error", err.Error())
		}
	}
}

func NewWechatEventHandler(log logger.Logger, wechat *officialaccount.OfficialAccount) mediator.EventHandler {
	return &WechatEventHandler{
		log:    log,
		wechat: wechat,
	}
}

func (w *WechatEventHandler) Listening() []mediator.EventKind {
	return []mediator.EventKind{
		domain.KindConversationCreated,
		domain.KindCoversationInterrupted,
		domain.KindConversationReplied,
	}
}

func (w *WechatEventHandler) Handle(ctx context.Context, ev mediator.Event) {
	event := ev.(domain.MetaEvent)
	if event.Channel() != domain.ChannelWechat {
		return
	}
	log := logger.With(w.log, "chat_id", event.ChatID, "telegram_user_id", event.From.ChannelUserID, "message_id", event.Conversation.MessageID, "event_kind", event.Kind())
	helper := logger.NewHelper(log)

	var msg *message.CustomerMessage
	switch event.Kind() {
	case domain.KindConversationCreated:

	case domain.KindConversationReplied:
		msg = &message.CustomerMessage{
			ToUser:  string(event.Conversation.MessageID),
			Msgtype: message.MsgTypeText,
			Text:    &message.MediaText{Content: event.Conversation.Completion},
		}

	case domain.KindCoversationInterrupted:
		msg = &message.CustomerMessage{
			ToUser:  string(event.Conversation.MessageID),
			Msgtype: message.MsgTypeText,
			Text:    &message.MediaText{Content: "[ERR] " + event.Error.Error()},
		}
	}

	if msg != nil {
		if err := w.wechat.GetCustomerMessageManager().Send(msg); err != nil {
			helper.Error("failed to send custom message to wechat user", "error", err.Error())
		}
	}
}
