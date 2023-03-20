package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-jimu/components/mediator"
	"github.com/oklog/ulid/v2"
)

type (
	Conversation struct {
		Prompt string
		Answer string
	}

	From struct {
		ChannelUserID     string
		Channel           Channel
		ChannelInternalID string
	}

	Channel int

	Status int

	Chat struct {
		ID            string
		From          From
		Conversations []*Conversation
		Status        Status
		Event         mediator.EventCollection
		Current       *Conversation
		Version       int
		Counts        int
		CreatedAt     time.Time
	}
)

const (
	ChannelTelegram Channel = iota + 1
	ChannelWechat
	ChannelDingtalk
)

const (
	StatusReady Status = iota
	StatusEnded
)

const (
	MaxConversationCounts = 20
	ChatExpirationTime    = 12 * time.Hour
)

var (
	emptyConversation = Conversation{}
)

func NewConversation(prompt string) *Conversation {
	return &Conversation{Prompt: prompt}
}

func (c *Conversation) Reply(answer string) {
	c.Answer = answer
}

func (c *Conversation) IsAnswed() bool {
	return c.Answer != ""
}

func (c *Conversation) String() string {
	return fmt.Sprintf("Prompt: %s\tAnswer: %s", c.Prompt, c.Answer)
}

func NewChat(f From) *Chat {
	ct := &Chat{
		ID:            ulid.Make().String(),
		From:          f,
		Conversations: make([]*Conversation, 0),
		Status:        StatusReady,
		Event:         mediator.NewEventCollection(),
		Version:       0,
		CreatedAt:     time.Now(),
	}
	ct.Event.Add(NewEventChatStartted(ct.ID, f, Conversation{}))
	return ct
}

func (ct *Chat) PreviousConversations() []*Conversation {
	return ct.Conversations[:]
}

func (ct *Chat) CurrentConversation() (*Conversation, error) {
	if ct.Current == nil {
		return nil, errors.New("no conversation")
	}
	return ct.Current, nil
}

func (ct *Chat) Prompt(q string) error {
	if ct.Current != nil {
		return errors.New("the previouse conversation has not yet ended")
	}
	if q == "" {
		return errors.New("disallow empty prompt")
	}
	ct.Current = NewConversation(q)
	ct.Counts++
	ct.Event.Add(NewEventConversationCreated(ct.ID, ct.From, *ct.Current))
	return nil
}

func (ct *Chat) Reply(a string) (*Conversation, error) {
	if ct.Current == nil {
		return nil, errors.New("there is no ongoing conversation")
	}

	if a == "" {
		return nil, errors.New("disallow empty string")
	}
	if ct.Current.IsAnswed() {
		return ct.Current, errors.New("current prompt has already been answered")
	}
	ct.Current.Reply(a)
	ct.Conversations = append(ct.Conversations, ct.Current)
	current := ct.Current
	ct.Current = nil
	ct.Event.Add(NewEventPromptAnswerd(ct.ID, ct.From, *current))
	return current, nil
}

func (ct *Chat) Interrupt(err error) (*Conversation, error) {
	current := ct.Current
	ct.Current = nil
	ct.Counts--
	ct.Event.Add(NewConversationInterrupted(ct.ID, ct.From, *current))
	return current, err
}

func (ct *Chat) Shutdown() {
	if ct.Status != StatusEnded {
		ct.Status = StatusEnded
		ct.Event.Add(NewEventChatShutdown(ct.ID, ct.From, emptyConversation))
	}
}

func (ct *Chat) IsFinished() bool {
	if ct.Status == StatusEnded {
		return true
	}
	if ct.Current != nil {
		return false
	}
	if time.Since(ct.CreatedAt) >= ChatExpirationTime || len(ct.Conversations) >= MaxConversationCounts {
		ct.Status = StatusEnded
		ct.Event.Add(NewEventChatFinished(ct.ID, ct.From, emptyConversation))
		return true
	}
	return false
}
