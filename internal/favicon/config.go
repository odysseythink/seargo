package favicon

import (
	"time"

	"github.com/seargo/seargo/internal/config"
)

// Config holds favicon service settings.
type Config struct {
	Proxy ProxyConfig
	Cache CacheConfig
}

type ProxyConfig struct {
	MaxAge          time.Duration
	ResolverTimeout time.Duration
	FaviconPath     string
}

type CacheConfig struct {
	HoldTime     time.Duration
	BlobMaxBytes int
}

// LoadConfig loads favicon configuration from a favicons.toml file.
func LoadConfig(path string) (*Config, error) {
	tomlCfg, err := config.LoadFaviconConfig(path)
	if err != nil {
		return nil, err
	}
	return fromFaviconTOML(tomlCfg), nil
}

func fromFaviconTOML(toml *config.FaviconTOMLConfig) *Config {
	maxAge, _ := time.ParseDuration(toml.Proxy.MaxAge)
	timeout, _ := time.ParseDuration(toml.Proxy.ResolverTimeout)
	hold, _ := time.ParseDuration(toml.Cache.HoldTime)

	return &Config{
		Proxy: ProxyConfig{
			MaxAge:          maxAge,
			ResolverTimeout: timeout,
			FaviconPath:     toml.Proxy.FaviconPath,
		},
		Cache: CacheConfig{
			HoldTime:     hold,
			BlobMaxBytes: toml.Cache.BlobMaxBytes,
		},
	}
}
