package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/autocomplete"
	"github.com/seargo/seargo/internal/bangs"
	"github.com/seargo/seargo/internal/botdetection"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/favicon"
	"github.com/seargo/seargo/internal/imageproxy"
	"github.com/seargo/seargo/internal/limiter"
	"github.com/seargo/seargo/internal/middleware"
	"github.com/seargo/seargo/internal/search"
	"github.com/seargo/seargo/internal/security"
)

type Server struct {
	router       *gin.Engine
	config       *config.Config
	scheduler    *search.Scheduler
	autocomplete *autocomplete.Service
	bangsService *bangs.BangTrie
	rateLimiter  *RateLimiter
	http         *http.Server

	// Phase 8 services
	botDetector *botdetection.Detector
	limiterSvc  limiter.Limiter
	imageProxy  imageproxy.Proxy
	favSvc      *favicon.Service
}

func New(cfg *config.Config, scheduler *search.Scheduler,
	ac *autocomplete.Service, bs *bangs.BangTrie, rl *RateLimiter,
	botDetector *botdetection.Detector, limiterSvc limiter.Limiter,
	imageProxy imageproxy.Proxy, favSvc *favicon.Service) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 1. TrustedProxy — must be first (all downstream middleware read clientIP from context)
	trustedProxyList, err := security.ParseProxyList(cfg.Server.TrustedProxies)
	if err != nil {
		panic(fmt.Sprintf("failed to parse trusted_proxies: %v", err))
	}
	extractor := security.NewIPExtractor(trustedProxyList)
	r.Use(middleware.TrustedProxy(extractor))

	// 2. Recovery
	r.Use(middleware.Recovery())

	// 3. RequestLogger
	r.Use(middleware.RequestLogger())

	// 4. BotDetection (optional)
	if botDetector != nil {
		r.Use(middleware.BotDetection(botDetector))
	}

	// 5. Limiter (optional, controlled by cfg.Server.Limiter)
	if limiterSvc != nil {
		r.Use(middleware.Limiter(cfg, limiterSvc))
	}

	// 6. SecurityHeaders
	r.Use(middleware.SecurityHeaders(cfg.Server.DefaultHTTPHeaders))

	// 7. ErrorHandler
	r.Use(middleware.ErrorHandler())

	s := &Server{
		router:       r,
		config:       cfg,
		scheduler:    scheduler,
		autocomplete: ac,
		bangsService: bs,
		rateLimiter:  rl,
		botDetector:  botDetector,
		limiterSvc:   limiterSvc,
		imageProxy:   imageProxy,
		favSvc:       favSvc,
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
