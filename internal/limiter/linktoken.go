package limiter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/seargo/seargo/internal/storage"
)

const tokenTTL = 10 * time.Minute
const pingTTL = 30 * time.Minute

// linkToken manages link-token challenges.
type linkToken struct {
	kv storage.KV
}

func newLinkToken(kv storage.KV) *linkToken {
	return &linkToken{kv: kv}
}

// Generate creates a new link token and stores it in KV.
func (t *linkToken) Generate(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	// SetNX ensures atomic creation
	set, err := t.kv.SetNX(ctx, "token:"+token, []byte("1"), tokenTTL)
	if err != nil {
		return "", err
	}
	if !set {
		// Collision — retry (very unlikely with 256-bit random)
		return t.Generate(ctx)
	}
	return token, nil
}

// Ping validates a token and extends its TTL.
func (t *linkToken) Ping(ctx context.Context, token string) error {
	key := "token:" + token
	val, ok, err := t.kv.Get(ctx, key)
	if err != nil {
		return err
	}
	if !ok || val == nil {
		return fmt.Errorf("invalid token")
	}

	// Extend TTL
	if err := t.kv.Expire(ctx, key, tokenTTL); err != nil {
		return err
	}
	return nil
}

// IsSuspicious reports whether the network has not recently pinged a link token.
func (t *linkToken) IsSuspicious(ctx context.Context, network *net.IPNet, acceptLanguage, userAgent string) (bool, error) {
	if network == nil {
		return true, nil
	}

	// Check if any ping exists for this network
	networkStr := network.String()
	pingKey := "ping:" + networkStr

	val, ok, err := t.kv.Get(ctx, pingKey)
	if err != nil {
		return true, nil // fail-closed
	}
	if !ok || val == nil {
		return true, nil // no ping = suspicious
	}
	return false, nil
}
