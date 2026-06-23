package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	engine.Register("docker_hub", &DockerHub{})
}

var apiBaseURL = "https://hub.docker.com"

// DockerHub searches Docker Hub's v3 catalog API.
type DockerHub struct {
	client *httpx.Client
}

func (d *DockerHub) Name() string { return "docker_hub" }

func (d *DockerHub) Categories() []models.Category {
	return []models.Category{models.CategoryIT}
}

func (d *DockerHub) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://hub.docker.com",
		WikidataID: "Q100769064",
	}
}

func (d *DockerHub) Capabilities() engine.Capabilities {
	return engine.Capabilities{SupportsPagination: true}
}

func (d *DockerHub) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (d *DockerHub) Setup(cfg engine.EngineInitConfig) bool {
	d.client = cfg.Client
	return d.client != nil
}

func (d *DockerHub) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	const pageSize = 10
	from := pageSize * (req.Page - 1)
	if from < 0 {
		from = 0
	}

	args := url.Values{}
	args.Set("query", req.Query)
	args.Set("from", fmt.Sprintf("%d", from))
	args.Set("size", fmt.Sprintf("%d", pageSize))
	searchURL := apiBaseURL + "/api/search/v3/catalog/search?" + args.Encode()

	resp, err := d.client.R().SetContext(ctx).Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("docker_hub request failed: %w", err)
	}

	var payload struct {
		Results []struct {
			Name             string `json:"name"`
			Slug             string `json:"slug"`
			Source           string `json:"source"`
			ShortDescription string `json:"short_description"`
			LogoURL          struct {
				Large string `json:"large"`
				Small string `json:"small"`
			} `json:"logo_url"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("docker_hub parse: %w", err)
	}

	var results []models.Result
	for _, item := range payload.Results {
		if item.Name == "" || item.Slug == "" {
			continue
		}
		isOfficial := item.Source == "store" || item.Source == "official"
		prefix := "/r/"
		if isOfficial {
			prefix = "/_/"
		}

		thumb := item.LogoURL.Large
		if thumb == "" {
			thumb = item.LogoURL.Small
		}

		results = append(results, models.Result{
			Title:        item.Name,
			URL:          apiBaseURL + prefix + item.Slug,
			Content:      item.ShortDescription,
			ThumbnailURL: thumb,
			Engine:       d.Name(),
			Category:     req.Category,
			Template:     "default",
		})
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results,
	}, nil
}
