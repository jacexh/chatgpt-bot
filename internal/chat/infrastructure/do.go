package infrastructure

import (
	"database/sql"
	"time"

	"github.com/go-jimu/components/mediator"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
)

type (
	Conversation struct {
		ID     sql.NullInt32  `db:"id"`
		ChatID sql.NullString `db:"chat_id"`
		Prompt sql.NullString `db:"prompt"`
		Answer sql.NullString `db:"answer"`
		CTime  sql.NullTime   `db:"ctime"`
		MTime  sql.NullTime   `db:"mtime"`
	}

	Chat struct {
		ID                string    `db:"id"`
		Counts            int       `db:"counts"`
		CurrentPrompt     string    `db:"current_prompt"`
		Channel           int       `db:"channel"`
		ChannelUserID     string    `db:"channel_user_id"`
		ChannelInternalID string    `db:"channel_internal_id"`
		Version           int       `db:"version"`
		CTime             time.Time `db:"ctime"`
		MTime             time.Time `db:"mtime"`
		Deleted           int       `db:"deleted"`
	}
)

func ConverDO(ch *Chat, cs ...*Conversation) *domain.Chat {
	c := &domain.Chat{
		ID:            ch.ID,
		From:          domain.From{Channel: domain.Channel(ch.Channel), ChannelUserID: ch.ChannelUserID, ChannelInternalID: ch.ChannelInternalID},
		Version:       ch.Version,
		Counts:        ch.Counts,
		Conversations: make([]*domain.Conversation, len(cs)),
		Status:        domain.StatusReady,
		CreatedAt:     ch.CTime,
		Event:         mediator.NewEventCollection(),
	}
	if ch.CurrentPrompt != "" {
		c.Current = &domain.Conversation{Prompt: ch.CurrentPrompt}
	}

	for index, con := range cs {
		c.Conversations[index] = &domain.Conversation{
			Prompt: con.Prompt.String,
			Answer: con.Answer.String,
		}
	}
	return c
}

func ConvertEntityChat(entity *domain.Chat) *Chat {
	c := &Chat{
		ID:                entity.ID,
		Counts:            entity.Counts,
		Channel:           int(entity.From.Channel),
		ChannelUserID:     entity.From.ChannelUserID,
		ChannelInternalID: entity.From.ChannelInternalID,
		Version:           entity.Version,
		Deleted:           int(entity.Status),
	}
	if entity.Current != nil {
		c.CurrentPrompt = entity.Current.Prompt
	}
	return c
}
