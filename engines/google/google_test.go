package google

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/search/processor"
	"github.com/seargo/seargo/pkg/models"
)

func TestMain(m *testing.M) {
	flag.Parse()
	_ = flag.Set("logtostderr", "true")
	code := m.Run()
	os.Exit(code)
}

func TestGoogleEngine(t *testing.T) {
	g := &Google{}
	ok := g.Init(context.Background(), engine.EngineInitConfig{})
	assert.True(t, ok)
	assert.Equal(t, "google", g.Name())
}

func TestGoogleInfo_DefaultLocale(t *testing.T) {
	g := &Google{}
	traits := engine.EngineTraits{
		DataType: "traits_v1",
		Languages: map[string]string{"en": "lang_en", "en-US": "lang_en"},
		Regions:   map[string]string{"en-US": "US"},
		AllLocale: "ZZ",
		Custom: map[string]any{
			"supported_domains": map[string]any{"US": "www.google.com"},
		},
	}

	info := g.googleInfo("en-US", traits, config.GoogleEngineParams{})
	assert.Equal(t, "www.google.com", info.subdomain)
	assert.Equal(t, "en-US", info.params["hl"])
	assert.Equal(t, "lang_en", info.params["lr"])
	assert.Equal(t, "countryUS", info.params["cr"])
	assert.Equal(t, "utf8", info.params["ie"])
	assert.Equal(t, "*/*", info.headers["Accept"])
	assert.Equal(t, "YES+", info.cookies["CONSENT"])
}

func TestGoogleInfo_AllLocale(t *testing.T) {
	g := &Google{}
	traits := engine.EngineTraits{
		DataType:  "traits_v1",
		Languages: map[string]string{"en": "lang_en"},
		Regions:   map[string]string{"en-US": "US"},
		AllLocale: "ZZ",
	}

	info := g.googleInfo("all", traits, config.GoogleEngineParams{})
	assert.Equal(t, "www.google.com", info.subdomain)
	assert.Equal(t, "", info.params["lr"])
	assert.Equal(t, "", info.params["cr"])
	assert.Contains(t, info.params["hl"], "ZZ")
}

func TestGoogleInfo_ConsentCookieOverride(t *testing.T) {
	g := &Google{}
	info := g.googleInfo("all", engine.EngineTraits{AllLocale: "ZZ"}, config.GoogleEngineParams{ConsentCookie: "YES+cb.20210328-17-p0.en+FX+"})
	assert.Equal(t, "YES+cb.20210328-17-p0.en+FX+", info.cookies["CONSENT"])
}

func TestGoogleInfo_ExtraParamsApplied(t *testing.T) {
	g := &Google{}
	traits := engine.EngineTraits{
		DataType:  "traits_v1",
		Languages: map[string]string{"en": "lang_en"},
		Regions:   map[string]string{"en-US": "US"},
		AllLocale: "ZZ",
	}
	info := g.googleInfo("en-US", traits, config.GoogleEngineParams{
		ExtraParams: []string{"safe=off", "hl=de"},
	})
	assert.Equal(t, "off", info.params["safe"], "extra param safe should be applied")
	assert.Equal(t, "de", info.params["hl"], "extra param hl should override default")
}

func TestCaptchaError_ErrorString(t *testing.T) {
	err := &CaptchaError{Msg: "sorry page"}
	assert.Equal(t, "google captcha: sorry page", err.Error())
	assert.Contains(t, err.Error(), "captcha")
}

func TestBuildSearchURL_Pagination(t *testing.T) {
	g := &Google{}
	info := googleInfo{
		subdomain: "www.google.com",
		params: map[string]string{
			"hl": "en-US",
			"lr": "lang_en",
			"ie": "utf8",
			"oe": "utf8",
		},
	}
	u := g.buildSearchURL("golang", 2, "", 0, info)
	assert.Contains(t, u, "q=golang")
	assert.Contains(t, u, "start=10")
	assert.Contains(t, u, "filter=0")
	assert.True(t, strings.HasPrefix(u, "https://www.google.com/search?"))
}

func TestBuildSearchURL_TimeRangeAndSafeSearch(t *testing.T) {
	g := &Google{}
	info := googleInfo{
		subdomain: "www.google.com",
		params:    map[string]string{"hl": "en-US"},
	}
	u := g.buildSearchURL("golang", 1, "year", 2, info)
	assert.Contains(t, u, "tbs=qdr%3Ay")
	assert.Contains(t, u, "safe=high")
}

