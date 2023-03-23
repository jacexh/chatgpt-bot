package application

import "github.com/jacexh/chatgpt-bot/internal/chat/domain"

type Converstaion struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

type Chat struct {
	ID            string          `json:"id"`
	Channel       int             `json:"channel"`
	ChannelUserID string          `json:"channel_user_id"`
	Current       *Converstaion   `json:"current,omitempty"`
	Previous      []*Converstaion `json:"previous,omitempty"`
}

func AssembleEntidy(entity *domain.Chat) *Chat {
	c := &Chat{
		ID:            entity.ID,
		Channel:       int(entity.From.Channel),
		ChannelUserID: string(entity.From.ChannelUserID),
		Previous:      make([]*Converstaion, len(entity.PreviousConversations())),
	}
	if cov, err := entity.CurrentConversation(); err == nil {
		c.Current = &Converstaion{Prompt: cov.Prompt}
	}
	for index, conv := range entity.PreviousConversations() {
		c.Previous[index] = &Converstaion{
			Prompt:     conv.Prompt,
			Completion: conv.Completion,
		}
	}
	return c
}
