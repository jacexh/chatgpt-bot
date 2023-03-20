package infrastructure

import "time"

type (
	Conversation struct {
		ID     int       `db:"id"`
		ChatID string    `db:"chat_id"`
		Prompt string    `db:"prompt"`
		Answer string    `db:"answer"`
		CTime  time.Time `db:"ctime"`
		MTime  time.Time `db:"mtime"`
	}

	Chat struct {
		ID                string    `db:"id"`
		Offset            int       `db:"offset"`
		Current           int       `db:"current"`
		Channel           int       `db:"channel"`
		ChannelUserID     string    `db:"channel_user_id"`
		ChannelInternalID string    `db:"channel_internal_id"`
		Version           int       `db:"version"`
		CTime             time.Time `db:"ctime"`
		MTime             time.Time `db:"mtime"`
		Deleted           bool      `db:"deleted"`
	}
)
