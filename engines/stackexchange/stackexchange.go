package stackexchange

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("stackexchange", &StackExchange{})
}

var searchAPI = "https://api.stackexchange.com/2.3/search/advanced?"

// StackExchange queries the Stack Exchange API v2.3.
type StackExchange struct {
	client     *httpx.Client
	apiSite    string
	categories []models.Category
}

func (s *StackExchange) Name() string { return "stackexchange" }

func (s *StackExchange) Categories() []models.Category { return s.categories }

func (s *StackExchange) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://stackexchange.com",
		WikidataID: "Q3495447",
	}
}

func (s *StackExchange) Capabilities() engine.Capabilities {
	return engine.Capabilities{SupportsPagination: true}
}

func (s *StackExchange) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (s *StackExchange) Setup(cfg engine.EngineInitConfig) bool {
	s.client = cfg.Client
	s.categories = cfg.Categories
	if len(s.categories) == 0 {
		s.categories = []models.Category{models.CategoryIT}
	}
	s.apiSite = "stackoverflow"
	if cfg.Extra != nil {
		if v, ok := cfg.Extra["api_site"].(string); ok && v != "" {
			s.apiSite = v
		}
	}
	return s.client != nil && s.apiSite != ""
}

func (s *StackExchange) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	args := url.Values{}
	args.Set("q", req.Query)
	args.Set("page", fmt.Sprintf("%d", req.Page))
	args.Set("pagesize", "10")
	args.Set("site", s.apiSite)
	args.Set("sort", "activity")
	args.Set("order", "desc")

	resp, err := s.client.R().SetContext(ctx).Get(searchAPI + args.Encode())
	if err != nil {
		return nil, fmt.Errorf("stackexchange request failed: %w", err)
	}

	var payload struct {
		Items []struct {
			QuestionID int      `json:"question_id"`
			Title      string   `json:"title"`
			Tags       []string `json:"tags"`
			Score      int      `json:"score"`
			IsAnswered bool     `json:"is_answered"`
			Owner      struct {
				DisplayName string `json:"display_name"`
			} `json:"owner"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("stackexchange parse: %w", err)
	}

	var results []models.Result
	for _, item := range payload.Items {
		if item.QuestionID == 0 {
			continue
		}
		content := fmt.Sprintf("[%s] %s", strings.Join(item.Tags, ", "), item.Owner.DisplayName)
		if item.IsAnswered {
			content += " // is answered"
		}
		content += fmt.Sprintf(" // score: %d", item.Score)

		results = append(results, models.Result{
			Title:    html.UnescapeString(item.Title),
			URL:      fmt.Sprintf("https://%s.com/q/%d", s.apiSite, item.QuestionID),
			Content:  html.UnescapeString(content),
			Engine:   s.Name(),
			Category: req.Category,
			Template: "default",
		})
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results,
	}, nil
}
