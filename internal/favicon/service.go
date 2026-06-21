package favicon

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/seargo/seargo/internal/security"
	"github.com/seargo/seargo/internal/storage"
)

// Service handles favicon proxy with caching.
type Service struct {
	cfg    Config
	signer security.HMACSigner
	kv     storage.KV
}

// New creates a new favicon service.
func New(cfg Config, signer security.HMACSigner, kv storage.KV) Service {
	return Service{cfg: cfg, signer: signer, kv: kv}
}

// SearchFavicon tries to resolve a favicon for the given authority via the named resolver.
// It caches both positive and negative results.
func (s *Service) SearchFavicon(ctx context.Context, resolver, authority string) ([]byte, string, error) {
	cacheKey := fmt.Sprintf("map:%s:%s", resolver, authority)

	// Check cache
	if s.kv != nil {
		hash, ok, err := s.kv.Get(ctx, cacheKey)
		if err == nil && ok && len(hash) > 0 {
			hashStr := string(hash)
			if hashStr == "-" { // negative cache
				return nil, "", fmt.Errorf("cached negative: %s", authority)
			}
			if hashStr != "" {
				blobKey := "blob:" + hashStr
				raw, ok2, err2 := s.kv.Get(ctx, blobKey)
				if err2 == nil && ok2 && len(raw) > 0 {
					parts := bytes.SplitN(raw, []byte("\n"), 2)
					if len(parts) == 2 {
						mime := string(parts[0])
						return parts[1], mime, nil
					}
				}
			}
		}
	}

	// Call resolver
	fn, err := GetResolver(resolver)
	if err != nil {
		return nil, "", err
	}
	data, mime, err := fn(ctx, authority)
	if err != nil {
		// Negative cache
		if s.kv != nil {
			s.kv.Set(ctx, cacheKey, []byte("-"), s.cfg.Cache.HoldTime)
		}
		return nil, "", err
	}

	if data == nil {
		if s.kv != nil {
			s.kv.Set(ctx, cacheKey, []byte("-"), s.cfg.Cache.HoldTime)
		}
		return nil, "", fmt.Errorf("no favicon data for %s", authority)
	}

	// Check blob size before caching
	tooBig := len(data) > s.cfg.Cache.BlobMaxBytes
	if !tooBig && s.kv != nil {
		raw := append([]byte(mime+"\n"), data...)
		hash := sha256.Sum256(raw)
		hashStr := hex.EncodeToString(hash[:])
		s.kv.Set(ctx, "blob:"+hashStr, raw, s.cfg.Cache.HoldTime)
		s.kv.Set(ctx, cacheKey, []byte(hashStr), s.cfg.Cache.HoldTime)
	}

	return data, mime, nil
}

// SignedURL creates an HMAC-signed favicon proxy URL.
func (s *Service) SignedURL(resolver, authority string) (string, error) {
	if s.signer == nil {
		return fmt.Sprintf("/favicon_proxy?resolver=%s&authority=%s", resolver, authority), nil
	}
	raw := fmt.Sprintf("%s::%s", resolver, authority)
	sig := s.signer.Sign([]byte(raw))
	return fmt.Sprintf("/favicon_proxy?resolver=%s&authority=%s&h=%s", resolver, authority, sig), nil
}

// RewriteFaviconURL rewrites a favicon URL through the favicon proxy.
func (s *Service) RewriteFaviconURL(pageURL, faviconURL string) string {
	if faviconURL == "" {
		return faviconURL
	}
	if faviconURL == pageURL || faviconURL == "" {
		return faviconURL
	}
	return faviconURL
}

// Serve serves a favicon through the proxy.
func (s *Service) Serve(ctx context.Context, resolver, authority, signature string) ([]byte, string, error) {
	if s.signer != nil {
		raw := fmt.Sprintf("%s::%s", resolver, authority)
		if !s.signer.Verify([]byte(raw), signature) {
			return nil, "", fmt.Errorf("invalid signature")
		}
	}

	if resolver == "" || authority == "" {
		return nil, "", fmt.Errorf("missing resolver or authority")
	}

	data, mime, err := s.SearchFavicon(ctx, resolver, authority)
	if err != nil {
		return nil, "", err
	}
	return data, mime, nil
}
