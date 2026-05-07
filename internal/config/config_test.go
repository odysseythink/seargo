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
    enabled: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	os.Setenv("SEARGO_SERVER_SECRET_KEY", "my-secret")
	defer os.Unsetenv("SEARGO_SERVER_SECRET_KEY")

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "my-secret", cfg.Server.SecretKey)
}
