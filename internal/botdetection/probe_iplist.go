package botdetection

import (
	"context"
	"net"
	"net/http"
)

// searxngOrgIPs are IPs from searx.space (public instance monitor).
var searxngOrgIPs = []string{}

type ipListProbe struct{}

func (p *ipListProbe) Name() string { return "ip_lists" }

func (p *ipListProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, clientIP string) (Decision, error) {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return Block, nil
	}

	ipStr := ip.String()

	for _, pass := range cfg.IPLists.PassIP {
		if matchesCIDROrIP(pass, ipStr) {
			return Allow, nil
		}
	}

	if cfg.IPLists.PassSearxngOrg {
		for _, pass := range searxngOrgIPs {
			if pass == ipStr {
				return Allow, nil
			}
		}
	}

	for _, block := range cfg.IPLists.BlockIP {
		if matchesCIDROrIP(block, ipStr) {
			return Block, nil
		}
	}

	return Allow, nil
}

func matchesCIDROrIP(cidrOrIP, ipStr string) bool {
	if cidrOrIP == ipStr {
		return true
	}
	_, network, err := net.ParseCIDR(cidrOrIP)
	if err != nil {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return network.Contains(ip)
}
