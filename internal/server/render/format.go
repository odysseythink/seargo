package render

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/seargo/seargo/pkg/models"
)

// Format is an output format for search results.
type Format string

const (
	FormatHTML Format = "html"
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
	FormatRSS  Format = "rss"
)

type mediaType struct {
	mime   string
	format Format
}

var acceptedMediaTypes = []mediaType{
	{"application/json", FormatJSON},
	{"text/csv", FormatCSV},
	{"application/rss+xml", FormatRSS},
	{"application/atom+xml", FormatRSS},
}

// ResolveFormat determines the output format from query parameter, Accept header,
// and allowed format whitelist.
func ResolveFormat(query, accept string, allowed []string) (Format, error) {
	if len(allowed) == 0 {
		return "", fmt.Errorf("no formats allowed")
	}
	allowedSet := make(map[Format]bool, len(allowed))
	for _, a := range allowed {
		allowedSet[Format(strings.ToLower(a))] = true
	}
	// 1. Query param takes precedence
	if query != "" {
		f := Format(strings.ToLower(query))
		switch f {
		case FormatJSON, FormatCSV, FormatRSS, FormatHTML:
			if allowedSet[f] {
				return f, nil
			}
			return "", fmt.Errorf("format %q not allowed", f)
		default:
			return "", fmt.Errorf("unknown format %q", query)
		}
	}
	// 2. Accept header negotiation
	if accept != "" && accept != "*/*" {
		f := negotiateAccept(accept, allowedSet)
		if f != "" {
			return f, nil
		}
		// A specific Accept header was provided but none matched allowed formats.
		return "", fmt.Errorf("no acceptable format for accept header %q", accept)
	}
	// 3. Default fallback
	if allowedSet[FormatHTML] {
		return FormatHTML, nil
	}
	return Format(strings.ToLower(allowed[0])), nil
}

type acceptEntry struct {
	mediaRange string
	q          float64
}

func negotiateAccept(accept string, allowed map[Format]bool) Format {
	entries := parseAccept(accept)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].q != entries[j].q {
			return entries[i].q > entries[j].q
		}
		return len(entries[i].mediaRange) > len(entries[j].mediaRange)
	})
	for _, e := range entries {
		for _, mt := range acceptedMediaTypes {
			if e.mediaRange == mt.mime && allowed[mt.format] {
				return mt.format
			}
		}
	}
	return ""
}

// Render dispatches the response to the appropriate writer based on format.
func Render(resp *models.Response, format Format, baseURL string) ([]byte, string, error) {
    switch format {
    case FormatJSON, FormatHTML:
        w := &JSONWriter{}
        data, err := w.Render(resp)
        return data, w.ContentType(), err
    case FormatCSV:
        w := &CSVWriter{}
        data, err := w.Render(resp)
        return data, w.ContentType(), err
    case FormatRSS:
        w := &RSSWriter{}
        data, err := w.Render(resp, baseURL)
        return data, w.ContentType(), err
    default:
        return nil, "", fmt.Errorf("unsupported format: %s", format)
    }
}

func parseAccept(accept string) []acceptEntry {
	parts := strings.Split(accept, ",")
	entries := make([]acceptEntry, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		semiIdx := strings.Index(part, ";")
		mediaRange := part
		q := 1.0
		if semiIdx >= 0 {
			mediaRange = strings.TrimSpace(part[:semiIdx])
			params := part[semiIdx+1:]
			for _, param := range strings.Split(params, ";") {
				param = strings.TrimSpace(param)
				if strings.HasPrefix(param, "q=") {
					if v, err := strconv.ParseFloat(param[2:], 64); err == nil {
						q = v
					}
				}
			}
		}
		entries = append(entries, acceptEntry{mediaRange: mediaRange, q: q})
	}
	return entries
}
