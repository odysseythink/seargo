package botdetection

import (
	"context"
	"net/http"
	"strings"
)

type acceptProbe struct{}

func (p *acceptProbe) Name() string { return "http_accept" }

func (p *acceptProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, _ string) (Decision, error) {
	accept := req.Header.Get("Accept")
	if !strings.Contains(accept, "text/html") {
		return Block, nil
	}
	return Allow, nil
}

type acceptEncodingProbe struct{}

func (p *acceptEncodingProbe) Name() string { return "http_accept_encoding" }

func (p *acceptEncodingProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, _ string) (Decision, error) {
	enc := req.Header.Get("Accept-Encoding")
	if !strings.Contains(enc, "gzip") && !strings.Contains(enc, "deflate") {
		return Block, nil
	}
	return Allow, nil
}

type acceptLanguageProbe struct{}

func (p *acceptLanguageProbe) Name() string { return "http_accept_language" }

func (p *acceptLanguageProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, _ string) (Decision, error) {
	lang := req.Header.Get("Accept-Language")
	if lang == "" {
		return Block, nil
	}
	return Allow, nil
}
