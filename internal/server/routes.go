package server

import (
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/i18n"
	"github.com/seargo/seargo/internal/middleware"
	"github.com/seargo/seargo/internal/preferences"
	"github.com/seargo/seargo/internal/server/render"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/web"
)

func (s *Server) setupRoutes() {
	// SearXNG-compatible search endpoint: /search is the canonical HTML/JSON
	// search path; /api/search remains available for the React frontend.
	s.router.GET("/search", s.handleSearch)

	api := s.router.Group("/api")
	{
		api.GET("/search", s.handleSearch)
		api.GET("/engines", s.handleEngines)
		api.GET("/categories", s.handleCategories)
		api.GET("/config", s.handleConfig)
		api.GET("/preferences", s.handleGetPreferences)
		api.PUT("/preferences", s.handlePutPreferences)
		api.GET("/preferences/export", s.handleExportPreferences)
		api.GET("/preferences/import", s.handleImportPreferences)
		api.GET("/autocomplete", s.handleAutocomplete)
		api.GET("/opensearch.xml", s.handleOpenSearch)
	}

	s.router.GET("/health", s.handleHealth)
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Stats endpoints (gated by enable_metrics)
	if s.config.General.EnableMetrics && s.enginesStatsStore != nil {
		s.router.GET("/api/stats/engines", s.handleStatsEngines)
		s.router.GET("/api/stats/errors", s.handleStatsErrors)
	}

	s.router.GET("/robots.txt", middleware.HandleRobotsTxt)

	// Phase 8 proxy + anti-abuse endpoints
	s.router.GET("/image_proxy", s.handleImageProxy)
	s.router.GET("/favicon_proxy", s.handleFaviconProxy)
	s.router.GET("/link_token", s.handleLinkToken)

	// Static files (React frontend)
	dist, err := fs.Sub(web.Dist, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(&spaFileSystem{fs: dist}))
		s.router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/") || path == "/health" || path == "/metrics" {
				return
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}
}

// spaFileSystem wraps an embed.FS and falls back to index.html for client-side
// routes of the React single-page application. Actual static assets that are
// missing still return their original error so broken assets do not silently
// serve the HTML shell.
type spaFileSystem struct {
	fs fs.FS
}

func (s *spaFileSystem) Open(name string) (fs.File, error) {
	f, err := s.fs.Open(name)
	if err == nil {
		return f, nil
	}
	if strings.HasPrefix(name, "assets/") ||
		name == "favicon.svg" ||
		name == "icons.svg" ||
		strings.HasPrefix(name, "locales/") {
		return nil, err
	}
	return s.fs.Open("index.html")
}

