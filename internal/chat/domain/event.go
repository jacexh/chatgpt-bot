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
	KindChatStartted           mediator.EventKind = "event_chat_startted"
	KindConversationCreated    mediator.EventKind = "event_conversation_created"
	KindConversationAnswered   mediator.EventKind = "event_conversation_answered"
	KindCoversationInterrupted mediator.EventKind = "event_conversation_interruptted"
	KindChatShutdown           mediator.EventKind = "event_chat_has_been_shutdown"
	KindChatFinished           mediator.EventKind = "event_chat_finished"
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

func NewEventChatStartted(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindChatStartted)
}

func NewEventConversationCreated(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindConversationCreated)
}

func NewEventPromptAnswerd(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindConversationAnswered)
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

func NewEventChatShutdown(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindChatShutdown)
}

func NewEventChatFinished(cid string, f From, c Conversation) Event {
	return NewEvent(cid, f, c, KindChatFinished)
}
