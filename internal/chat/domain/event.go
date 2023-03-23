package domain

import "github.com/go-jimu/components/mediator"

type MetaEvent struct {
	ChatID       string
	From         From
	Conversation Conversation
	kind         mediator.EventKind
	Error        error
}

type Event interface {
	mediator.Event
	Channel() Channel
}

const (
	KindConversationCreated    mediator.EventKind = "event_conversation_created"
	KindConversationReplied    mediator.EventKind = "event_conversation_replied"
	KindCoversationInterrupted mediator.EventKind = "event_conversation_interruptted"
)

func (me MetaEvent) Kind() mediator.EventKind {
	return me.kind
}

func (me MetaEvent) Channel() Channel {
	return me.From.Channel
}

func NewEvent(cid string, f From, c Conversation, kind mediator.EventKind) Event {
	return MetaEvent{
		ChatID:       cid,
		From:         f,
		Conversation: c,
		kind:         kind,
	}
}

func NewEventConversationCreated(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindConversationCreated)
}

func NewEventPromptAnswerd(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindConversationReplied)
}

func NewConversationInterrupted(cid string, f From, c Conversation, err error) Event {
	return &MetaEvent{
		ChatID:       cid,
		From:         f,
		Conversation: c,
		kind:         KindCoversationInterrupted,
		Error:        err,
	}
}
