package models

import (
	"fmt"
	"hash/fnv"
	"time"
)

type Category string

const (
	CategoryGeneral     Category = "general"
	CategoryImages      Category = "images"
	CategoryVideos      Category = "videos"
	CategoryNews        Category = "news"
	CategoryMap         Category = "map"
	CategoryMusic       Category = "music"
	CategoryIT          Category = "it"
	CategoryScience     Category = "science"
	CategoryFiles       Category = "files"
	CategorySocialMedia Category = "social media"
)

func AllCategories() []Category {
	return []Category{
		CategoryGeneral, CategoryImages, CategoryVideos, CategoryNews,
		CategoryMap, CategoryMusic, CategoryIT, CategoryScience,
		CategoryFiles, CategorySocialMedia,
	}
}

type Request struct {
	Query      string     `form:"q" binding:"required"`
	Category   Category   `form:"category"`
	Categories []Category `form:"categories"`
	Language   string     `form:"language"`
	Locale     string     `form:"locale"`
	SafeSearch int        `form:"safesearch"`
	TimeRange  string     `form:"time_range"`
	Page       int        `form:"page"`
	PageSize   int        `form:"page_size"`
}

func (r *Request) CacheKey() string {
	h := fnv.New64a()
	h.Write([]byte(r.Query))
	cats := r.Categories
	if len(cats) == 0 && r.Category != "" {
		cats = []Category{r.Category}
	}
	catStr := ""
	for _, c := range cats {
		if catStr != "" {
			catStr += ","
		}
		catStr += string(c)
	}
	return fmt.Sprintf("search:%s:%s:%s:%d:%s:%d:%d:%x",
		catStr, r.Language, r.Locale, r.SafeSearch,
		r.TimeRange, r.Page, r.PageSize, h.Sum64())
}

type NormalizeDefaults struct {
	DefaultLang     string
	DefaultCategory Category
	DefaultPageSize int
	MaxResults      int
}

func (r *Request) Normalize(d NormalizeDefaults) {
	if r.Language == "" {
		r.Language = d.DefaultLang
	}
	if len(r.Categories) == 0 {
		if r.Category != "" {
			r.Categories = []Category{r.Category}
		} else {
			r.Categories = []Category{d.DefaultCategory}
		}
	}
	if r.Category == "" {
		r.Category = r.Categories[0]
	}
	if r.PageSize <= 0 {
		r.PageSize = d.DefaultPageSize
	}
	if r.Page <= 0 {
		r.Page = 1
	}
	// Cap PageSize to MaxResults
	if r.PageSize > d.MaxResults && d.MaxResults > 0 {
		r.PageSize = d.MaxResults
	}
}

type Answer struct {
	Answer  string `json:"answer"`
	URL     string `json:"url,omitempty"`
	Content string `json:"content"`
	Engine  string `json:"engine,omitempty"`
}

// InfoboxAttribute — a key-value attribute in an infobox (API type).
type InfoboxAttribute struct {
	Value string `json:"value"`
	Label string `json:"label"`
	URL   string `json:"url,omitempty"`
}

// InfoboxURL — a URL entry in an infobox (API type).
type InfoboxURL struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Infobox struct {
	Title         string             `json:"title"`
	URL           string             `json:"url,omitempty"`
	Content       string             `json:"content,omitempty"`
	Engine        string             `json:"engine,omitempty"`
	Engines       []string           `json:"engines,omitempty"`
	ImgSrc        string             `json:"img_src,omitempty"`
	URLs          []InfoboxURL       `json:"urls,omitempty"`
	Attributes    []InfoboxAttribute `json:"attributes,omitempty"`
	RelatedTopics []string           `json:"related_topics,omitempty"`
}

type Result struct {
	Kind         string         `json:"kind"`
	Template     string         `json:"template,omitempty"`
	Title        string         `json:"title"`
	URL          string         `json:"url"`
	Content      string         `json:"content,omitempty"`
	Engine       string         `json:"engine"`
	Engines      []string       `json:"engines,omitempty"`
	Category     Category       `json:"category"`
	Score        float64        `json:"score"`
	Positions    []int          `json:"-"`
	ThumbnailURL string         `json:"thumbnail_url,omitempty"`
	PublishedAt  *time.Time     `json:"published_at,omitempty"`
	Domain       string         `json:"domain,omitempty"`
	Favicon      string         `json:"favicon,omitempty"`
	IsOnion      bool           `json:"is_onion,omitempty"`
	Extra        map[string]any `json:"extra,omitempty"`
}

// FilterURLs calls fn for every URL field on the result.
// fn receives (result, fieldName, currentURL) and returns (newURL, keep).
// If keep is false, the field is set to "".
// If keep is true, the field is set to newURL.
func (r *Result) FilterURLs(fn func(r *Result, field string, url string) (string, bool)) {
	if r.URL != "" {
		newURL, keep := fn(r, "url", r.URL)
		if keep {
			r.URL = newURL
		} else {
			r.URL = ""
		}
	}
	if r.ThumbnailURL != "" {
		newURL, keep := fn(r, "thumbnail_url", r.ThumbnailURL)
		if keep {
			r.ThumbnailURL = newURL
		} else {
			r.ThumbnailURL = ""
		}
	}
	if r.Favicon != "" {
		newURL, keep := fn(r, "favicon", r.Favicon)
		if keep {
			r.Favicon = newURL
		} else {
			r.Favicon = ""
		}
	}
}

type Response struct {
	Query          string         `json:"query"`
	Category       Category       `json:"category"`
	Results        []Result       `json:"results"`
	Suggestions    []string       `json:"suggestions"`
	Answers        []Answer       `json:"answers"`
	Corrections    []string       `json:"corrections"`
	Infoboxes      []Infobox      `json:"infoboxes"`
	EngineData     map[string]any `json:"engine_data"`
	Total          int            `json:"total"`
	Page           int            `json:"page"`
	PageSize       int            `json:"page_size"`
	EnginesUsed    []string       `json:"engines_used"`
	EnginesFailed  []string       `json:"engines_failed"`
	ResponseTimeMs int64          `json:"response_time_ms"`
	RedirectURL    string         `json:"redirect_url,omitempty"`
}
