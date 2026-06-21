package botdetection

import (
	"context"
	"net"
	"net/http"
)

// Decision is the result of a probe.
type Decision int

const (
	Allow    Decision = iota
	Block
	Redirect
)

func (d Decision) String() string {
	switch d {
	case Allow:
		return "allow"
	case Block:
		return "block"
	case Redirect:
		return "redirect"
	default:
		return "unknown"
	}
}

// Probe inspects a single request aspect.
type Probe interface {
	Name() string
	Filter(ctx context.Context, req *http.Request, cfg *Config, clientIP string) (Decision, error)
}

// State exposes shared state needed by probes.
type State interface {
	IsSuspicious(ctx context.Context, network *net.IPNet, acceptLanguage, userAgent string) (bool, error)
}
