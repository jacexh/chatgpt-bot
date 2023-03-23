package infrastructure

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-jimu/components/mediator"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
)

type (
	Conversation struct {
		ID               sql.NullInt32  `db:"id"`
		ChatID           sql.NullString `db:"chat_id"`
		Prompt           sql.NullString `db:"prompt"`
		Completion       sql.NullString `db:"completion"`
		ChannelMessageID sql.NullString `db:"channel_message_id"`
		CTime            sql.NullTime   `db:"ctime"`
		MTime            sql.NullTime   `db:"mtime"`
	}

	Chat struct {
		ID            string    `db:"id"`
		Counts        int       `db:"counts"`
		Current       string    `db:"current"`
		Channel       int       `db:"channel"`
		ChannelUserID string    `db:"channel_user_id"`
		Version       int       `db:"version"`
		CTime         time.Time `db:"ctime"`
		MTime         time.Time `db:"mtime"`
		Deleted       int       `db:"deleted"`
	}
)

func ConverDO(ch *Chat, cs ...*Conversation) (*domain.Chat, error) {
	c := &domain.Chat{
		ID:            ch.ID,
		From:          domain.From{Channel: domain.Channel(ch.Channel), ChannelUserID: domain.ChannelUserID(ch.ChannelUserID)},
		Version:       ch.Version,
		Counts:        ch.Counts,
		Conversations: make([]*domain.Conversation, len(cs)),
		Status:        domain.StatusReady,
		CreatedAt:     ch.CTime,
		Event:         mediator.NewEventCollection(),
	}
	if ch.Current != "" && ch.Current != "{}" {
		cu := new(domain.Conversation)
		if err := json.Unmarshal([]byte(ch.Current), cu); err != nil {
			return nil, err
		}
		c.Current = cu
	}

	for index, con := range cs {
		c.Conversations[index] = &domain.Conversation{
			MessageID:  domain.ChannelMessageID(con.ChannelMessageID.String),
			Prompt:     con.Prompt.String,
			Completion: con.Completion.String,
		}
	}
	return c, nil
}

func ConvertEntityChat(entity *domain.Chat) (*Chat, error) {
	c := &Chat{
		ID:            entity.ID,
		Counts:        entity.Counts,
		Channel:       int(entity.From.Channel),
		ChannelUserID: string(entity.From.ChannelUserID),
		Version:       entity.Version,
		Deleted:       int(entity.Status),
	}
	if entity.Current != nil {
		data, err := json.Marshal(entity.Current)
		if err != nil {
			return nil, err
		}

		str := string(data)
		if str != "" && str != "{}" {
			c.Current = str
		}
	}
	return c, nil
}
