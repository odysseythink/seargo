package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/preferences"
)

// PluginPrefItem describes a plugin in the preferences response.
type PluginPrefItem struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Active            bool     `json:"active"`
	PreferenceSection string   `json:"preference_section"`
	Examples          []string `json:"examples,omitempty"`
}

// AnswererPrefItem describes an answerer in the preferences response.
type AnswererPrefItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Active      bool     `json:"active"`
	Keywords    []string `json:"keywords"`
	Examples    []string `json:"examples,omitempty"`
}

// LocaleOption is a minimal locale info for the preferences response.
type LocaleOption struct {
	Tag  string `json:"tag"`
	Name string `json:"name"`
}

// PreferencesResponse is the GET /api/preferences response.
type PreferencesResponse struct {
	Plugins      []PluginPrefItem            `json:"plugins"`
	Answerers    []AnswererPrefItem          `json:"answerers"`
	Autocomplete string                      `json:"autocomplete"`
	Settings     preferences.UserPreferences `json:"settings"`
	Categories   []string                    `json:"categories"`
	Themes       []string                    `json:"themes"`
	Locales      []LocaleOption              `json:"locales"`
	DOIResolvers []string                    `json:"doi_resolvers"`
}

func (s *Server) handleGetPreferences(c *gin.Context) {
	prefs := preferences.CtxPreferences(c)

	resp := PreferencesResponse{
		Autocomplete: prefs.Autocomplete,
		Settings:     *prefs,
		Categories:   []string{"general", "images", "videos", "news", "map", "music", "it", "science", "files", "social_media"},
		Themes:       []string{"simple"},
		Locales: []LocaleOption{
			{Tag: "en", Name: "English"},
			{Tag: "zh-CN", Name: "简体中文"},
		},
		DOIResolvers: []string{},
	}

	if ps := plugin.GlobalPlugin(); ps != nil {
		for _, p := range ps.All() {
			info := p.Info()
			resp.Plugins = append(resp.Plugins, PluginPrefItem{
				ID:                p.ID(),
				Name:              info.Name,
				Description:       info.Description,
				Active:            !isPluginDisabled(p.ID(), prefs.DisabledPlugins),
				PreferenceSection: info.PreferenceSection,
				Examples:          info.Examples,
			})
		}
	}

	if as := answerer.GlobalAnswerer(); as != nil {
		for _, a := range as.All() {
			info := a.Info()
			id := ""
			if kw := a.Keywords(); len(kw) > 0 {
				id = kw[0]
			} else {
				id = info.Name
			}
			resp.Answerers = append(resp.Answerers, AnswererPrefItem{
				ID:          id,
				Name:        info.Name,
				Description: info.Description,
				Active:      !isPluginDisabled(id, prefs.DisabledPlugins),
				Keywords:    info.Keywords,
				Examples:    info.Examples,
			})
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handlePutPreferences(c *gin.Context) {
	var update preferences.PreferencesUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	current := preferences.CtxPreferences(c)
	next, err := s.preferencesStore.ApplyUpdate(current, update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.preferencesStore.WriteCookie(next, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := PreferencesResponse{
		Autocomplete: next.Autocomplete,
		Settings:     *next,
		Categories:   []string{"general", "images", "videos", "news", "map", "music", "it", "science", "files", "social_media"},
		Themes:       []string{"simple"},
		Locales: []LocaleOption{
			{Tag: "en", Name: "English"},
			{Tag: "zh-CN", Name: "简体中文"},
		},
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleExportPreferences(c *gin.Context) {
	prefs := preferences.CtxPreferences(c)
	blob, err := s.preferencesStore.ExportURL(prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, blob)
}

func (s *Server) handleImportPreferences(c *gin.Context) {
	blob := c.Query("blob")
	if blob == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing blob parameter"})
		return
	}
	next, err := s.preferencesStore.ImportURL(blob)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.preferencesStore.WriteCookie(next, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, next)
}

func isPluginDisabled(pluginID string, disabled []string) bool {
	for _, d := range disabled {
		if d == pluginID {
			return true
		}
	}
	return false
}
