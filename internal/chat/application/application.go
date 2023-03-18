package application

import (
	"context"
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

func (app *Application) NewChat(ctx context.Context, log logger.Logger, f domain.From) error {
	chat := domain.NewChat(f)
	err := app.repo.Save(ctx, chat)
	if err != nil {
		logger.NewHelper(log).WithContext(ctx).Error("failed to start new chat", "error", err.Error())
		return err
	}
	chat.Event.Raise(app.mediator)
	return nil
}

func (app *Application) Prompt(ctx context.Context, log logger.Logger, f domain.From, q string) error {
	helper := logger.NewHelper(log).WithContext(ctx)
	chat, err := app.repo.Get(ctx, f)
	if err != nil {
		helper.Error("failed to get chat", "error", err.Error())
		return err
	}

	if err = chat.Prompt(q); err != nil {
		helper.Error("failed to prompt", "error", err.Error())
		return err
	}

	if err = app.repo.Save(ctx, chat); err != nil {
		helper.Error("failet to save chat", "error", err.Error())
		return err
	}
	chat.Event.Raise(app.mediator)

	go func(f domain.From, log logger.Logger) {
		ctx, cancel := pkgCtx.GenContextWithTimeout(3 * time.Minute)
		defer cancel()
		helper := logger.NewHelper(log).WithContext(ctx)

		chat, err := app.repo.Get(ctx, f)
		if err != nil {
			helper.Error("failed to get chat from repository to call chatgpt api", "error", err.Error())
			return
		}
		if _, err := app.api.Chat(ctx, chat); err != nil {
			helper.Error("failed to get answer from chatgpt", "error", err.Error())
			return
		}
		if err := app.repo.Save(ctx, chat); err != nil {
			helper.Error("failed to save chat after get answer", "error", err.Error())
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
