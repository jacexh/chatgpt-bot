package domain_test

import (
	"testing"

	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestChat(t *testing.T) {
	chat := domain.NewChat(domain.From{ChannelUserID: domain.ChannelUserID(ulid.Make().String()), Channel: domain.ChannelTelegram})
	err := chat.Prompt("foobar", domain.ChannelMessageID(ulid.Make().String()))
	assert.NoError(t, err)
	err = chat.Prompt("foobar", domain.ChannelMessageID(ulid.Make().String()))
	assert.Error(t, err)

	c, err := chat.Reply("foobar answer")
	assert.NoError(t, err)
	t.Log(c)

	previsoue := chat.PreviousConversations()
	assert.Equal(t, len(previsoue), 1)
}