func (s *Server) handleSearch(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Use cookie preferences for defaults, with request param override
	prefs := preferences.CtxPreferences(c)
	defaultLang := s.config.Search.DefaultLang
	defaultCategory := models.Category(s.config.Search.DefaultCategory)
	if prefs != nil {
		if prefs.Language != "" {
			defaultLang = prefs.Language
		}
		if len(prefs.Categories) > 0 {
			defaultCategory = models.Category(prefs.Categories[0])
		}
	}

	// Resolve locale and enabled plugins for plugins/answerers
	if req.Locale == "" {
		if prefs != nil && prefs.Locale != "" {
			req.Locale = prefs.Locale
		} else {
			req.Locale = defaultLang
		}
	}
	req.EnabledPlugins = s.computeEnabledPlugins(prefs)

	req.Normalize(models.NormalizeDefaults{
		DefaultLang:     defaultLang,
		DefaultCategory: defaultCategory,
		DefaultPageSize: s.config.Search.MaxResults,
		MaxResults:      s.config.Search.MaxResults,
	})

	// Resolve output format from query param, Accept header, and config whitelist
	format, fmtErr := render.ResolveFormat(c.Query("format"), c.GetHeader("Accept"), s.config.Search.Formats)
	if fmtErr != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": fmtErr.Error()})
		return
	}

	resp, err := s.scheduler.Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// Phase 8: rewrite image URLs through proxy
	if s.config.Server.ImageProxy && s.imageProxy != nil {
		for i := range resp.Results {
			r := &resp.Results[i]
			// Rewrite ThumbnailURL for all result types
			if r.ThumbnailURL != "" {
				if proxyURL, err := s.imageProxy.SignedURL(r.ThumbnailURL); err == nil {
					r.ThumbnailURL = proxyURL
				}
			}
			// Rewrite type-specific URLs stored in Extra
			if r.Extra != nil {
				if src, ok := r.Extra["img_src"]; ok {
					if str, ok := src.(string); ok && str != "" {
						if proxyURL, err := s.imageProxy.SignedURL(str); err == nil {
							r.Extra["img_src"] = proxyURL
						}
					}
				}
				if src, ok := r.Extra["thumbnail_src"]; ok {
					if str, ok := src.(string); ok && str != "" {
						if proxyURL, err := s.imageProxy.SignedURL(str); err == nil {
							r.Extra["thumbnail_src"] = proxyURL
						}
					}
				}
				if src, ok := r.Extra["thumbnail"]; ok {
					if str, ok := src.(string); ok && str != "" {
						if proxyURL, err := s.imageProxy.SignedURL(str); err == nil {
							r.Extra["thumbnail"] = proxyURL
						}
					}
				}
			}
		}
	}

	// Phase 8: rewrite favicon URLs through favicon proxy
	if s.config.Search.FaviconResolver != "" && s.favSvc != nil {
		for i := range resp.Results {
			r := &resp.Results[i]
			r.Favicon = s.favSvc.RewriteFaviconURL(r.URL, r.Favicon)
		}
	}

	// Render response in the requested format
	if format != render.FormatJSON && format != render.FormatHTML {
		data, ct, renderErr := render.Render(resp, format, s.baseURL(c))
		if renderErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "render failed: " + renderErr.Error()})
			return
		}
		c.Data(http.StatusOK, ct, data)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// computeEnabledPlugins returns the plugin IDs that should run for this request.
// It starts from all registered plugins, applies the configured active flags,
// and finally applies user preference enabled/disabled overrides.
func (s *Server) computeEnabledPlugins(prefs *preferences.UserPreferences) []string {
	if s.scheduler == nil {
		return nil
	}
	all := s.scheduler.PluginIDs()
	enabled := make(map[string]bool, len(all))
	for _, id := range all {
		if pc, ok := s.config.Plugins[id]; ok && !pc.Active {
			continue
		}
		enabled[id] = true
	}

	if prefs != nil {
		if len(prefs.EnabledPlugins) > 0 {
			filtered := make(map[string]bool, len(prefs.EnabledPlugins))
			for _, id := range prefs.EnabledPlugins {
				if enabled[id] {
					filtered[id] = true
				}
			}
			enabled = filtered
		}
		for _, id := range prefs.DisabledPlugins {
			delete(enabled, id)
		}
	}

	result := make([]string, 0, len(enabled))
	for id := range enabled {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
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
		Version     string                    `json:"version"`
		Brand       brandConfigResponse       `json:"brand"`
		General     generalConfigResponse     `json:"general"`
		Search      searchConfigResponse      `json:"search"`
		Server      serverConfigResponse      `json:"server"`
		UI          uiConfigResponse          `json:"ui"`
		Preferences preferencesConfigResponse `json:"preferences"`
	}

	// Negotiate locale from Accept-Language and config
	var locale string
	var rtl bool
	if s.localeRegistry != nil {
		neg := i18n.NewNegotiator(s.localeRegistry)
		locale = neg.Negotiate(c.Request.Header.Get("Accept-Language"), "", s.config.UI.DefaultLocale)
		rtl = s.localeRegistry.IsRTL(locale)
	} else {
		locale = s.config.UI.DefaultLocale
		if locale == "" {
			locale = "en"
		}
	}

	resp := configResponse{
		Version: "1.0.0",
		Brand: brandConfigResponse{
			IssueURL:        s.config.Brand.IssueURL,
			DocsURL:         s.config.Brand.DocsURL,
			PublicInstances: s.config.Brand.PublicInstances,
			WikiURL:         s.config.Brand.WikiURL,
			NewIssueURL:     s.config.Brand.NewIssueURL,
			PWAColors: pwaColorsConfigResponse{
				ThemeColorLight:      s.config.Brand.PWAColors.ThemeColorLight,
				BackgroundColorLight: s.config.Brand.PWAColors.BackgroundColorLight,
				ThemeColorDark:       s.config.Brand.PWAColors.ThemeColorDark,
				BackgroundColorDark:  s.config.Brand.PWAColors.BackgroundColorDark,
				ThemeColorBlack:      s.config.Brand.PWAColors.ThemeColorBlack,
				BackgroundColorBlack: s.config.Brand.PWAColors.BackgroundColorBlack,
			},
			Links: s.config.Brand.Custom.Links,
		},
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
			DefaultLocale:          locale,
			RTL:                    rtl,
			CenterAlignment:        s.config.UI.CenterAlignment,
			ResultsOnNewTab:        s.config.UI.ResultsOnNewTab,
			QueryInTitle:           s.config.UI.QueryInTitle,
			SearchOnCategorySelect: s.config.UI.SearchOnCategorySelect,
			Hotkeys:                s.config.UI.Hotkeys,
			URLFormatting:          s.config.UI.URLFormatting,
			SimpleStyle:            s.config.UI.ThemeArgs.SimpleStyle,
		},
		Preferences: preferencesConfigResponse{
			Lock:   s.config.Preferences.Lock,
			Themes: availableThemes(s.config.UI.ThemeArgs.SimpleStyle),
		},
	}

	c.JSON(http.StatusOK, resp)
}

