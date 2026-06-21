package limiter

import (
	"context"
	"fmt"
	"net"

	"github.com/seargo/seargo/internal/security"
	"github.com/seargo/seargo/internal/storage"
)

// Limiter is the rate limiter interface.
type Limiter interface {
	Allow(ctx context.Context, clientIP string, isAPI bool) (bool, string, error)
	Token(ctx context.Context) (string, error)
	Ping(ctx context.Context, token string) error
	IsSuspicious(ctx context.Context, network *net.IPNet, acceptLanguage, userAgent string) (bool, error)
	DropCounter(key string)
}

type limiterSvc struct {
	cfg    *Config
	cnt    *counter
	kv     storage.KV
	tokens *linkToken
}

// New creates a new rate limiter.
func New(cfg *Config, kv storage.KV) Limiter {
	return &limiterSvc{
		cfg:    cfg,
		cnt:    newCounter(kv),
		kv:     kv,
		tokens: newLinkToken(kv),
	}
}

func (l *limiterSvc) DropCounter(key string) {
	ctx := context.Background()
	l.cnt.Drop(ctx, key)
}

func (l *limiterSvc) Allow(ctx context.Context, clientIP string, isAPI bool) (bool, string, error) {
	// Check link-local exemption
	if l.cfg.FilterLinkLocal {
		ip := net.ParseIP(clientIP)
		if ip != nil && security.IsLinkLocal(ip) {
			return true, "", nil
		}
	}

	// API rate limit (non-HTML requests)
	if isAPI && l.cfg.APIMax > 0 {
		val, ok, err := l.cnt.Incr(ctx, "api:"+clientIP, l.cfg.APIWindow, l.cfg.APIMax)
		if err != nil {
			return true, "", nil
		}
		if !ok {
			return false, "api", nil
		}
		_ = val
	}

	// Burst window
	burstKey := fmt.Sprintf("burst:%s", clientIP)
	burstVal, burstOK, err := l.cnt.Incr(ctx, burstKey, l.cfg.BurstWindow, l.cfg.BurstMax)
	if err != nil {
		return true, "", nil
	}
	if !burstOK {
		_ = burstVal
		return false, "burst", nil
	}

	// Long window
	longKey := fmt.Sprintf("long:%s", clientIP)
	longVal, longOK, err := l.cnt.Incr(ctx, longKey, l.cfg.LongWindow, l.cfg.LongMax)
	if err != nil {
		return true, "", nil
	}
	if !longOK {
		_ = longVal
		return false, "long", nil
	}

	return true, "", nil
}

// Token generates a new link token.
func (l *limiterSvc) Token(ctx context.Context) (string, error) {
	return l.tokens.Generate(ctx)
}

// Ping validates a link token.
func (l *limiterSvc) Ping(ctx context.Context, token string) error {
	return l.tokens.Ping(ctx, token)
}

// IsSuspicious implements botdetection.State.
func (l *limiterSvc) IsSuspicious(ctx context.Context, network *net.IPNet, acceptLanguage, userAgent string) (bool, error) {
	return l.tokens.IsSuspicious(ctx, network, acceptLanguage, userAgent)
}
