package plugin

import (
	"fmt"
	"net/rpc"
	"os"
	"sync/atomic"

	"github.com/hashicorp/go-plugin"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/pkg/models"
)

// externalPluginAdapter wraps an external plugin process so it satisfies the
// internal Plugin interface. Hook calls are forwarded over RPC.
type externalPluginAdapter struct {
	id        string
	info      PluginInfo
	client    *plugin.Client
	rpc       *rpc.Client
	available atomic.Bool
}

var _ Plugin = (*externalPluginAdapter)(nil)

// newExternalPluginAdapter creates an adapter. It expects info to have been
// retrieved via RPC already (the loader calls Info before registration).
func newExternalPluginAdapter(id string, info PluginInfo, client *plugin.Client, rpcClient *rpc.Client) *externalPluginAdapter {
	a := &externalPluginAdapter{
		id:     id,
		info:   info,
		client: client,
		rpc:    rpcClient,
	}
	a.available.Store(true)
	return a
}

func (a *externalPluginAdapter) ID() string       { return a.id }
func (a *externalPluginAdapter) Info() PluginInfo { return a.info }

func (a *externalPluginAdapter) Init(appCtx *AppContext) bool {
	snapshot := extractPluginExtra(appCtx, a.id)
	args := InitArgs{ConfigSnapshot: snapshot}
	var reply InitReply
	if err := a.call("Plugin.Init", &args, &reply); err != nil {
		fmt.Fprintf(os.Stderr, "[seargo] plugin Init RPC failed %q: %v\n", a.id, err)
		return false
	}
	return reply.Ok
}

func (a *externalPluginAdapter) PreSearch(ctx *SearchContext) bool {
	if !a.available.Load() {
		return true
	}
	args := PreSearchArgs{Ctx: *ctx}
	var reply PreSearchReply
	if err := a.call("Plugin.PreSearch", &args, &reply); err != nil {
		fmt.Fprintf(os.Stderr, "[seargo] plugin PreSearch RPC failed %q: %v\n", a.id, err)
		return true
	}
	return reply.Ok
}

func (a *externalPluginAdapter) OnResult(ctx *SearchContext, r *models.Result) bool {
	if !a.available.Load() {
		return true
	}
	args := OnResultArgs{Ctx: *ctx, Result: *r}
	var reply OnResultReply
	if err := a.call("Plugin.OnResult", &args, &reply); err != nil {
		fmt.Fprintf(os.Stderr, "[seargo] plugin OnResult RPC failed %q: %v\n", a.id, err)
		return true
	}
	return reply.Keep
}

func (a *externalPluginAdapter) PostSearch(ctx *SearchContext) []models.Result {
	if !a.available.Load() {
		return nil
	}
	args := PostSearchArgs{Ctx: *ctx}
	var reply PostSearchReply
	if err := a.call("Plugin.PostSearch", &args, &reply); err != nil {
		fmt.Fprintf(os.Stderr, "[seargo] plugin PostSearch RPC failed %q: %v\n", a.id, err)
		return nil
	}
	return reply.Results
}

// Close closes the RPC connection and kills the plugin child process.
func (a *externalPluginAdapter) Close() error {
	if a.rpc != nil {
		_ = a.rpc.Close()
	}
	if a.client != nil {
		a.client.Kill()
	}
	return nil
}

func (a *externalPluginAdapter) markUnavailable() {
	a.available.Store(false)
}

// call wraps the RPC call with recover so a panic in encoding/decoding does
// not crash the host. On any error the plugin is marked unavailable.
func (a *externalPluginAdapter) call(method string, args, reply any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
		if err != nil {
			a.markUnavailable()
		}
	}()
	return a.rpc.Call(method, args, reply)
}

// extractPluginExtra returns only the plugin's own Extra configuration.
// It intentionally does not expose the global Config tree (e.g. secret_key).
func extractPluginExtra(appCtx *AppContext, id string) map[string]any {
	if appCtx == nil || appCtx.Config == nil {
		return nil
	}
	cfg, ok := appCtx.Config.(*config.Config)
	if !ok {
		return nil
	}
	pc, ok := cfg.Plugins[id]
	if !ok {
		return nil
	}
	return pc.Extra
}
