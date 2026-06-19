package httpx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAgentPool_Random(t *testing.T) {
	pool := &UserAgentPool{
		OSes:     []string{"Windows NT 10.0; Win64; x64", "X11; Linux x86_64"},
		Template: "Mozilla/5.0 ({os}; rv:{version}) Gecko/20100101 Firefox/{version}",
		Versions: []string{"151.0", "150.0"},
	}

	ua := pool.Random()
	assert.Contains(t, ua, "Mozilla/5.0")
	assert.Contains(t, ua, "Firefox/")
	assert.Contains(t, ua, "rv:")
	assert.NotContains(t, ua, "{os}")
	assert.NotContains(t, ua, "{version}")
}

func TestUserAgentPool_Random_Variation(t *testing.T) {
	pool := &UserAgentPool{
		OSes:     []string{"Windows NT 10.0; Win64; x64", "X11; Linux x86_64", "Macintosh; Intel Mac OS X 10.15"},
		Template: "Mozilla/5.0 ({os}; rv:{version}) Gecko/20100101 Firefox/{version}",
		Versions: []string{"151.0", "150.0", "149.0", "148.0"},
	}

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		seen[pool.Random()] = true
	}
	assert.Greater(t, len(seen), 1, "random should produce varied UAs")
}

func TestNewUserAgentPool_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "useragents.json")
	content := `{"os":["X11; Linux x86_64"],"ua":"Mozilla/5.0 ({os}; rv:{version}) Gecko/20100101 Firefox/{version}","versions":["100.0"]}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	pool, err := NewUserAgentPool(path)
	require.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, 1, len(pool.OSes))
	assert.Equal(t, "X11; Linux x86_64", pool.OSes[0])
}

func TestNewUserAgentPool_Fallback(t *testing.T) {
	pool, err := NewUserAgentPool("/nonexistent/path.json")
	require.NoError(t, err)
	assert.NotNil(t, pool)
	ua := pool.Random()
	assert.NotEmpty(t, ua)
}

func TestNewUserAgentPool_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	_, err := NewUserAgentPool(path)
	assert.Error(t, err)
}

func TestUserAgentPool_Reload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ua.json")
	content1 := `{"os":["A"],"ua":"{os}/{version}","versions":["1"]}`
	require.NoError(t, os.WriteFile(path, []byte(content1), 0644))

	pool, err := NewUserAgentPool(path)
	require.NoError(t, err)
	assert.Equal(t, "A/1", pool.Random())

	content2 := `{"os":["B"],"ua":"{os}-{version}","versions":["2"]}`
	require.NoError(t, os.WriteFile(path, []byte(content2), 0644))

	err = pool.Reload(path)
	require.NoError(t, err)
	assert.Equal(t, "B-2", pool.Random())
}
