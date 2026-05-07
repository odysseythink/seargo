package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/pkg/models"
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
}

func (s *Server) handleSearch(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, models.Response{
		Query:   req.Query,
		Results: []models.Result{},
	})
}

func (s *Server) handleEngines(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"engines": []any{}})
}

func (s *Server) handleCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"categories": []string{"general", "images", "videos", "news"},
	})
}

func (s *Server) handleConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"default_language":   s.config.Search.DefaultLang,
		"default_category":   s.config.Search.DefaultCategory,
		"safe_search":        s.config.Search.SafeSearch,
		"autocomplete":       s.config.Search.Autocomplete,
		"max_results":        s.config.Search.MaxResults,
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}
