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
	Query      string   `form:"q" binding:"required"`
	Category   Category `form:"category"`
	Language   string   `form:"language"`
	SafeSearch int      `form:"safesearch"`
	TimeRange  string   `form:"time_range"`
	Page       int      `form:"page"`
	PageSize   int      `form:"page_size"`
}

func (r *Request) CacheKey() string {
	h := fnv.New64a()
	h.Write([]byte(r.Query))
	return fmt.Sprintf("search:%s:%s:%d:%s:%d:%d:%x",
		r.Category, r.Language, r.SafeSearch,
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
	if r.Category == "" {
		r.Category = d.DefaultCategory
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


type Result struct {
	Title        string     `json:"title"`
	URL          string     `json:"url"`
	Content      string     `json:"content"`
	Engine       string     `json:"engine"`
	Category     Category   `json:"category"`
	Score        float64    `json:"score"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
}

type Response struct {
	Query          string   `json:"query"`
	Category       Category `json:"category"`
	Results        []Result `json:"results"`
	Suggestions    []string `json:"suggestions"`
	Total          int      `json:"total"`
	Page           int      `json:"page"`
	PageSize       int      `json:"page_size"`
	EnginesUsed    []string `json:"engines_used"`
	EnginesFailed  []string `json:"engines_failed"`
	ResponseTimeMs int64    `json:"response_time_ms"`
}
