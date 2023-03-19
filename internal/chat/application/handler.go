package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-jimu/components/logger"
	"github.com/go-jimu/components/mediator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
)

type TelegramEventHandler struct {
	bot *tgbotapi.BotAPI
	log logger.Logger
}

func (ev *TelegramEventHandler) Listening() []mediator.EventKind {
	return []mediator.EventKind{
		domain.KindChatFinished,
		domain.KindChatShutdown,
		domain.KindChatStartted,
		domain.KindConversationCreated,
		domain.KindCoversationInterrupted,
		domain.KindConversationAnswered,
	}
}

func (ec *TelegramEventHandler) ParseIDs(internalID string) (int64, int, error) {
	slice := strings.Split(internalID, "+")
	if len(slice) != 2 {
		return 0, 0, errors.New("bad internal id")
	}
	cid, err := strconv.ParseInt(slice[0], 10, 0)
	if err != nil {
		return 0, 0, err
	}
	mid, err := strconv.Atoi(slice[1])
	if err != nil {
		return 0, 0, err
	}
	return cid, mid, nil
}

func (ev *TelegramEventHandler) Handle(ctx context.Context, event mediator.Event) {
	e := event.(domain.MetaEvent)
	if e.Channel() != domain.ChannelTelegram {
		return
	}

	log := logger.With(ev.log, "chat_id", e.ChatID, "telegram_user_id", e.From.ChannelUserID, "telegrame_internal_id", e.From.ChannelInternalID, "event_kind", event.Kind())
	helper := logger.NewHelper(log)

	chatID, msgID, err := ev.ParseIDs(e.From.ChannelInternalID)
	if err != nil {
		helper.Error("failed to parse internal id", "error", err.Error())
		return
	}

	var chattable tgbotapi.Chattable
	switch e.Kind() {
	case domain.KindChatStartted:
		return

	case domain.KindChatFinished, domain.KindChatShutdown:
		chattable = tgbotapi.NewMessage(chatID, "The current chat has been ended.")

	case domain.KindConversationCreated:
		chattable = tgbotapi.NewChatAction(chatID, "typing...")

	case domain.KindConversationAnswered:
		chattable = tgbotapi.NewMessage(chatID, e.Conversation.Answer)
		mc := chattable.(tgbotapi.MessageConfig)
		mc.ReplyToMessageID = msgID

	case domain.KindCoversationInterrupted:
		chattable = tgbotapi.NewMessage(chatID, fmt.Sprintf("[ERR] this conversation was be interruptted, you can try again"))
		mc := chattable.(tgbotapi.MessageConfig)
		mc.ReplyToMessageID = msgID
	}

	if chattable != nil {
		if _, err := ev.bot.Send(chattable); err != nil {
			helper.Error("failed to send message to telegram", "error", err.Error())
		}
	}
}
