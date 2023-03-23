package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

var _ domain.Repository = (*repository)(nil)

func NewRepository(db *sqlx.DB) domain.Repository {
	return &repository{db: db}
}

func (repo *repository) Get(ctx context.Context, from domain.From) (*domain.Chat, error) {
	type record struct {
		*Chat         `db:"c1"`
		*Conversation `db:"c2"`
	}
	var data []record
	err := repo.db.SelectContext(ctx, &data,
		"SELECT c1.id 'c1.id', c1.counts 'c1.counts', c1.current 'c1.current', c1.channel 'c1.channel', c1.channel_user_id 'c1.channel_user_id',"+
			" c1.version 'c1.version', c1.ctime 'c1.ctime', c1.mtime 'c1.mtime', c1.deleted 'c1.deleted', c2.id 'c2.id', c2.chat_id 'c2.chat_id', "+
			" c2.prompt 'c2.prompt', c2.completion 'c2.completion', c2.channel_message_id 'c2.channel_message_id', c2.ctime 'c2.ctime', c2.mtime 'c2.mtime' "+
			" FROM chat AS c1 LEFT JOIN conversation AS c2 ON c1.id=c2.chat_id WHERE c1.channel=? AND c1.channel_user_id=? AND c1.deleted=0 ORDER BY c2.id",
		from.Channel, from.ChannelUserID,
	)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, sql.ErrNoRows
	}
	convs := make([]*Conversation, 0)
	for _, rec := range data {
		if rec.Conversation.ID.Valid {
			convs = append(convs, rec.Conversation)
		}
	}
	return ConverDO(data[0].Chat, convs...)
}

func (repo *repository) GetByChatID(ctx context.Context, cid string) (*domain.Chat, error) {
	type record struct {
		*Chat         `db:"c1"`
		*Conversation `db:"c2"`
	}
	var data []record
	err := repo.db.SelectContext(ctx, &data,
		"SELECT c1.id 'c1.id', c1.counts 'c1.counts', c1.current 'c1.current', c1.channel 'c1.channel', c1.channel_user_id 'c1.channel_user_id',"+
			" c1.version 'c1.version', c1.ctime 'c1.ctime', c1.mtime 'c1.mtime', c1.deleted 'c1.deleted', c2.id 'c2.id', c2.chat_id 'c2.chat_id', "+
			" c2.prompt 'c2.prompt', c2.completion 'c2.completion', c2.channel_message_id 'c2.channel_message_id', c2.ctime 'c2.ctime', c2.mtime 'c2.mtime' "+
			" FROM chat AS c1 LEFT JOIN conversation AS c2 ON c1.id=c2.chat_id WHERE c1.id=? ORDER BY c2.id",
		cid,
	)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, sql.ErrNoRows
	}
	convs := make([]*Conversation, 0)
	for _, rec := range data {
		if rec.Conversation.ID.Valid {
			convs = append(convs, rec.Conversation)
		}
	}
	return ConverDO(data[0].Chat, convs...)
}

func (repo *repository) Save(ctx context.Context, chat *domain.Chat) error {
	if chat.Version == 0 { // 新增
		tx, err := repo.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return err
		}

		do, err := ConvertEntityChat(chat)
		if err != nil {
			return err
		}
		ret, err := tx.NamedExec("INSERT INTO chat (id, counts, current, channel, channel_user_id, version)  SELECT * FROM ( "+
			"SELECT :id as id, :counts as counts, :current as current, :channel as channel, :channel_user_id as channel_user_id, 1 as version) AS tmp "+
			"WHERE NOT EXISTS(SELECT * FROM chat WHERE id<>:id AND channel_user_id=:channel_user_id AND channel=:channel AND deleted=0 LIMIT 1)",
			do,
		)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		rows, err := ret.RowsAffected()
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if rows == 0 {
			_ = tx.Rollback()
			return errors.New("duplicated chat")
		}
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}
		return nil
	}

	// 更新记录
	do, err := ConvertEntityChat(chat)
	if err != nil {
		return err
	}
	tx, err := repo.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE chat SET counts=?, current=?, version=version+1, deleted=? WHERE id=? AND deleted=0", do.Counts, do.Current, do.Deleted, do.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if chat.Current == nil && len(chat.Conversations) > 0 { // conversation完成，则insert最后一条
		lastConversation := chat.Conversations[len(chat.Conversations)-1]
		_, err = tx.Exec("INSERT INTO conversation (chat_id, prompt, completion, channel_message_id) SELECT * FROM "+
			"(SELECT ? AS chat_id, ? AS prompt, ? AS completion, ? AS channel_message_id) AS tmp WHERE (SELECT COUNT(id) FROM conversation WHERE chat_id=?) < ?",
			do.ID, lastConversation.Prompt, lastConversation.Completion, lastConversation.MessageID, do.ID, len(chat.Conversations))
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}
