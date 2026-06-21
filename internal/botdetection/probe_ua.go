package botdetection

import (
	"context"
	"net/http"
	"regexp"
)

type userAgentProbe struct {
	patterns []*regexp.Regexp
}

func newUserAgentProbe(patterns []string) *userAgentProbe {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		compiled = append(compiled, re)
	}
	return &userAgentProbe{patterns: compiled}
}

func (p *userAgentProbe) Name() string { return "http_user_agent" }

func (p *userAgentProbe) Filter(ctx context.Context, req *http.Request, cfg *Config, _ string) (Decision, error) {
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		return Block, nil
	}

	for _, re := range p.patterns {
		if re.MatchString(ua) {
			return Block, nil
		}
	}
	return Allow, nil
}
