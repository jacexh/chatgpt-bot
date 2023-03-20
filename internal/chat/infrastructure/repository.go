package infrastructure

import (
	"context"

	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

var _ domain.Repository = (*repository)(nil)

func (repo *repository) Get(ctx context.Context, from domain.From) (*domain.Chat, error) {
	// type record struct {
	// 	*Chat         `db:"c1"`
	// 	*Conversation `db:"c2"`
	// }
	// var data []record
	// repo.db.SelectContext(ctx)
	return nil, nil
}

func (repo *repository) Save(ctx context.Context, chat *domain.Chat) error {
	return nil
}
