package botdetection

import (
	"github.com/seargo/seargo/internal/config"
)

// Config holds bot detection settings.
type Config struct {
	IPv4Prefix        int
	IPv6Prefix        int
	IPLimit           IPLimitConfig
	IPLists           IPListsConfig
	LinkToken         LinkTokenConfig
	UserAgentPatterns []string
}

type IPLimitConfig struct {
	FilterLinkLocal bool
	LinkToken       bool
}

type IPListsConfig struct {
	BlockIP        []string
	PassIP         []string
	PassSearxngOrg bool
}

type LinkTokenConfig struct {
	Enabled bool
}

// LoadConfig loads bot detection configuration from a limiter.toml file.
func LoadConfig(path string) (*Config, error) {
	tomlCfg, err := config.LoadLimiterConfig(path)
	if err != nil {
		return nil, err
	}
	return fromLimiterTOML(tomlCfg), nil
}

func fromLimiterTOML(toml *config.LimiterTOMLConfig) *Config {
	return &Config{
		IPv4Prefix: 32,
		IPv6Prefix: 48,
		IPLists: IPListsConfig{
			BlockIP:        toml.IPLists.BlockIP,
			PassIP:         toml.IPLists.PassIP,
			PassSearxngOrg: toml.IPLists.PassSearxngOrg,
		},
		IPLimit: IPLimitConfig{
			FilterLinkLocal: toml.IPLimit.FilterLinkLocal,
			LinkToken:       toml.IPLimit.LinkToken,
		},
		LinkToken: LinkTokenConfig{
			Enabled: toml.IPLimit.LinkToken,
		},
		UserAgentPatterns: []string{
			`^$`,
			`(?i)curl/`,
			`(?i)wget/`,
			`(?i)python-requests/`,
			`(?i)scrapy`,
			`(?i)\bbot\b`,
			`(?i)\bcrawler\b`,
			`(?i)\bspider\b`,
			`(?i)\bheadless\b`,
		},
	}
}
