package bases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	logger.Init("warn", "stderr")
}

func TestXPathEngine_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<article class="result">
				<h3><a href="https://example.com/page1">Result One</a></h3>
				<p class="snippet">This is the first result</p>
			</article>
			<article class="result">
				<h3><a href="https://example.com/page2">Result Two</a></h3>
				<p class="snippet">Second result snippet</p>
			</article>
		</body></html>`))
	}))
	defer server.Close()

	eng := NewXPathEngine("test_xpath", []models.Category{models.CategoryGeneral}, XPathConfig{
		SearchURL:    server.URL + "/search?q={query}",
		ResultXPath:  "//article[@class='result']",
		URLXPath:     ".//h3/a/@href",
		TitleXPath:   ".//h3/a",
		ContentXPath: ".//p[@class='snippet']",
	})

	ok := eng.Setup(engine.EngineInitConfig{Name: "test_xpath"})
	assert.True(t, ok)

	// Create a proper Registry with a default network so httpx.Client doesn't panic
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
		},
		Engines: []config.EngineConfig{},
	}
	reg, err := httpx.NewRegistry(cfg)
	require.NoError(t, err)
	client := httpx.NewClient(reg, "", "test_xpath", "test-ua", 0)
	eng.(*xpathEngine).SetClient(client)

	req := &models.Request{Query: "test", Category: models.CategoryGeneral}
	resp, err := eng.Search(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Results, 2)
	assert.Equal(t, "Result One", resp.Results[0].Title)
	assert.Equal(t, "https://example.com/page1", resp.Results[0].URL)
	assert.Equal(t, "Second result snippet", resp.Results[1].Content)
}
