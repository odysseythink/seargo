package imageproxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/seargo/seargo/internal/security"
	"github.com/seargo/seargo/pkg/models/results"
)

func TestSignedURL(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: true}, signer, nil)

	signed, err := proxy.SignedURL("https://example.com/a.png")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(signed, "/image_proxy?url=") {
		t.Fatalf("expected /image_proxy?url= prefix, got %q", signed)
	}
	u, _ := url.Parse("http://localhost" + signed)
	if u.Query().Get("h") == "" {
		t.Fatal("signed URL must have h= query parameter")
	}
}

func TestSignedURL_Consistency(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: true}, signer, nil)

	s1, _ := proxy.SignedURL("https://example.com/a.png")
	s2, _ := proxy.SignedURL("https://example.com/a.png")
	if s1 != s2 {
		t.Fatalf("same URL should produce same signed URL: %q vs %q", s1, s2)
	}
}

func TestSignedURL_DifferentSigner(t *testing.T) {
	s1 := security.NewHMACSigner("secret-a")
	s2 := security.NewHMACSigner("secret-b")
	p1 := New(Config{Enabled: true}, s1, nil)
	p2 := New(Config{Enabled: true}, s2, nil)

	url1, _ := p1.SignedURL("https://example.com/a.png")
	url2, _ := p2.SignedURL("https://example.com/a.png")
	if url1 == url2 {
		t.Fatal("different signers should produce different signed URLs")
	}
}

func TestSignedURL_DataURI(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: true}, signer, nil)

	dataURI := "data:image/png;base64,iVBORw0KGgo="
	signed, err := proxy.SignedURL(dataURI)
	if err != nil {
		t.Fatal(err)
	}
	if signed != dataURI {
		t.Fatalf("data URI should be returned unchanged: got %q", signed)
	}
}

func TestRewriteURLs_TypedResults(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: true}, signer, nil)

	results := results.ResultTypes{
		Images: []results.ImageResult{
			{BaseResult: results.BaseResult{URL: "https://ex.com/img.jpg"}, ImgSrc: "https://ex.com/src.jpg", ThumbnailSrc: "https://ex.com/thumb.jpg"},
		},
		Videos: []results.VideoResult{
			{BaseResult: results.BaseResult{URL: "https://ex.com/vid"}, Thumbnail: "https://ex.com/vthumb.jpg"},
		},
		Main: []results.MainResult{
			{BaseResult: results.BaseResult{URL: "https://ex.com/page", ThumbnailURL: "https://ex.com/thumb.png", Favicon: "https://ex.com/icon.png"}},
		},
		News: []results.NewsResult{
			{BaseResult: results.BaseResult{URL: "https://ex.com/news", ThumbnailURL: "https://ex.com/nthumb.jpg"}},
		},
		Infoboxes: []results.InfoboxResult{
			{BaseResult: results.BaseResult{URL: "https://ex.com/info"}, ImgSrc: "https://ex.com/infoimg.jpg"},
		},
	}

	proxy.RewriteURLs(&results)

	if !strings.HasPrefix(results.Images[0].ImgSrc, "/image_proxy?url=") {
		t.Fatalf("ImgSrc not rewritten: %q", results.Images[0].ImgSrc)
	}
	if !strings.HasPrefix(results.Images[0].ThumbnailSrc, "/image_proxy?url=") {
		t.Fatalf("ThumbnailSrc not rewritten: %q", results.Images[0].ThumbnailSrc)
	}
	if !strings.HasPrefix(results.Videos[0].Thumbnail, "/image_proxy?url=") {
		t.Fatalf("Video Thumbnail not rewritten: %q", results.Videos[0].Thumbnail)
	}
	if !strings.HasPrefix(results.Main[0].ThumbnailURL, "/image_proxy?url=") {
		t.Fatalf("MainResult ThumbnailURL not rewritten: %q", results.Main[0].ThumbnailURL)
	}
	if !strings.HasPrefix(results.Main[0].Favicon, "/image_proxy?url=") {
		t.Fatalf("MainResult Favicon not rewritten: %q", results.Main[0].Favicon)
	}
	if !strings.HasPrefix(results.News[0].ThumbnailURL, "/image_proxy?url=") {
		t.Fatalf("NewsResult ThumbnailURL not rewritten: %q", results.News[0].ThumbnailURL)
	}
	if !strings.HasPrefix(results.Infoboxes[0].ImgSrc, "/image_proxy?url=") {
		t.Fatalf("InfoboxResult ImgSrc not rewritten: %q", results.Infoboxes[0].ImgSrc)
	}
}

func TestRewriteURLs_Disabled(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: false}, signer, nil)

	results := results.ResultTypes{
		Images: []results.ImageResult{
			{BaseResult: results.BaseResult{URL: "https://ex.com/img.jpg"}, ImgSrc: "https://ex.com/src.jpg"},
		},
	}

	proxy.RewriteURLs(&results)
	if strings.HasPrefix(results.Images[0].ImgSrc, "/image_proxy") {
		t.Fatal("disabled image_proxy should not rewrite URLs")
	}
}

func TestServe_BadSignature(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: true}, signer, nil)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/image_proxy?url=https%3A%2F%2Fexample.com%2Fa.png&h=badsig", nil)
	_ = proxy.Serve(r.Context(), "https://example.com/a.png", "badsig", w)

	if w.Code != 400 {
		t.Fatalf("bad signature should return 400, got %d", w.Code)
	}
}

func TestServe_InvalidScheme(t *testing.T) {
	signer := security.NewHMACSigner("test-secret")
	proxy := New(Config{Enabled: true}, signer, nil)

	sig := signer.Sign([]byte("ftp://example.com/a.png"))

	w := httptest.NewRecorder()
	_ = proxy.Serve(nil, "ftp://example.com/a.png", sig, w)

	if w.Code != 400 {
		t.Fatalf("ftp scheme should return 400, got %d", w.Code)
	}
}

func TestAllowedContentType(t *testing.T) {
	cfg := Config{AllowedContentTypes: []string{"image/*", "binary/octet-stream"}}
	proxy := &imageProxy{cfg: cfg}

	tests := []struct {
		ct       string
		expected bool
	}{
		{"image/png", true},
		{"image/jpeg", true},
		{"image/webp", true},
		{"binary/octet-stream", true},
		{"text/html", false},
		{"application/json", false},
	}
	for _, tt := range tests {
		got := proxy.isAllowedContentType(tt.ct)
		if got != tt.expected {
			t.Errorf("isAllowedContentType(%q) = %v, want %v", tt.ct, got, tt.expected)
		}
	}
}
