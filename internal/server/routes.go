package server

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/seargo/seargo/internal/config"
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
		c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	req.Normalize(models.NormalizeDefaults{
		DefaultLang:     s.config.Search.DefaultLang,
		DefaultCategory: models.Category(s.config.Search.DefaultCategory),
		DefaultPageSize: s.config.Search.MaxResults,
		MaxResults:      s.config.Search.MaxResults,
	})

	resp, err := s.scheduler.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
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
		caps := e.Capabilities()

		enabled := true
		shortcut := ""
		if ec, ok := s.configEngineConfigs()[name]; ok {
			if ec.Enabled {
				enabled = true
			} else if ec.Disabled {
				enabled = false
			} else {
				enabled = true
			}
			shortcut = ec.Shortcut
		}
		caps.Shortcut = shortcut

		infos = append(infos, engine.Info{
			Name:         name,
			Categories:   cats,
			Shortcut:     shortcut,
			Capabilities: caps,
			Enabled:      enabled,
		})
	}
	c.JSON(http.StatusOK, gin.H{"engines": infos})
}

func (s *Server) configEngineConfigs() map[string]config.EngineConfig {
	result := make(map[string]config.EngineConfig, len(s.config.Engines))
	for _, ec := range s.config.Engines {
		key := ec.Engine
		if key == "" {
			key = ec.Name
		}
		result[key] = ec
	}
	return result
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
	type configResponse struct {
		General     generalConfigResponse     `json:"general"`
		Search      searchConfigResponse      `json:"search"`
		Server      serverConfigResponse      `json:"server"`
		UI          uiConfigResponse          `json:"ui"`
		Preferences preferencesConfigResponse `json:"preferences"`
	}

	resp := configResponse{
		General: generalConfigResponse{
			InstanceName:  s.config.General.InstanceName,
			Debug:         s.config.General.Debug,
			EnableMetrics: s.config.General.EnableMetrics,
			DonationURL:   s.config.General.DonationURL,
		},
		Search: searchConfigResponse{
			DefaultLanguage: s.config.Search.DefaultLang,
			DefaultCategory: s.config.Search.DefaultCategory,
			SafeSearch:      s.config.Search.SafeSearch,
			Autocomplete:    s.config.Search.Autocomplete,
			AutocompleteMin: s.config.Search.AutocompleteMin,
			MaxResults:      s.config.Search.MaxResults,
			Formats:         s.config.Search.Formats,
		},
		Server: serverConfigResponse{
			PublicInstance:      s.config.Server.PublicInstance,
			HTTPProtocolVersion: s.config.Server.HTTPProtocolVersion,
			Method:              s.config.Server.Method,
			ImageProxy:          s.config.Server.ImageProxy,
			Limiter:             s.config.Server.Limiter,
		},
		UI: uiConfigResponse{
			DefaultTheme:           s.config.UI.DefaultTheme,
			DefaultLocale:          s.config.UI.DefaultLocale,
			CenterAlignment:        s.config.UI.CenterAlignment,
			ResultsOnNewTab:        s.config.UI.ResultsOnNewTab,
			QueryInTitle:           s.config.UI.QueryInTitle,
			SearchOnCategorySelect: s.config.UI.SearchOnCategorySelect,
			Hotkeys:                s.config.UI.Hotkeys,
			URLFormatting:          s.config.UI.URLFormatting,
			SimpleStyle:            s.config.UI.ThemeArgs.SimpleStyle,
		},
		Preferences: preferencesConfigResponse{
			Lock: s.config.Preferences.Lock,
		},
	}

	c.JSON(http.StatusOK, resp)
}

type generalConfigResponse struct {
	InstanceName  string `json:"instance_name"`
	Debug         bool   `json:"debug"`
	EnableMetrics bool   `json:"enable_metrics"`
	DonationURL   string `json:"donation_url,omitempty"`
}

type searchConfigResponse struct {
	DefaultLanguage string   `json:"default_language"`
	DefaultCategory string   `json:"default_category"`
	SafeSearch      int      `json:"safe_search"`
	Autocomplete    string   `json:"autocomplete"`
	AutocompleteMin int      `json:"autocomplete_min"`
	MaxResults      int      `json:"max_results"`
	Formats         []string `json:"formats"`
}

type serverConfigResponse struct {
	PublicInstance      bool   `json:"public_instance"`
	HTTPProtocolVersion string `json:"http_protocol_version"`
	Method              string `json:"method"`
	ImageProxy          bool   `json:"image_proxy"`
	Limiter             bool   `json:"limiter"`
}

type uiConfigResponse struct {
	DefaultTheme           string `json:"default_theme"`
	DefaultLocale          string `json:"default_locale"`
	CenterAlignment        bool   `json:"center_alignment"`
	ResultsOnNewTab        bool   `json:"results_on_new_tab"`
	QueryInTitle           bool   `json:"query_in_title"`
	SearchOnCategorySelect bool   `json:"search_on_category_select"`
	Hotkeys                string `json:"hotkeys"`
	URLFormatting          string `json:"url_formatting"`
	SimpleStyle            string `json:"simple_style"`
}

type preferencesConfigResponse struct {
	Lock []string `json:"lock"`
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}
