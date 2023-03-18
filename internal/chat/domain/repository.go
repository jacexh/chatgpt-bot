package domain

import "context"

type (
	Repository interface {
		Get(context.Context, From) (*Chat, error)
		Save(context.Context, *Chat) error
	}

	ChatGTPService interface {
		Chat(context.Context, *Chat) (*Conversation, error)
	}
)