type brandConfigResponse struct {
	IssueURL        string                  `json:"issue_url"`
	DocsURL         string                  `json:"docs_url"`
	PublicInstances string                  `json:"public_instances"`
	WikiURL         string                  `json:"wiki_url"`
	NewIssueURL     string                  `json:"new_issue_url"`
	PWAColors       pwaColorsConfigResponse `json:"pwa_colors"`
	Links           map[string]string       `json:"links"`
}

type pwaColorsConfigResponse struct {
	ThemeColorLight      string `json:"theme_color_light"`
	BackgroundColorLight string `json:"background_color_light"`
	ThemeColorDark       string `json:"theme_color_dark"`
	BackgroundColorDark  string `json:"background_color_dark"`
	ThemeColorBlack      string `json:"theme_color_black"`
	BackgroundColorBlack string `json:"background_color_black"`
}

func availableThemes(simpleStyle string) []string {
	if simpleStyle != "" && simpleStyle != "auto" {
		return []string{simpleStyle}
	}
	return []string{"auto", "light", "dark", "black"}
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
	RTL                    bool   `json:"rtl"`
	CenterAlignment        bool   `json:"center_alignment"`
	ResultsOnNewTab        bool   `json:"results_on_new_tab"`
	QueryInTitle           bool   `json:"query_in_title"`
	SearchOnCategorySelect bool   `json:"search_on_category_select"`
	Hotkeys                string `json:"hotkeys"`
	URLFormatting          string `json:"url_formatting"`
	SimpleStyle            string `json:"simple_style"`
}

type preferencesConfigResponse struct {
	Lock   []string `json:"lock"`
	Themes []string `json:"themes"`
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

func (s *Server) handleImageProxy(c *gin.Context) {
	rawURL := c.Query("url")
	signature := c.Query("h")
	if s.imageProxy == nil {
		c.JSON(400, gin.H{"error": "image proxy not configured"})
		return
	}
	if err := s.imageProxy.Serve(c.Request.Context(), rawURL, signature, c.Writer); err != nil {
		return
	}
}

func (s *Server) handleFaviconProxy(c *gin.Context) {
	resolver := c.Query("resolver")
	authority := c.Query("authority")
	signature := c.Query("h")
	if s.favSvc == nil {
		c.JSON(400, gin.H{"error": "favicon proxy not configured"})
		return
	}
	data, mime, err := s.favSvc.Serve(c.Request.Context(), resolver, authority, signature)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", "public, max-age=86400")
		c.String(200, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"></svg>`)
		return
	}
	c.Header("Content-Type", mime)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Writer.Write(data)
}

func (s *Server) handleLinkToken(c *gin.Context) {
	if s.limiterSvc == nil {
		c.JSON(503, gin.H{"error": "limiter not configured"})
		return
	}
	token, err := s.limiterSvc.Token(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token})
}
