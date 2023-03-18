package user

import (
	"github.com/go-jimu/components/mediator"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/httpsrv"
	"github.com/jacexh/chatgpt-bot/internal/business/user/application"
	"github.com/jacexh/chatgpt-bot/internal/business/user/application/transport"
	"github.com/jacexh/chatgpt-bot/internal/business/user/infrastructure"
	"github.com/jmoiron/sqlx"
)

func Init(m mediator.Mediator, db *sqlx.DB, g httpsrv.HTTPServer) {
	repo := infrastructure.NewRepository(db)
	read := infrastructure.NewQueryRepository(db)
	app := application.NewApplication(m, repo, read)
	controller := transport.NewController(app)
	g.With(controller)
}
