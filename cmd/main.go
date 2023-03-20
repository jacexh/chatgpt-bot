package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/go-jimu/components/logger"
	"github.com/go-jimu/components/mediator"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/gpt"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/httpsrv"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/mysql"
	"github.com/jacexh/chatgpt-bot/internal/bootstrap/telegram"
	"github.com/jacexh/chatgpt-bot/internal/chat"
	"github.com/jacexh/chatgpt-bot/internal/pkg/context"
	"github.com/jacexh/chatgpt-bot/internal/pkg/eventbus"
	"github.com/jacexh/chatgpt-bot/internal/pkg/log"
	"github.com/jacexh/chatgpt-bot/internal/pkg/option"
)

type Option struct {
	Logger     log.Option      `json:"logger" toml:"logger" yaml:"logger"`
	Context    context.Option  `json:"context" toml:"context" yaml:"context"`
	MySQL      mysql.Option    `json:"mysql" toml:"mysql" yaml:"mysql"`
	HTTPServer httpsrv.Option  `json:"http-server" toml:"http-server" yaml:"http-server"`
	Telegram   telegram.Option `json:"telegram" yaml:"telegram"`
	ChatGPT    gpt.Option      `json:"chatgpt" yaml:"chatgpt"`
}

func main() {
	opt := new(Option)
	conf := option.Load()
	if err := conf.Scan(opt); err != nil {
		panic(err)
	}

	// pkg layer
	log := log.NewLog(opt.Logger).(*logger.Helper)
	log.Info("loaded configurations", "option", *opt)

	context.New(opt.Context)

	// eventbus layer
	eb := mediator.NewInMemMediator(1)
	eventbus.SetDefault(eb)

	// driver layer
	db := mysql.NewMySQLDriver(opt.MySQL)
	cg := httpsrv.NewHTTPServer(opt.HTTPServer, log)
	bot := telegram.NewBotAPI(opt.Telegram, log)
	gpt := gpt.NewChatGPT(opt.ChatGPT)

	// each business layer
	chat.Init(log, db, cg, eb, bot, gpt)

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.RootContext(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	if err := cg.Serve(ctx); err != nil {
		log.Error("failed to shutdown http server", "error", err.Error())
	}
	log.Warnf("kill all available contexts in %s", opt.Context.ShutdownTimeout)
	context.KillContextAfterTimeout()
	log.Info("bye")
	os.Exit(0)
}