func TestDetectSorry_CaptchaHost(t *testing.T) {
	g := &Google{}
	assert.True(t, g.detectSorry(&httpx.Response{URL: "https://sorry.google.com/sorry/index", StatusCode: 200, Body: []byte("<html></html>")}))
}

func TestDetectSorry_MetaRefresh(t *testing.T) {
	g := &Google{}
	body := []byte(`<html><meta http-equiv="refresh" content="0;url=/sorry/index?continue=..."></html>`)
	assert.True(t, g.detectSorry(&httpx.Response{URL: "https://www.google.com/sorry", StatusCode: 200, Body: body}))
}

func TestDetectSorry_NormalPage(t *testing.T) {
	g := &Google{}
	body := make([]byte, 2500)
	copy(body, []byte("<html>google results</html>"))
	assert.False(t, g.detectSorry(&httpx.Response{URL: "https://www.google.com/search?q=test", StatusCode: 200, Body: body}))
}

func TestParseResults(t *testing.T) {
	g := &Google{}
	html := `
<html>
<body>
  <a data-ved="abc" href="/url?q=https%3A%2F%2Fgo.dev&sa=Uxyz">
    <div style="">The Go Programming Language</div>
  </a>
  <div class="ilUpNd H66NU aSRlid">Go is an open source programming language.</div>
  <div class="gGQDvd iIWm4b"><a>Suggestion one</a></div>
</body>
</html>`

	results, suggestions := g.parseResults(&httpx.Response{Body: []byte(html)})
	require.Len(t, results, 1)
	assert.Equal(t, "The Go Programming Language", results[0].Title)
	assert.Equal(t, "https://go.dev", results[0].URL)
	assert.Equal(t, "Go is an open source programming language.", results[0].Content)
	assert.Equal(t, "google", results[0].Engine)
	require.Len(t, suggestions, 1)
	assert.Equal(t, "Suggestion one", suggestions[0])
}

func TestParseResults_Thumbnail_DataImage(t *testing.T) {
	g := &Google{}
	html := `
<html>
<body>
<script>var s='data:image/jpeg;base64,ABC';var i='dimg_123';</script>
  <a data-ved="x" href="/url?q=https%3A%2F%2Fexample.com">
    <div style="">Example</div>
    <img id="dimg_123" src="data:image/jpeg;base64,ABC">
  </a>
</body>
</html>`

	results, _ := g.parseResults(&httpx.Response{Body: []byte(html)})
	require.Len(t, results, 1)
	assert.Equal(t, "data:image/jpeg;base64,ABC", results[0].ThumbnailURL)
}

func newGoogleTestClient(t *testing.T, handler http.Handler) (*httpx.Client, *httptest.Server) {
	srv := httptest.NewTLSServer(handler)
	verifyFalse := false
	cfg := &config.Config{
		Outgoing: config.OutgoingConfig{
			RequestTimeout:  3.0,
			PoolConnections: 100,
			PoolMaxsize:     10,
			KeepaliveExpiry: 5.0,
			MaxRedirects:    30,
			EnableHTTP:      true,
			Networks: map[string]config.OutgoingNetworkOverride{
				"google": {Verify: &verifyFalse},
			},
		},
		Engines: []config.EngineConfig{},
	}
	reg, err := httpx.NewRegistry(cfg)
	require.NoError(t, err)
	return httpx.NewClient(reg, "", "google", "", 5*time.Second), srv
}

func TestGoogleSearch_MockHTML(t *testing.T) {
	g := &Google{}

	var requestedPath string
	client, srv := newGoogleTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`
<html>
<body>
  <a data-ved="x" href="/url?q=https%3A%2F%2Fgo.dev&sa=U">
    <div style="">Go</div>
  </a>
  <div class="ilUpNd H66NU aSRlid">The Go Programming Language</div>
  <div class="gGQDvd iIWm4b"><a>go programming language</a></div>
</body>
</html>`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "https://")
	g.client = client
	g.cfg = config.GoogleEngineParams{UseMobileUI: false}
	g.traits = engine.EngineTraits{
		DataType:  "traits_v1",
		Languages: map[string]string{"en": "lang_en", "en-US": "lang_en"},
		Regions:   map[string]string{"en-US": "US"},
		AllLocale: "ZZ",
		Custom: map[string]any{
			"supported_domains": map[string]any{"US": host},
		},
	}
	g.uaPool, _ = httpx.NewUserAgentPool("/nonexistent/pool.json")

	ctx := context.WithValue(context.Background(), processor.CtxKeyUserLocale, "en-US")
	resp, err := g.Search(ctx, &models.Request{
		Query:    "golang",
		Category: models.CategoryGeneral,
		Page:     1,
	})
	require.NoError(t, err)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "https://go.dev", resp.Results[0].URL)
	assert.Equal(t, "go programming language", resp.Suggestions[0])
	assert.Contains(t, requestedPath, "q=golang")
	assert.Contains(t, requestedPath, "start=0")
}

