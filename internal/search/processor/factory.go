package processor

import (
	"fmt"

	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
)

// NewProcessorFromConfig 根据引擎配置创建对应的 Processor。
// 目前所有在线引擎统一使用 OnlineProcessor；离线引擎使用 OfflineProcessor。
// 特殊类型（Currency、Dictionary、URLSearch）后续版本通过插件注册。
func NewProcessorFromConfig(eng engine.Engine, ec config.EngineConfig, suspension Suspension, client *httpx.Client) (Processor, error) {
	if eng == nil {
		return nil, fmt.Errorf("engine is nil for %s", ec.Name)
	}
	return NewOnlineProcessor(eng, suspension, client), nil
}
