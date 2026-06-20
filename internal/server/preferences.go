package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/internal/plugin"
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

// PreferencesResponse is the GET /api/preferences response.
type PreferencesResponse struct {
	Plugins   []PluginPrefItem   `json:"plugins"`
	Answerers []AnswererPrefItem `json:"answerers"`
}

// PreferencesUpdate is the PUT /api/preferences request body.
type PreferencesUpdate struct {
	Plugins   map[string]bool `json:"plugins"`
	Answerers map[string]bool `json:"answerers"`
}

func (s *Server) handleGetPreferences(c *gin.Context) {
	var resp PreferencesResponse

	if ps := plugin.GlobalPlugin(); ps != nil {
		for _, p := range ps.All() {
			info := p.Info()
			active := false
			if cfg, ok := s.config.Plugins[p.ID()]; ok {
				active = cfg.Active
			}
			resp.Plugins = append(resp.Plugins, PluginPrefItem{
				ID:                p.ID(),
				Name:              info.Name,
				Description:       info.Description,
				Active:            active,
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
			active := false
			if cfg, ok := s.config.Answerers[id]; ok {
				active = cfg.Active
			}
			resp.Answerers = append(resp.Answerers, AnswererPrefItem{
				ID:          id,
				Name:        info.Name,
				Description: info.Description,
				Active:      active,
				Keywords:    info.Keywords,
				Examples:    info.Examples,
			})
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handlePutPreferences(c *gin.Context) {
	var update PreferencesUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for id, active := range update.Plugins {
		if cfg, ok := s.config.Plugins[id]; ok {
			cfg.Active = active
			s.config.Plugins[id] = cfg
		}
	}

	for id, active := range update.Answerers {
		if cfg, ok := s.config.Answerers[id]; ok {
			cfg.Active = active
			s.config.Answerers[id] = cfg
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
