package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/autocomplete"
	"github.com/seargo/seargo/internal/bangs"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/middleware"
	"github.com/seargo/seargo/internal/search"
)

type Server struct {
	router       *gin.Engine
	config       *config.Config
	scheduler    *search.Scheduler
	autocomplete *autocomplete.Service
	bangsService *bangs.BangTrie
	rateLimiter  *RateLimiter
	http         *http.Server
}

func New(cfg *config.Config, scheduler *search.Scheduler,
	ac *autocomplete.Service, bs *bangs.BangTrie, rl *RateLimiter) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorHandler())

	s := &Server{
		router:       r,
		config:       cfg,
		scheduler:    scheduler,
		autocomplete: ac,
		bangsService: bs,
		rateLimiter:  rl,
	}

	s.setupRoutes()

	s.http = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.BindAddress, cfg.Server.Port),
		Handler: r,
	}

	return s
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
