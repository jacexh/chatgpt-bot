package httpsrv

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-jimu/components/logger"
	internalCtx "github.com/jacexh/chatgpt-bot/internal/pkg/context"
)

type (
	Option struct {
		Addr     string `json:"addr" yaml:"addr" toml:"addr"`
		CertFile string `json:"cert_file" toml:"cert_file" yaml:"cert_file"`
		KeyFile  string `json:"key_file" toml:"key_file" yaml:"key_file"`
	}

	API struct {
		Pattern string
		Method  string
		Func    http.HandlerFunc
	}

	MiddlewareScope int

	Middleware struct {
		Middleware func(http.Handler) http.Handler
		Scope      MiddlewareScope
	}

	Controller interface {
		Slug() string
		Middlewares() []Middleware
		APIs() []API
	}

	HTTPServer interface {
		With(Controller)
		Serve(context.Context) error
	}

	router struct {
		router      *chi.Mux
		option      Option
		logger      *logger.Helper
		root        Controller
		controllers []Controller
	}
)

const (
	ScopeController MiddlewareScope = iota // controller 层面
	ScopeGlobal                            // 全局中间件
)

var readTimeout = 3 * time.Second

func NewHTTPServer(opt Option, log logger.Logger, cs ...Controller) HTTPServer {
	g := &router{
		router:      chi.NewRouter(),
		option:      opt,
		logger:      logger.NewHelper(log),
		root:        newRootController(log),
		controllers: make([]Controller, 0),
	}

	for _, controller := range cs {
		g.With(controller)
	}
	return g
}

func (g *router) With(c Controller) {
	g.controllers = append(g.controllers, c)
}

// chi: all middlewares must be defined before routes on a mux
func (g *router) lazyLoad() {
	// apply global middlewares
	if g.root != nil {
		for _, middleware := range g.root.Middlewares() {
			g.router.Use(middleware.Middleware)
		}
	}

	for _, controller := range g.controllers {
		for _, middleware := range controller.Middlewares() {
			if middleware.Scope == ScopeGlobal {
				g.router.Use(middleware.Middleware)
			}
		}
	}

	for _, api := range g.root.APIs() {
		g.router.Method(api.Method, api.Pattern, api.Func)
	}

	// each child controller
	for _, controller := range g.controllers {
		g.router.Route(controller.Slug(), func(r chi.Router) {
			for _, middleware := range controller.Middlewares() {
				if middleware.Scope != ScopeGlobal {
					r.Use(middleware.Middleware)
				}
			}

			for _, api := range controller.APIs() {
				r.Method(api.Method, api.Pattern, api.Func)
			}
		})
	}
}

func (g *router) Serve(ctx context.Context) error {
	g.lazyLoad()

	srv := &http.Server{
		Addr:              g.option.Addr,
		Handler:           g.router,
		ReadHeaderTimeout: readTimeout, // https://cwe.mitre.org/data/definitions/400.html
	}
	internalErr := make(chan error, 1)
	defer close(internalErr)

	go func() {
		var err error
		if g.option.KeyFile != "" && g.option.CertFile != "" {
			err = srv.ListenAndServeTLS(g.option.CertFile, g.option.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if !errors.Is(err, http.ErrServerClosed) {
			internalErr <- err
		}
	}()

	var err error
	select {
	case <-ctx.Done():
		g.logger.Warn("caught quit signal")
	case err = <-internalErr:
		g.logger.Error("an unknown error occurred in http server", "error", err.Error())
	}

	ctx, cancel := internalCtx.GenDefaultContext()
	defer cancel()
	g.logger.Warn("try to shutdown http server")
	return srv.Shutdown(ctx)
}
