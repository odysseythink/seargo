package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/search"
	"github.com/seargo/seargo/internal/server"

	// Import engines to trigger init() registration
	_ "github.com/seargo/seargo/engines/bing"
	_ "github.com/seargo/seargo/engines/brave"
	_ "github.com/seargo/seargo/engines/duckduckgo"
	_ "github.com/seargo/seargo/engines/google"
	_ "github.com/seargo/seargo/engines/wikipedia"
	_ "github.com/seargo/seargo/engines/yahoo"
)

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

	// Init cache
	c, err := cache.NewMultiLevel(cfg.Cache.RedisAddr)
	if err != nil {
		logger.Error("Failed to init cache", "error", err)
		os.Exit(1)
	}

	// Create shared HTTP client
	httpClient := httpx.New(
		cfg.Outgoing.UserAgent,
		time.Duration(cfg.Outgoing.RequestTimeout)*time.Second,
	)

	// Init scheduler (handles engine registration internally)
	sched, err := search.NewScheduler(cfg, c, httpClient)
	if err != nil {
		logger.Error("Failed to init scheduler", "error", err)
		os.Exit(1)
	}

	// Create server
	srv := server.New(cfg, sched)

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
