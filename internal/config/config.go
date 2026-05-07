package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Search   SearchConfig   `yaml:"search"`
	Engines  []EngineConfig `yaml:"engines"`
	Outgoing OutgoingConfig `yaml:"outgoing"`
	Cache    CacheConfig    `yaml:"cache"`
}

type ServerConfig struct {
	Port        int    `yaml:"port"`
	BindAddress string `yaml:"bind_address"`
	SecretKey   string `yaml:"secret_key"`
}

type SearchConfig struct {
	SafeSearch      int    `yaml:"safe_search"`
	Autocomplete    string `yaml:"autocomplete"`
	DefaultLang     string `yaml:"default_lang"`
	DefaultCategory string `yaml:"default_category"`
	MaxResults      int    `yaml:"max_results"`
}

type EngineConfig struct {
	Name    string                 `yaml:"name"`
	Enabled bool                   `yaml:"enabled"`
	Weight  float64                `yaml:"weight"`
	Timeout int                    `yaml:"timeout"`
	APIKey  string                 `yaml:"api_key"`
	Extra   map[string]interface{} `yaml:"extra"`
}

type OutgoingConfig struct {
	Timeout   int    `yaml:"request_timeout"`
	UserAgent string `yaml:"useragent"`
}

type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	LocalTTL  int    `yaml:"local_ttl"`
	RedisTTL  int    `yaml:"redis_ttl"`
	RedisAddr string `yaml:"redis_addr"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyEnvOverrides(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Search.MaxResults <= 0 {
		c.Search.MaxResults = 10
	}
	return nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SEARGO_SERVER_SECRET_KEY"); v != "" {
		cfg.Server.SecretKey = v
	}
	if v := os.Getenv("SEARGO_CACHE_REDIS_ADDR"); v != "" {
		cfg.Cache.RedisAddr = v
	}
	for i := range cfg.Engines {
		envKey := fmt.Sprintf("SEARGO_ENGINE_%s_API_KEY", strings.ToUpper(cfg.Engines[i].Name))
		if v := os.Getenv(envKey); v != "" {
			cfg.Engines[i].APIKey = v
		}
	}
}
