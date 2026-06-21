package limiter

import (
	"context"
	"fmt"
	"net"

	"github.com/seargo/seargo/internal/storage"
)

// linkToken manages link-token challenges.
type linkToken struct {
	kv storage.KV
}

func newLinkToken(kv storage.KV) *linkToken {
	return &linkToken{kv: kv}
}

// Generate creates a new link token.
func (t *linkToken) Generate(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// Ping validates a link token.
func (t *linkToken) Ping(ctx context.Context, token string) error {
	return fmt.Errorf("not implemented")
}

// IsSuspicious reports whether the network is suspicious.
func (t *linkToken) IsSuspicious(ctx context.Context, network *net.IPNet, acceptLanguage, userAgent string) (bool, error) {
	return false, nil
}
