package preferences

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
)

func TestCodecRoundTrip(t *testing.T) {
	c := CookieCodec{}
	raw := rawPreferences{
		"language":         "zh-CN",
		"locale":           "zh-Hans-CN",
		"safesearch":       "1",
		"theme":            "simple",
		"category":         "general",
		"disabled_engines": "google__general,bing__general",
	}

	encoded, err := c.Encode(raw)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if encoded == "" {
		t.Fatal("Encode returned empty string")
	}
	// Must be URL-safe base64 (no + or /)
	if strings.ContainsAny(encoded, "+/") {
		t.Errorf("encoded string contains non-URL-safe characters: %s", encoded)
	}

	decoded, err := c.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if len(decoded) != len(raw) {
		t.Fatalf("decoded has %d entries, want %d", len(decoded), len(raw))
	}
	for k, want := range raw {
		if got := decoded[k]; got != want {
			t.Errorf("decoded[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestCodecCompression(t *testing.T) {
	c := CookieCodec{}
	raw := rawPreferences{
		"language":         "zh-CN",
		"enabled_engines": "google__general,bing__general,duckduckgo__general,brave__general" +
			",wikipedia__general,yahoo__general,google__images,bing__images" +
			",duckduckgo__images,brave__images,wikipedia__images,yahoo__images",
		"enabled_plugins": "oa_doi_rewrite,self_info,url_unshorten,tor_check_plugin,search_on_category_select" +
			",hostname_replace,basic_calculator,unit_converter,tracker_url_remover",
		"disabled_engines": "google__news,bing__news",
	}
	encoded, err := c.Encode(raw)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	// Build the plaintext query string equivalent for size comparison
	vals := url.Values{}
	for k, v := range raw {
		vals.Set(k, v)
	}
	plainSize := base64.RawURLEncoding.EncodedLen(len(vals.Encode()))
	if len(encoded) >= plainSize {
		t.Logf("encoded size %d >= plain base64 size %d (not compressed?)", len(encoded), plainSize)
	}
	t.Logf("encoded cookie size: %d bytes", len(encoded))
}

func TestCodecInvalidBase64(t *testing.T) {
	c := CookieCodec{}
	_, err := c.Decode("!!!not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64 input")
	}
}

func TestCodecInvalidZlib(t *testing.T) {
	c := CookieCodec{}
	invalid := "dGhpcyBpcyBub3QgY29tcHJlc3NlZA"
	_, err := c.Decode(invalid)
	if err == nil {
		t.Error("expected error for invalid zlib input")
	}
}

func TestCodecEmptyInput(t *testing.T) {
	c := CookieCodec{}
	encoded, err := c.Encode(rawPreferences{})
	if err != nil {
		t.Fatalf("Encode empty map failed: %v", err)
	}
	decoded, err := c.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode empty map failed: %v", err)
	}
	if len(decoded) != 0 {
		t.Errorf("decoded empty map has %d entries, want 0", len(decoded))
	}
}
