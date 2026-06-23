package upstream

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

// NormalizedResult is a vendor-neutral result used for comparison.
type NormalizedResult struct {
	URL           string
	Title         string
	Content       string
	Engine        string
	Engines       []string
	Category      string
	Template      string
	Score         float64
	Positions     []int
	ThumbnailURL  string
	PublishedDate string
	TypedFields   map[string]string
}

// NormalizedResponse is a vendor-neutral search response.
type NormalizedResponse struct {
	Query               string
	Total               int
	Results             []NormalizedResult
	Suggestions         []string
	Corrections         []string
	Answers             []NormalizedAnswer
	Infoboxes           []NormalizedInfobox
	UnresponsiveEngines []string
}

// NormalizedAnswer is a vendor-neutral answer.
type NormalizedAnswer struct {
	Answer  string
	URL     string
	Content string
}

// NormalizedInfobox is a vendor-neutral infobox.
type NormalizedInfobox struct {
	Title   string
	URL     string
	Content string
	Engines []string
}

// NormalizeUpstream converts an upstream response to the common shape.
func NormalizeUpstream(up *UpstreamResponse) NormalizedResponse {
	r := NormalizedResponse{
		Query:       up.Query,
		Total:       len(up.Results),
		Suggestions: sortedCopy(up.Suggestions),
		Corrections: sortedCopy(up.Corrections),
	}
	for _, res := range up.Results {
		r.Results = append(r.Results, NormalizedResult{
			URL:           canonicalURL(res.URL),
			Title:         strings.TrimSpace(res.Title),
			Content:       strings.TrimSpace(res.Content),
			Engine:        res.Engine,
			Engines:       sortedCopy(res.Engines),
			Category:      strings.ToLower(res.Category),
			Template:      res.Template,
			Score:         roundScore(res.Score),
			Positions:     sortedInts(res.Positions),
			ThumbnailURL:  canonicalURL(coalesce(res.Thumbnail, res.ImgSrc)),
			PublishedDate: parseUpstreamDate(res.PublishedDate),
			TypedFields:   upstreamTypedFields(res),
		})
	}
	for _, a := range up.Answers {
		r.Answers = append(r.Answers, NormalizedAnswer{
			Answer:  strings.TrimSpace(a.Answer),
			URL:     canonicalURL(a.URL),
			Content: strings.TrimSpace(a.Content),
		})
	}
	for _, i := range up.Infoboxes {
		r.Infoboxes = append(r.Infoboxes, NormalizedInfobox{
			Title:   strings.TrimSpace(i.Title),
			URL:     canonicalURL(i.URL),
			Content: strings.TrimSpace(i.Content),
			Engines: sortedCopy(i.Engines),
		})
	}
	for _, u := range up.UnresponsiveEngines {
		r.UnresponsiveEngines = append(r.UnresponsiveEngines, u.Engine+":"+u.Reason)
	}
	sort.Slice(r.Results, func(i, j int) bool { return r.Results[i].URL < r.Results[j].URL })
	return r
}

// NormalizeSearGo converts a SearGo response to the common shape.
func NormalizeSearGo(sg *models.Response) NormalizedResponse {
	r := NormalizedResponse{
		Query:       sg.Query,
		Total:       sg.Total,
		Suggestions: sortedCopy(sg.Suggestions),
		Corrections: sortedCopy(sg.Corrections),
	}
	for _, res := range sg.Results {
		thumb := res.ThumbnailURL
		if thumb == "" && res.Extra != nil {
			if s, ok := res.Extra["thumbnail"].(string); ok {
				thumb = s
			}
		}
		var pub string
		if res.PublishedAt != nil {
			pub = res.PublishedAt.UTC().Format("2006-01-02")
		}
		r.Results = append(r.Results, NormalizedResult{
			URL:           canonicalURL(res.URL),
			Title:         strings.TrimSpace(res.Title),
			Content:       strings.TrimSpace(res.Content),
			Engine:        res.Engine,
			Engines:       sortedCopy(res.Engines),
			Category:      strings.ToLower(string(res.Category)),
			Template:      res.Template,
			Score:         roundScore(res.Score),
			Positions:     sortedInts(res.Positions),
			ThumbnailURL:  canonicalURL(thumb),
			PublishedDate: pub,
			TypedFields:   seargoTypedFields(res),
		})
	}
	for _, a := range sg.Answers {
		r.Answers = append(r.Answers, NormalizedAnswer{
			Answer:  strings.TrimSpace(a.Answer),
			URL:     canonicalURL(a.URL),
			Content: strings.TrimSpace(a.Content),
		})
	}
	for _, i := range sg.Infoboxes {
		r.Infoboxes = append(r.Infoboxes, NormalizedInfobox{
			Title:   strings.TrimSpace(i.Title),
			URL:     canonicalURL(i.URL),
			Content: strings.TrimSpace(i.Content),
			Engines: sortedCopy(i.Engines),
		})
	}
	for _, e := range sg.EnginesFailed {
		r.UnresponsiveEngines = append(r.UnresponsiveEngines, e)
	}
	sort.Slice(r.Results, func(i, j int) bool { return r.Results[i].URL < r.Results[j].URL })
	return r
}

func upstreamTypedFields(res UpstreamResult) map[string]string {
	m := map[string]string{}
	add := func(k, v string) {
		if v != "" {
			m[k] = canonicalURL(v)
		}
	}
	add("img_src", res.ImgSrc)
	add("thumbnail", res.Thumbnail)
	return m
}

func seargoTypedFields(res models.Result) map[string]string {
	m := map[string]string{}
	if res.Kind != "" && res.Kind != "main" {
		m["kind"] = res.Kind
	}
	add := func(k, v string) {
		if v != "" {
			m[k] = canonicalURL(v)
		}
	}
	if res.Extra != nil {
		for _, k := range []string{"img_src", "thumbnail_src", "thumbnail", "iframe_src", "audio_src", "length", "views", "author", "metadata"} {
			if v, ok := res.Extra[k].(string); ok {
				add(k, v)
			}
		}
	}
	return m
}

func canonicalURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Fragment = ""
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func sortedCopy(ss []string) []string {
	out := make([]string, len(ss))
	copy(out, ss)
	sort.Strings(out)
	return out
}

func sortedInts(ii []int) []int {
	out := make([]int, len(ii))
	copy(out, ii)
	sort.Ints(out)
	return out
}

func roundScore(s float64) float64 {
	return float64(int64(s*1000+0.5)) / 1000
}

func parseUpstreamDate(raw string) string {
	if raw == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if d, err := time.Parse(layout, raw); err == nil {
			return d.UTC().Format("2006-01-02")
		}
	}
	return raw
}
