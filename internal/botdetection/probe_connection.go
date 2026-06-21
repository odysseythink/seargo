package botdetection

import (
	"context"
	"net/http"
)

type connectionProbe struct{}

func (p *connectionProbe) Name() string { return "http_connection" }

func (p *connectionProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, _ string) (Decision, error) {
	if req.Header.Get("Connection") == "close" {
		return Block, nil
	}
	return Allow, nil
}
