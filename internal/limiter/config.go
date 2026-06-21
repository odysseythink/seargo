package limiter

import (
	"time"

	"github.com/seargo/seargo/internal/config"
)

// Config holds rate limiter settings.
type Config struct {
	BurstWindow     time.Duration
	BurstMax        int64
	LongWindow      time.Duration
	LongMax         int64
	APIWindow       time.Duration
	APIMax          int64
	FilterLinkLocal bool
	LinkToken       bool
}

// LoadConfig loads limiter configuration from a limiter.toml file.
func LoadConfig(path string) (*Config, error) {
	tomlCfg, err := config.LoadLimiterConfig(path)
	if err != nil {
		return nil, err
	}

	burstDur, _ := time.ParseDuration(tomlCfg.Windows.BurstDuration)
	longDur, _ := time.ParseDuration(tomlCfg.Windows.LongDuration)

	return &Config{
		BurstWindow:     burstDur,
		BurstMax:        tomlCfg.Windows.BurstMax,
		LongWindow:      longDur,
		LongMax:         tomlCfg.Windows.LongMax,
		APIWindow:       time.Minute,
		APIMax:          30,
		FilterLinkLocal: tomlCfg.IPLimit.FilterLinkLocal,
		LinkToken:       tomlCfg.IPLimit.LinkToken,
	}, nil
}