func TestGoogleInfo_SGSSCookie(t *testing.T) {
	g := &Google{}
	info := g.googleInfo("all", engine.EngineTraits{AllLocale: "ZZ"}, config.GoogleEngineParams{
		ConsentCookie: "YES+cb.20210328-17-p0.en+FX+",
		SGSSCookie:    "ES_ICY:12345:",
	})
	assert.Equal(t, "YES+cb.20210328-17-p0.en+FX+", info.cookies["CONSENT"])
	assert.Equal(t, "ES_ICY:12345:", info.cookies["SG_SS"])
}

func TestDoRequest_SGSSCookieSent(t *testing.T) {
	g := &Google{}
	var cookieHeader string
	client, srv := newGoogleTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieHeader = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	g.client = client
	info := googleInfo{
		headers: map[string]string{"Accept": "*/*"},
		cookies: map[string]string{
			"CONSENT": "YES+",
			"SG_SS":   "ES_ICY:12345:",
		},
	}
	_, err := g.doRequest(context.Background(), srv.URL, info)
	require.NoError(t, err)
	assert.Contains(t, cookieHeader, "CONSENT=YES+")
	assert.Contains(t, cookieHeader, "SG_SS=ES_ICY:12345:")
}

func TestCandidateURLs(t *testing.T) {
	g := &Google{}
	base := "https://www.google.com/search?q=golang"

	urls := g.candidateURLs(base, "golang", config.GoogleEngineParams{})
	require.Len(t, urls, 2)
	assert.Equal(t, base+"&udm=14", urls[0])
	assert.Equal(t, base+"&gbv=1", urls[1])

	mobile := g.candidateURLs(base, "golang", config.GoogleEngineParams{UseMobileUI: true})
	require.Len(t, mobile, 3)
	assert.Equal(t, "https://www.google.com/search?hl=en&gbv=1&q=golang", mobile[1])
}

func TestCandidateURLs_Expanded(t *testing.T) {
	g := &Google{}
	base := "https://www.google.com/search?q=golang"

	urls := g.candidateURLs(base, "golang", config.GoogleEngineParams{})
	require.Len(t, urls, 3)
	assert.Equal(t, base+"&udm=14", urls[0])
	assert.Equal(t, "https://www.google.com/search?hl=en&gws_rd=ssl&q=golang", urls[1])
	assert.Equal(t, base+"&gbv=1", urls[2])

	mobile := g.candidateURLs(base, "golang", config.GoogleEngineParams{UseMobileUI: true})
	require.Len(t, mobile, 4)
	assert.Equal(t, "https://www.google.com/search?hl=en&gbv=1&q=golang", mobile[2])
}

func TestParseResults_NewMarkup(t *testing.T) {
	g := &Google{}
	html := `
<html>
<body>
  <div class="g">
    <div class="yuRUbf">
      <a data-ved="abc" href="/url?q=https%3A%2F%2Fgo.dev&sa=Uxyz">
        <h3>The Go Programming Language</h3>
      </a>
    </div>
    <div class="VwiC3b yXK7lf lVm3ye r025kc hJNv6b Hdw6tb">Go is an open source programming language.</div>
  </div>
  <div class="gGQDvd iIWm4b"><a>go programming language</a></div>
</body>
</html>`

	results, suggestions := g.parseResults(&httpx.Response{Body: []byte(html)})
	require.Len(t, results, 1)
	assert.Equal(t, "The Go Programming Language", results[0].Title)
	assert.Equal(t, "https://go.dev", results[0].URL)
	assert.Equal(t, "Go is an open source programming language.", results[0].Content)
	require.Len(t, suggestions, 1)
	assert.Equal(t, "go programming language", suggestions[0])
}

func TestParseResults_NonResultLinkSurvives(t *testing.T) {
	g := &Google{}
	html := `
<html>
<body>
  <a href="/about">About Google</a>
  <a data-ved="x" href="/url?q=https%3A%2F%2Fgo.dev"><div style="">Go</div></a>
  <div class="ilUpNd H66NU aSRlid">The Go Programming Language.</div>
</body>
</html>`

	results, _ := g.parseResults(&httpx.Response{Body: []byte(html)})
	require.Len(t, results, 1)
	assert.Equal(t, "Go", results[0].Title)
}
