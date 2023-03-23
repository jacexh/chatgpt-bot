package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-jimu/components/logger"
	"github.com/go-jimu/components/mediator"
	"github.com/jacexh/chatgpt-bot/internal/chat/domain"
	pkgCtx "github.com/jacexh/chatgpt-bot/internal/pkg/context"
)

type Application struct {
	repo     domain.Repository
	mediator mediator.Mediator
	api      domain.ChatGTPService
}

func NewApplication(repo domain.Repository, mediator mediator.Mediator, api domain.ChatGTPService) *Application {
	return &Application{repo: repo, mediator: mediator, api: api}
}

func (app *Application) NewChat(ctx context.Context, log logger.Logger, f domain.From) error {
	chat := domain.NewChat(f)
	err := app.repo.Save(ctx, chat)
	if err != nil {
		logger.NewHelper(log).WithContext(ctx).Error("failed to start new chat", "chat_id", chat.ID, "error", err.Error())
		return err
	}
	chat.Event.Raise(app.mediator)
	return nil
}

func (app *Application) Get(ctx context.Context, log logger.Logger, f domain.From) (*Chat, error) {
	helper := logger.NewHelper(log).WithContext(ctx)
	chat, err := app.repo.Get(ctx, f)
	if err != nil {
		helper.Error("failed to get chat", "error", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("没有进行中的会话")
		}
		return nil, err
	}
	helper.Info("fetched chat details")
	return AssembleEntidy(chat), nil
}

func (app *Application) Prompt(ctx context.Context, log logger.Logger, f domain.From, q string, msgID domain.ChannelMessageID) error {
	helper := logger.NewHelper(log).WithContext(ctx)
	chat, err := app.repo.Get(ctx, f)
	if err != nil {
		helper.Error("failed to get chat", "error", err.Error())
		return err
	}

	if err = chat.Prompt(q, msgID); err != nil {
		helper.Error("failed to prompt", "chat_id", chat.ID, "error", err.Error())
		return err
	}

	if err = app.repo.Save(ctx, chat); err != nil {
		helper.Error("failet to save chat", "chat_id", chat.ID, "error", err.Error())
		return err
	}
	chat.Event.Raise(app.mediator)
	helper.Info("prompt", "chat_id", chat.ID, "prompt", q)

	go func(f domain.From, log logger.Logger) {
		ctx, cancel := pkgCtx.GenContextWithTimeout(3 * time.Minute)
		defer cancel()
		helper := logger.NewHelper(log).WithContext(ctx)

		chat, err := app.repo.Get(ctx, f)
		if err != nil {
			helper.Error("failed to get chat from repository to call chatgpt api", "chat_id", chat.ID, "error", err.Error())
			return
		}
		conv, err := app.api.Chat(ctx, chat)
		if err != nil {
			helper.Error("failed to get completion from chatgpt", "chat_id", chat.ID, "error", err.Error())
			return
		}
		helper.Info("got completion", "chat_id", chat.ID, "answer", conv.Completion)

		if err := app.repo.Save(ctx, chat); err != nil {
			helper.Error("failed to save chat", "chat_id", chat.ID, "error", err.Error())
			return
		}
		chat.Event.Raise(app.mediator)
	}(f, log)

	return nil
}

func (app *Application) End(ctx context.Context, log logger.Logger, f domain.From) {
	helper := logger.NewHelper(log).WithContext(ctx)
	chat, err := app.repo.Get(ctx, f)
	if err != nil {
		helper.Warn("failed to get chat to shutdown", "error", err.Error())
		return
	}

	chat.Shutdown()
	if err = app.repo.Save(ctx, chat); err != nil {
		helper.Warn("failed to save ended chat", "error", err.Error())
	}
	chat.Event.Raise(app.mediator)
}
