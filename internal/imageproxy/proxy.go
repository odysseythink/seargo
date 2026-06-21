package imageproxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/seargo/seargo/internal/security"
	"github.com/seargo/seargo/pkg/models/results"
)

// Proxy handles URL signing and upstream fetching.
type Proxy interface {
	SignedURL(rawURL string) (string, error)
	Serve(ctx context.Context, rawURL, signature string, w http.ResponseWriter) error
	RewriteURLs(r *results.ResultTypes)
}

type imageProxy struct {
	cfg    Config
	signer security.HMACSigner
	client *http.Client
}

// Config holds image proxy settings.
type Config struct {
	Enabled             bool
	MaxSize             int64
	AllowedContentTypes []string
}

// New creates a new image proxy. client may be nil (falls back to http.DefaultClient).
func New(cfg Config, signer security.HMACSigner, client *http.Client) Proxy {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 5 << 20 // 5 MiB
	}
	if len(cfg.AllowedContentTypes) == 0 {
		cfg.AllowedContentTypes = []string{"image/*", "binary/octet-stream"}
	}
	return &imageProxy{cfg: cfg, signer: signer, client: client}
}

func (p *imageProxy) SignedURL(rawURL string) (string, error) {
	if rawURL == "" {
		return rawURL, nil
	}
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}
	if strings.HasPrefix(rawURL, "data:") {
		return rawURL, nil
	}
	sig := p.signer.Sign([]byte(rawURL))
	return "/image_proxy?url=" + url.QueryEscape(rawURL) + "&h=" + sig, nil
}

func (p *imageProxy) proxify(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}
	signed, err := p.SignedURL(rawURL)
	if err != nil {
		return rawURL
	}
	return signed
}

func (p *imageProxy) RewriteURLs(r *results.ResultTypes) {
	if !p.cfg.Enabled {
		return
	}
	for i := range r.Images {
		r.Images[i].ImgSrc = p.proxify(r.Images[i].ImgSrc)
		r.Images[i].ThumbnailSrc = p.proxify(r.Images[i].ThumbnailSrc)
		for j := range r.Images[i].Formats {
			r.Images[i].Formats[j].URL = p.proxify(r.Images[i].Formats[j].URL)
		}
	}
	for i := range r.Videos {
		r.Videos[i].Thumbnail = p.proxify(r.Videos[i].Thumbnail)
		r.Videos[i].ThumbnailURL = p.proxify(r.Videos[i].ThumbnailURL)
	}
	for i := range r.Main {
		r.Main[i].ThumbnailURL = p.proxify(r.Main[i].ThumbnailURL)
		r.Main[i].Favicon = p.proxify(r.Main[i].Favicon)
	}
	for i := range r.News {
		r.News[i].ThumbnailURL = p.proxify(r.News[i].ThumbnailURL)
		r.News[i].Favicon = p.proxify(r.News[i].Favicon)
	}
	for i := range r.Infoboxes {
		r.Infoboxes[i].ImgSrc = p.proxify(r.Infoboxes[i].ImgSrc)
	}
}

func (p *imageProxy) Serve(ctx context.Context, rawURL, signature string, w http.ResponseWriter) error {
	if rawURL == "" {
		http.Error(w, "missing url", 400)
		return fmt.Errorf("missing url")
	}
	if !p.signer.Verify([]byte(rawURL), signature) {
		http.Error(w, "invalid signature", 400)
		return fmt.Errorf("invalid signature")
	}

	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		http.Error(w, "invalid URL scheme", 400)
		return fmt.Errorf("invalid scheme: %q", rawURL)
	}

	client := p.client
	if client == nil {
		client = http.DefaultClient
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		http.Error(w, "bad request", 400)
		return err
	}
	httpReq.Header.Set("Accept", "image/webp,*/*")
	httpReq.Header.Set("Sec-GPC", "1")
	httpReq.Header.Set("DNT", "1")

	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, "upstream error", 502)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "upstream error", http.StatusBadRequest)
		return fmt.Errorf("upstream status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !p.isAllowedContentType(contentType) {
		http.Error(w, "content type not allowed", 400)
		return fmt.Errorf("disallowed Content-Type: %q", contentType)
	}

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		var length int64
		fmt.Sscanf(cl, "%d", &length)
		if length > p.cfg.MaxSize {
			http.Error(w, "content too large", 400)
			return fmt.Errorf("Content-Length %d > max %d", length, p.cfg.MaxSize)
		}
	}

	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	limited := io.LimitReader(resp.Body, p.cfg.MaxSize)
	written, err := io.Copy(w, limited)
	if err != nil {
		return err
	}
	if written >= p.cfg.MaxSize {
		return fmt.Errorf("response exceeded max size")
	}
	return nil
}

func (p *imageProxy) isAllowedContentType(ct string) bool {
	for _, pattern := range p.cfg.AllowedContentTypes {
		if ok, _ := filepath.Match(pattern, ct); ok {
			return true
		}
	}
	return false
}
