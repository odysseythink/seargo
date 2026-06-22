package wikipedia

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func TestMain(m *testing.M) {
	_ = flag.Set("logtostderr", "true")
	os.Exit(m.Run())
}

func TestWikipediaEngine_Basic(t *testing.T) {
	w := &Wikipedia{}
	ok := w.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "wikipedia", w.Name())
	assert.Contains(t, w.Categories(), models.CategoryGeneral)
}

func TestWikipediaEngine_ArticleWithInfobox(t *testing.T) {
	html := `<!doctype html>
<html>
<body>
  <h1 id="firstHeading">Berlin</h1>
  <div class="shortdescription">Capital of Germany</div>
  <table class="infobox">
    <tr><th>Country</th><td><a href="/wiki/Germany">Germany</a></td></tr>
    <tr><th>Area</th><td>891.85 km<sup>2</sup></td></tr>
    <tr><th>Population</th><td>3,669,495</td></tr>
    <tr><td colspan="2"><img src="//upload.wikimedia.org/wikipedia/commons/thumb/Berlin.jpg" alt="Berlin"></td></tr>
  </table>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	origSearchURL := wikipediaSearchURL
	wikipediaSearchURL = server.URL + "?search=%[2]s"
	defer func() { wikipediaSearchURL = origSearchURL }()

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
	defer reg.Close()

	w := &Wikipedia{}
	ok := w.Setup(engine.EngineInitConfig{Client: httpx.NewClient(reg, "", "wikipedia", "SearGoTest/1.0", 0)})
	require.True(t, ok)

	resp, err := w.Search(context.Background(), &models.Request{
		Query:    "Berlin",
		Category: models.CategoryGeneral,
		Language: "en",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	var infobox *models.Result
	for i := range resp.Results {
		if resp.Results[i].Kind == "infobox" {
			infobox = &resp.Results[i]
			break
		}
	}
	require.NotNil(t, infobox, "expected an infobox result")
	assert.Equal(t, "Berlin", infobox.Title)
	assert.Equal(t, "Capital of Germany", infobox.Content)
	assert.Equal(t, "wikipedia", infobox.Engine)
	require.NotNil(t, infobox.Extra)

	attrs := infobox.Extra["attributes"].([]results.InfoboxAttribute)
	require.Len(t, attrs, 3)
	assert.Equal(t, "Country", attrs[0].Label)
	assert.Equal(t, "Germany", attrs[0].Value)
	assert.Equal(t, "https://en.wikipedia.org/wiki/Germany", attrs[0].URL)
	assert.Equal(t, "Area", attrs[1].Label)
	assert.Equal(t, "Population", attrs[2].Label)

	assert.Contains(t, infobox.Extra["img_src"], "upload.wikimedia.org/wikipedia/commons/thumb/Berlin.jpg")

	urls := infobox.Extra["urls"].([]results.InfoboxURL)
	require.Len(t, urls, 1)
	assert.Equal(t, "Wikipedia", urls[0].Title)
}
