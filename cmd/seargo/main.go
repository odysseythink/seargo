package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seargo/seargo/internal/autocomplete"
	"github.com/seargo/seargo/internal/bangs"
	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/storage"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/search"
	"github.com/seargo/seargo/internal/server"
	"github.com/seargo/seargo/pkg/models"

	// Import engines to trigger init() registration
	_ "github.com/seargo/seargo/engines/bing"
	_ "github.com/seargo/seargo/engines/brave"
	_ "github.com/seargo/seargo/engines/duckduckgo"
	_ "github.com/seargo/seargo/engines/google"
	_ "github.com/seargo/seargo/engines/wikipedia"
	_ "github.com/seargo/seargo/engines/yahoo"
)

func convertTTLByCategory(raw map[string]int) map[models.Category]int {
	if raw == nil {
		return nil
	}
	out := make(map[models.Category]int, len(raw))
	for k, v := range raw {
		out[models.Category(k)] = v
	}
	return out
}

func loadEngineTraits(path string) engine.EngineTraitsMap {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warn("Engine traits file not found, continuing without traits", "path", path)
		return nil
	}
	var traits engine.EngineTraitsMap
	if err := json.Unmarshal(data, &traits); err != nil {
		logger.Warn("Failed to parse engine traits, continuing without traits", "error", err)
		return nil
	}
	logger.Info("Loaded engine traits", "engines", len(traits))
	return traits
}

func main() {
	configPath := flag.String("config", "configs/settings.yml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init("info", "stdout"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting SearGo", "config", *configPath, "port", cfg.Server.Port)

	// Init shared storage
	sharedStorage, err := storage.NewFromConfig(cfg)
	if err != nil {
		logger.Error("Failed to init shared storage", "error", err)
		os.Exit(1)
	}
	defer sharedStorage.Close()

	// Init cache
	c, err := cache.NewMultiLevel(sharedStorage, cache.Config{
		Enabled:       cfg.Cache.Enabled,
		LocalTTL:      cfg.Cache.LocalTTL,
		RemoteTTL:     cfg.Cache.RemoteTTL,
		TTLByCategory: convertTTLByCategory(cfg.Cache.TTLByCategory),
	})
	if err != nil {
		logger.Error("Failed to init cache", "error", err)
		os.Exit(1)
	}

	// Create network registry
	registry, err := httpx.NewRegistry(cfg)
	if err != nil {
		logger.Error("Failed to init network registry", "error", err)
		os.Exit(1)
	}

	// Create default HTTP client bound to registry
	httpClient := httpx.NewClient(
		registry,
		"", // networkName empty → resolved by engine name
		"", // engineName empty → per-engine client created inside Scheduler
		cfg.Outgoing.UserAgent,
		time.Duration(cfg.Outgoing.RequestTimeout)*time.Second,
	)

	// Load engine traits
	traits := loadEngineTraits("data/engine_traits.json")

	// Build EngineInitConfigs from config
	initConfigs := make([]engine.EngineInitConfig, 0, len(cfg.Engines))
	for _, ec := range cfg.Engines {
		if ec.Disabled {
			continue
		}
		cfgCategories := make([]models.Category, len(ec.Categories))
		for i, c := range ec.Categories {
			cfgCategories[i] = models.Category(c)
		}
		initConfigs = append(initConfigs, engine.EngineInitConfig{
			Name:                ec.Name,
			Shortcut:            ec.Shortcut,
			Categories:          cfgCategories,
			Timeout:             ec.Timeout,
			Extra:               ec.Extra,
			Paging:              ec.Paging,
			TimeRangeSupport:    ec.TimeRangeSupport,
			LanguageSupport:     ec.LanguageSupport,
			SafeSearch:          ec.SafeSearch,
			Weight:              ec.Weight,
			DisplayErrorMsgs:    ec.DisplayErrorMessages,
			EnableHTTP:          ec.EnableHTTP,
			Inactive:            ec.Inactive,
			Disabled:            ec.Disabled,
			Tokens:              ec.Tokens,
			Network:             ec.Network,
			SoftMaxRedirects:    ec.SoftMaxRedirects,
			NoResultForHTTPStatus: ec.NoResultForHTTPStatus,
			RaiseForHTTPError:   ec.RaiseForHTTPError,
		})
	}

	// Initialize engines via Loader
	loader := engine.NewLoader(traits)
	loadResult, err := loader.Load(context.Background(), initConfigs)
	if err != nil {
		logger.Error("Failed to load engines", "error", err)
		os.Exit(1)
	}
	logger.Info("Engines loaded", "categories", len(loadResult.Categories), "shortcuts", len(loadResult.Shortcuts))

	// Load bangs trie
	bangTrie, err := bangs.NewBangTrie()
	if err != nil {
		logger.Warn("failed to load external bangs database, bangs disabled", "error", err)
		bangTrie = nil
	}

	// Create autocomplete service and register providers
	acCache := autocomplete.NewResultCache(sharedStorage.WithNamespace("autocomplete"), autocomplete.DefaultCacheTTL)
	defer acCache.Close()
	acSvc := autocomplete.NewService(httpClient, acCache)
	autocomplete.Register("google", autocomplete.NewGoogleProvider(httpClient))
	autocomplete.Register("bing", autocomplete.NewBingProvider(httpClient))
	autocomplete.Register("duckduckgo", autocomplete.NewDuckDuckGoProvider(httpClient))
	autocomplete.Register("brave", autocomplete.NewBraveProvider(httpClient))
	autocomplete.Register("qwant", autocomplete.NewQwantProvider(httpClient))
	autocomplete.Register("startpage", autocomplete.NewStartpageProvider(httpClient))
	autocomplete.Register("wikipedia", autocomplete.NewWikipediaProvider(httpClient))
	autocomplete.Register("dbpedia", autocomplete.NewDBpediaProvider(httpClient))
	autocomplete.Register("swisscows", autocomplete.NewSwisscowsProvider(httpClient))
	autocomplete.Register("baidu", autocomplete.NewBaiduProvider(httpClient))
	autocomplete.Register("360search", autocomplete.NewQihoo360Provider(httpClient))
	autocomplete.Register("naver", autocomplete.NewNaverProvider(httpClient))
	autocomplete.Register("yandex", autocomplete.NewYandexProvider(httpClient))
	autocomplete.Register("seznam", autocomplete.NewSeznamProvider(httpClient))
	autocomplete.Register("sogou", autocomplete.NewSogouProvider(httpClient))
	autocomplete.Register("mwmbl", autocomplete.NewMwmblProvider(httpClient))
	autocomplete.Register("privacywall", autocomplete.NewPrivacyWallProvider(httpClient))
	autocomplete.Register("quark", autocomplete.NewQuarkProvider(httpClient))

	// Create rate limiter for /api/autocomplete (30 req/min/IP)
	rateLimiter := server.NewRateLimiter(server.DefaultRateLimit, time.Minute)
	defer rateLimiter.Close()

	// Init scheduler (handles engine registration internally)
	sched, err := search.NewScheduler(cfg, c, httpClient, nil, nil, bangTrie)
	if err != nil {
		logger.Error("Failed to init scheduler", "error", err)
		os.Exit(1)
	}

	// Create server
	srv := server.New(cfg, sched, acSvc, bangTrie, rateLimiter)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
