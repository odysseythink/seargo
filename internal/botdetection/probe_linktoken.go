package botdetection

import (
	"context"
	"net"
	"net/http"

	"github.com/seargo/seargo/internal/security"
)

type linkTokenProbe struct {
	state State
}

func newLinkTokenProbe(state State) *linkTokenProbe {
	return &linkTokenProbe{state: state}
}

func (p *linkTokenProbe) Name() string { return "link_token" }

func (p *linkTokenProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, clientIP string) (Decision, error) {
	if !cfg.IPLimit.LinkToken || p.state == nil {
		return Allow, nil
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return Block, nil
	}

	network, err := security.NetworkOf(ip, cfg.IPv4Prefix, cfg.IPv6Prefix)
	if err != nil {
		return Block, nil
	}

	suspicious, err := p.state.IsSuspicious(ctx, network,
		req.Header.Get("Accept-Language"),
		req.Header.Get("User-Agent"))
	if err != nil {
		return Block, nil // fail-closed
	}
	if suspicious {
		return Redirect, nil
	}
	return Allow, nil
}
