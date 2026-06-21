package botdetection

import (
	"context"
	"net/http"
)

type secFetchProbe struct{}

func (p *secFetchProbe) Name() string { return "http_sec_fetch" }

func (p *secFetchProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, _ string) (Decision, error) {
	if req.TLS == nil && req.Header.Get("X-Forwarded-Proto") != "https" {
		return Allow, nil
	}

	mode := req.Header.Get("Sec-Fetch-Mode")
	site := req.Header.Get("Sec-Fetch-Site")
	dest := req.Header.Get("Sec-Fetch-Dest")

	validModes := map[string]bool{"navigate": true, "cors": true}
	validSites := map[string]bool{"same-origin": true, "same-site": true, "none": true}
	validDests := map[string]bool{"document": true, "empty": true}

	if mode != "" && !validModes[mode] {
		return Redirect, nil
	}
	if site != "" && !validSites[site] {
		return Redirect, nil
	}
	if dest != "" && !validDests[dest] {
		return Redirect, nil
	}
	return Allow, nil
}
