package builtin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/stretchr/testify/assert"
)

func TestTorCheck_KeywordTriggersPostSearch(t *testing.T) {
	p := &torCheckPlugin{
		httpClient: newMockTorClient(t, "1.2.3.4", "5.6.7.8"),
	}

	ctx := &plugin.SearchContext{
		Query:      "tor-check",
		RemoteAddr: "1.2.3.4",
		PageNo:     1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Equal(t, "answer", results[0].Kind)
	assert.Contains(t, results[0].Content, "1.2.3.4")
	assert.Contains(t, results[0].Content, "Tor exit node")
}

func TestTorCheck_NoMatchReturnsNil(t *testing.T) {
	p := &torCheckPlugin{}

	ctx := &plugin.SearchContext{
		Query:      "something else",
		RemoteAddr: "1.2.3.4",
		PageNo:     1,
	}

	results := p.PostSearch(ctx)
	assert.Nil(t, results)
}

func TestTorCheck_PageGreaterThanOneReturnsNil(t *testing.T) {
	p := &torCheckPlugin{}

	ctx := &plugin.SearchContext{
		Query:      "tor-check",
		RemoteAddr: "1.2.3.4",
		PageNo:     2,
	}

	results := p.PostSearch(ctx)
	assert.Nil(t, results)
}

func TestTorCheck_KeywordsList(t *testing.T) {
	p := &torCheckPlugin{}
	keywords := p.Info().Keywords

	expected := []string{"tor-check", "tor_check", "torcheck", "tor", "tor check"}
	assert.ElementsMatch(t, expected, keywords)
}

func TestTorCheck_UnknownIP(t *testing.T) {
	p := &torCheckPlugin{
		httpClient: newMockTorClient(t),
	}

	ctx := &plugin.SearchContext{
		Query:      "tor-check",
		RemoteAddr: "",
		PageNo:     1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Content, "unknown")
}

func TestTorCheck_AllKeywordsTrigger(t *testing.T) {
	for _, kw := range []string{"tor-check", "tor_check", "torcheck", "tor", "tor check"} {
		t.Run(kw, func(t *testing.T) {
			p := &torCheckPlugin{
				httpClient: newMockTorClient(t),
			}
			ctx := &plugin.SearchContext{
				Query:      kw,
				RemoteAddr: "10.0.0.1",
				PageNo:     1,
			}
			results := p.PostSearch(ctx)
			assert.Len(t, results, 1, "keyword %q should trigger", kw)
		})
	}
}

func TestTorCheck_DetectsTorExitNode(t *testing.T) {
	p := &torCheckPlugin{
		httpClient: newMockTorClient(t, "10.0.0.1", "10.0.0.2"),
	}

	ctx := &plugin.SearchContext{
		Query:      "tor-check",
		RemoteAddr: "10.0.0.1",
		PageNo:     1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Content, "appears to be a Tor exit node")
}

func TestTorCheck_NotTorExitNode(t *testing.T) {
	p := &torCheckPlugin{
		httpClient: newMockTorClient(t, "10.0.0.1", "10.0.0.2"),
	}

	ctx := &plugin.SearchContext{
		Query:      "tor-check",
		RemoteAddr: "192.168.1.1",
		PageNo:     1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Content, "does not appear to be a Tor exit node")
}

func TestTorCheck_FetchFailure(t *testing.T) {
	p := &torCheckPlugin{
		httpClient: &http.Client{Transport: errorRoundTripper{}},
	}

	ctx := &plugin.SearchContext{
		Query:      "tor-check",
		RemoteAddr: "1.2.3.4",
		PageNo:     1,
	}

	results := p.PostSearch(ctx)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Content, "Tor check unavailable")
}

// --- test helpers ---

// newMockTorClient creates an *http.Client whose transport routes all requests
// to a local test server that responds with the given IPs as Tor exit addresses.
func newMockTorClient(t *testing.T, ips ...string) *http.Client {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ip := range ips {
			fmt.Fprintf(w, "ExitAddress %s 2025-01-01 12:00:00\n", ip)
		}
	}))
	t.Cleanup(ts.Close)

	return &http.Client{
		Transport: &redirectTransport{target: ts.URL},
	}
}

// redirectTransport is an http.RoundTripper that redirects all requests to targetURL.
type redirectTransport struct {
	target string // e.g. "http://127.0.0.1:PORT"
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	redirectReq, err := http.NewRequest(req.Method, rt.target+req.URL.Path, req.Body)
	if err != nil {
		return nil, err
	}
	redirectReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(redirectReq)
}

// errorRoundTripper always returns an error for any request.
type errorRoundTripper struct{}

func (errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("simulated network error")
}
