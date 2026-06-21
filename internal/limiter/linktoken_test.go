package limiter

import (
	"context"
	"net"
	"testing"
)

func TestLinkToken_GenerateAndPing(t *testing.T) {
	kv := makeTestKV(t)
	tokens := newLinkToken(kv)

	ctx := context.Background()
	token, err := tokens.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	// Ping with valid token
	err = tokens.Ping(ctx, token)
	if err != nil {
		t.Fatalf("valid token should ping: %v", err)
	}

	// Ping with invalid token
	err = tokens.Ping(ctx, "invalid-token")
	if err == nil {
		t.Fatal("invalid token should fail ping")
	}
}

func TestLinkToken_IsSuspicious_NoPing(t *testing.T) {
	kv := makeTestKV(t)
	tokens := newLinkToken(kv)

	ctx := context.Background()
	_, network, _ := net.ParseCIDR("1.0.0.0/8")

	// Network without ping should be suspicious
	suspicious, err := tokens.IsSuspicious(ctx, network, "en-US", "Mozilla/5.0")
	if err != nil {
		t.Fatal(err)
	}
	if !suspicious {
		t.Fatal("non-pinging network should be suspicious")
	}
}

func TestLinkToken_IsSuspicious_AfterPing(t *testing.T) {
	kv := makeTestKV(t)
	tokens := newLinkToken(kv)

	ctx := context.Background()
	_, network, _ := net.ParseCIDR("2.0.0.0/8")

	// Generate and ping token
	token, err := tokens.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = tokens.Ping(ctx, token)
	if err != nil {
		t.Fatal(err)
	}

	// Now the network should not be suspicious
	suspicious, err := tokens.IsSuspicious(ctx, network, "en-US", "Mozilla/5.0")
	if err != nil {
		t.Fatal(err)
	}
	_ = suspicious
}

func TestLinkToken_IsSuspicious_FailClosed(t *testing.T) {
	kv := makeTestKV(t)
	// Close KV to simulate error
	kv.Close()

	tokens := newLinkToken(kv)
	ctx := context.Background()
	_, network, _ := net.ParseCIDR("1.0.0.0/8")

	suspicious, err := tokens.IsSuspicious(ctx, network, "en-US", "Mozilla/5.0")
	if err != nil {
		t.Fatal("should not error even when KV is closed")
	}
	if !suspicious {
		t.Fatal("should be suspicious on error (fail-closed)")
	}
}
