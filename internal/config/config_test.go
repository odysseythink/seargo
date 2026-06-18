package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	cfg, err := Load("../../configs/settings.yml")
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "general", cfg.Search.DefaultCategory)
}

func TestValidate(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
		Search: SearchConfig{MaxResults: 0},
	}
	err := cfg.Validate()
	require.NoError(t, err)
	assert.Equal(t, 10, cfg.Search.MaxResults)
}

func TestEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")
	content := `
server:
  port: 8080
search:
  max_results: 10
engines:
  - name: google
    engine: google
    enabled: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	os.Setenv("SEARGO_SERVER_SECRET_KEY", "my-secret")
	defer os.Unsetenv("SEARGO_SERVER_SECRET_KEY")

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "my-secret", cfg.Server.SecretKey)
}

func TestLayeredLoading(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "user.yml")
	content := `
general:
  instance_name: "MyInstance"
server:
  port: 9090
search:
  safe_search: 2
  default_lang: "en"
engines:
  - name: google
    engine: google
    categories: [general]
    weight: 1.0
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// User value wins
	assert.Equal(t, "MyInstance", cfg.General.InstanceName)
	// Default preserved (user did not set debug)
	assert.Equal(t, false, cfg.General.Debug)
	// User value wins
	assert.Equal(t, 9090, cfg.Server.Port)
	// Default preserved
	assert.Equal(t, "127.0.0.1", cfg.Server.BindAddress)
	// User value wins
	assert.Equal(t, 2, cfg.Search.SafeSearch)
	// Default preserved
	assert.Equal(t, 10, cfg.Search.MaxResults)
	// User engines replace defaults
	assert.Len(t, cfg.Engines, 1)
	assert.Equal(t, "google", cfg.Engines[0].Engine)
}
