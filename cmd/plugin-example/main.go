package main

import (
	"strings"

	"github.com/hashicorp/go-plugin"

	seargoplugin "github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

// EchoPlugin is a minimal example of an external SearGo plugin.
type EchoPlugin struct{}

func (e *EchoPlugin) ID() string { return "echo" }

func (e *EchoPlugin) Info() seargoplugin.PluginInfo {
	return seargoplugin.PluginInfo{
		ID:          "echo",
		Name:        "Echo",
		Description: "Returns an answer result echoing the query.",
	}
}

func (e *EchoPlugin) Init(configSnapshot map[string]any) bool {
	_ = configSnapshot
	return true
}

func (e *EchoPlugin) PreSearch(ctx seargoplugin.SearchContext) bool {
	return true
}

func (e *EchoPlugin) OnResult(ctx seargoplugin.SearchContext, r models.Result) bool {
	return true
}

func (e *EchoPlugin) PostSearch(ctx seargoplugin.SearchContext) []models.Result {
	prefix := ""
	if ctx.Query != "" {
		prefix = strings.ToUpper(ctx.Query[:1]) + ctx.Query[1:]
	}
	return []models.Result{
		{
			Kind:    "answer",
			Title:   "Echo",
			Content: prefix,
			Engine:  "echo",
		},
	}
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: seargoplugin.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"external_plugin": &seargoplugin.ExternalPluginPlugin{Impl: &EchoPlugin{}},
		},
	})
}
