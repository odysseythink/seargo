package plugin

import (
	"encoding/gob"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	// Register concrete types used in RPC so gob can encode/decode interface values.
	gob.Register(SearchContext{})
	gob.Register(models.Result{})
	gob.Register([]models.Result{})
	gob.Register(map[string]any{})
}

// HandshakeConfig is used by both host and plugin to confirm the protocol.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SEARGO_PLUGIN",
	MagicCookieValue: "seargo-external-plugin",
}

// ExternalPlugin is the interface that third-party plugin authors implement.
type ExternalPlugin interface {
	ID() string
	Info() PluginInfo
	Init(configSnapshot map[string]any) bool
	PreSearch(ctx SearchContext) bool
	OnResult(ctx SearchContext, r models.Result) bool
	PostSearch(ctx SearchContext) []models.Result
}

// Args / Reply structs for net/rpc. Field names are exported for gob.

type InitArgs struct {
	ConfigSnapshot map[string]any
}

type InitReply struct {
	Ok bool
}

type PreSearchArgs struct {
	Ctx SearchContext
}

type PreSearchReply struct {
	Ok bool
}

type OnResultArgs struct {
	Ctx    SearchContext
	Result models.Result
}

type OnResultReply struct {
	Keep bool
}

type PostSearchArgs struct {
	Ctx SearchContext
}

type PostSearchReply struct {
	Results []models.Result
}

type InfoArgs struct{}

type InfoReply struct {
	Info PluginInfo
}

// ExternalPluginRPC is the net/rpc service exposed by the plugin process.
type ExternalPluginRPC struct {
	Impl ExternalPlugin
}

func (s *ExternalPluginRPC) Info(args *InfoArgs, reply *InfoReply) error {
	reply.Info = s.Impl.Info()
	return nil
}

func (s *ExternalPluginRPC) Init(args *InitArgs, reply *InitReply) error {
	reply.Ok = s.Impl.Init(args.ConfigSnapshot)
	return nil
}

func (s *ExternalPluginRPC) PreSearch(args *PreSearchArgs, reply *PreSearchReply) error {
	reply.Ok = s.Impl.PreSearch(args.Ctx)
	return nil
}

func (s *ExternalPluginRPC) OnResult(args *OnResultArgs, reply *OnResultReply) error {
	reply.Keep = s.Impl.OnResult(args.Ctx, args.Result)
	return nil
}

func (s *ExternalPluginRPC) PostSearch(args *PostSearchArgs, reply *PostSearchReply) error {
	reply.Results = s.Impl.PostSearch(args.Ctx)
	return nil
}

// ExternalPluginPlugin implements hashicorp/go-plugin's plugin.Plugin interface.
type ExternalPluginPlugin struct {
	Impl ExternalPlugin
}

func (p *ExternalPluginPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &ExternalPluginRPC{Impl: p.Impl}, nil
}

func (p *ExternalPluginPlugin) Client(broker *plugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	return client, nil
}
