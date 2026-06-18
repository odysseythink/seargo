package server

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/web"
)

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		api.GET("/search", s.handleSearch)
		api.GET("/engines", s.handleEngines)
		api.GET("/categories", s.handleCategories)
		api.GET("/config", s.handleConfig)
	}

	s.router.GET("/health", s.handleHealth)
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Static files (React frontend)
	dist, err := fs.Sub(web.Dist, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(dist))
		s.router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/") || path == "/health" || path == "/metrics" {
				return
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}
}

func (s *Server) handleSearch(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	if req.Category == "" {
		req.Category = models.Category(s.config.Search.DefaultCategory)
	}
	if req.PageSize <= 0 {
		req.PageSize = s.config.Search.MaxResults
	}

	resp, err := s.scheduler.Search(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleEngines(c *gin.Context) {
	allEngines := engine.All()
	var infos []engine.Info
	for name, e := range allEngines {
		cats := make([]string, len(e.Categories()))
		for i, c := range e.Categories() {
			cats[i] = string(c)
		}
		infos = append(infos, engine.Info{
			Name:         name,
			Categories:   cats,
			Capabilities: e.Capabilities(),
			Enabled:      true, // TODO: read from config
		})
	}
	c.JSON(http.StatusOK, gin.H{"engines": infos})
}

func (s *Server) handleCategories(c *gin.Context) {
	type categoryEntry struct {
		Name    string   `json:"name"`
		Engines []string `json:"engines"`
	}

	var categories []categoryEntry
	for _, cat := range models.AllCategories() {
		catName := string(cat)
		tabCfg, inTabs := s.config.CategoriesAsTabs[catName]
		if !inTabs {
			continue
		}
		engines := tabCfg.Engines
		if engines == nil {
			engines = []string{}
		}
		categories = append(categories, categoryEntry{
			Name:    catName,
			Engines: engines,
		})
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (s *Server) handleConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"default_language": s.config.Search.DefaultLang,
		"default_category": s.config.Search.DefaultCategory,
		"safe_search":      s.config.Search.SafeSearch,
		"autocomplete":     s.config.Search.Autocomplete,
		"max_results":      s.config.Search.MaxResults,
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}
