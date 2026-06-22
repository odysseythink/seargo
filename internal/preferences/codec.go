package preferences

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// rawPreferences is the codec's internal representation: key→value mapping
// for each preference field. Values use comma-separated lists where needed
// (matching SearXNG's URL-encoded format).
type rawPreferences map[string]string

// CookieCodec encodes and decodes the preference cookie blob
// using SearXNG-compatible base64 + zlib + URL encoding.
type CookieCodec struct{}

// Encode encodes raw preferences as a URL-safe base64-encoded zlib-compressed query string.
func (c CookieCodec) Encode(raw rawPreferences) (string, error) {
	values := url.Values{}
	for k, v := range raw {
		values.Set(k, v)
	}
	queryString := values.Encode()

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write([]byte(queryString)); err != nil {
		return "", fmt.Errorf("zlib compress: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("zlib close: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(compressed.Bytes())
	return encoded, nil
}

// Decode decodes a URL-safe base64-encoded zlib-compressed query string back to rawPreferences.
func (c CookieCodec) Decode(blob string) (rawPreferences, error) {
	compressed, err := base64.RawURLEncoding.DecodeString(blob)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("zlib decompress: %w", err)
	}
	defer r.Close()

	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("zlib read: %w", err)
	}

	queryString := string(rawBytes)
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return nil, fmt.Errorf("parse query string: %w", err)
	}

	raw := make(rawPreferences, len(values))
	for k, v := range values {
		if len(v) > 0 {
			raw[k] = strings.TrimSpace(v[0])
		}
	}
	return raw, nil
}
