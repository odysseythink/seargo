package githubcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("github_code", &GitHubCode{})
}

var searchAPI = "https://api.github.com/search/code"

// GitHubCode queries GitHub's REST code search API.
type GitHubCode struct {
	client          *httpx.Client
	categories      []models.Category
	authType        string
	token           string
	highlight       bool
	stripNewlines   bool
	stripWhitespace bool
	insertSeparator bool
	apiVersion      string
}

func (g *GitHubCode) Name() string { return "github_code" }

func (g *GitHubCode) Categories() []models.Category { return g.categories }

func (g *GitHubCode) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://github.com",
		WikidataID: "Q364",
	}
}

func (g *GitHubCode) Capabilities() engine.Capabilities {
	return engine.Capabilities{SupportsPagination: true}
}

func (g *GitHubCode) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (g *GitHubCode) Setup(cfg engine.EngineInitConfig) bool {
	g.client = cfg.Client
	g.categories = cfg.Categories
	if len(g.categories) == 0 {
		g.categories = []models.Category{models.Category("code")}
	}

	extra := cfg.Extra
	if extra == nil {
		extra = map[string]any{}
	}

	g.authType = "none"
	if a, ok := extra["ghc_auth"].(map[string]any); ok {
		if t, ok := a["type"].(string); ok {
			g.authType = t
		}
		if tok, ok := a["token"].(string); ok {
			g.token = tok
		}
	}

	g.highlight = true
	if v, ok := extra["ghc_highlight_matching_lines"].(bool); ok {
		g.highlight = v
	}
	g.stripNewlines = true
	if v, ok := extra["ghc_strip_new_lines"].(bool); ok {
		g.stripNewlines = v
	}
	g.stripWhitespace = false
	if v, ok := extra["ghc_strip_whitespace"].(bool); ok {
		g.stripWhitespace = v
	}
	g.insertSeparator = false
	if v, ok := extra["ghc_insert_block_separator"].(bool); ok {
		g.insertSeparator = v
	}
	g.apiVersion = "2022-11-28"
	if v, ok := extra["ghc_api_version"].(string); ok && v != "" {
		g.apiVersion = v
	}

	return g.client != nil
}

func (g *GitHubCode) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	u := fmt.Sprintf("%s?q=%s&page=%s&sort=indexed",
		searchAPI,
		url.QueryEscape(req.Query),
		url.QueryEscape(strconv.Itoa(req.Page)),
	)

	r := g.client.R().SetContext(ctx)
	r.SetHeader("Accept", "application/vnd.github.text-match+json")
	r.SetHeader("X-GitHub-Api-Version", g.apiVersion)
	switch g.authType {
	case "personal_access_token":
		r.SetHeader("Authorization", "token "+g.token)
	case "bearer":
		r.SetHeader("Authorization", "Bearer "+g.token)
	default:
		r.SetHeader("Authorization", "placeholder")
	}

	resp, err := r.Get(u)
	if err != nil {
		return nil, fmt.Errorf("github_code request failed: %w", err)
	}
	if resp.StatusCode == 422 {
		return &models.Response{Query: req.Query, Category: req.Category}, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github_code status %d", resp.StatusCode)
	}

	var payload struct {
		Items []struct {
			Name        string `json:"name"`
			Path        string `json:"path"`
			HTMLURL     string `json:"html_url"`
			TextMatches []struct {
				ObjectType string `json:"object_type"`
				Property   string `json:"property"`
				Fragment   string `json:"fragment"`
				Matches    []struct {
					Indices []int `json:"indices"`
				} `json:"matches"`
			} `json:"text_matches"`
			Repository struct {
				FullName    string `json:"full_name"`
				HTMLURL     string `json:"html_url"`
				Description string `json:"description"`
			} `json:"repository"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("github_code parse: %w", err)
	}

	var typed []results.Result
	for _, item := range payload.Items {
		matches := make([]textMatch, 0)
		for _, m := range item.TextMatches {
			if m.ObjectType != "FileContent" || m.Property != "content" {
				continue
			}
			hls := make([]highlight, 0, len(m.Matches))
			for _, hg := range m.Matches {
				hls = append(hls, highlight{Indices: hg.Indices})
			}
			matches = append(matches, textMatch{Fragment: m.Fragment, Highlights: hls})
		}

		lines, hl := extractCode(matches, g.stripWhitespace, g.stripNewlines, g.insertSeparator)
		if !g.highlight {
			hl = nil
		}

		typed = append(typed, &results.CodeResult{
			BaseResult: results.BaseResult{
				Title:    fmt.Sprintf("%s \u00b7 %s", item.Repository.FullName, item.Name),
				URL:      item.HTMLURL,
				Content:  item.Repository.Description,
				Engine:   g.Name(),
				Category: string(req.Category),
			},
			Repository:   item.Repository.HTMLURL,
			Filename:     item.Path,
			CodeLines:    lines,
			HLLines:      hl,
		})
	}

	return &models.Response{
		Query:        req.Query,
		Category:     req.Category,
		Results:      results.ToAPIResult(typed),
		TypedResults: toAnySlice(typed),
	}, nil
}

type highlight struct {
	Indices []int
}

type textMatch struct {
	Fragment   string
	Highlights []highlight
}

func extractCode(matches []textMatch, stripWS, stripNL, insertSep bool) ([]results.CodeLine, []int) {
	var lines []results.CodeLine
	var hl []int

	for mi, match := range matches {
		if mi > 0 && insertSep {
			lines = append(lines, results.CodeLine{Line: len(lines) + 1, Text: "..."})
		}

		code := match.Fragment
		origLen := len(code)
		if stripWS {
			code = strings.TrimSpace(code)
		}
		if stripNL {
			code = strings.Trim(code, "\n")
		}
		offset := origLen - len(code)

		hgroups := make([][]int, 0, len(match.Highlights))
		for _, h := range match.Highlights {
			hgroups = append(hgroups, h.Indices)
		}

		var buf strings.Builder
		for i := 0; i < len(code); i++ {
			if len(hgroups) > 0 {
				after, before := hgroups[0][0], hgroups[0][1]
				if after <= (i+offset) && (i+offset) < before {
					hl = append(hl, len(lines)+1)
					hgroups = hgroups[1:]
				}
			}
			ch := code[i]
			if ch == '\n' {
				lines = append(lines, results.CodeLine{Line: len(lines) + 1, Text: buf.String()})
				buf.Reset()
				continue
			}
			buf.WriteByte(ch)
		}
		lines = append(lines, results.CodeLine{Line: len(lines) + 1, Text: buf.String()})
	}

	return lines, hl
}

func toAnySlice(typed []results.Result) []any {
	out := make([]any, len(typed))
	for i, r := range typed {
		out[i] = r
	}
	return out
}
