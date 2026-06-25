# SearGo 外部插件开发指南

SearGo 使用 [`hashicorp/go-plugin`](https://github.com/hashicorp/go-plugin) 的 `net/rpc + gob` 协议加载第三方插件。第三方插件是**独立可执行文件**，不再是 `.so` 共享库。

## 目录与启用

1. 在 `configs/settings.yml` 中配置 `plugin_dir`：
   ```yaml
   plugin_dir: /var/lib/seargo/plugins
   plugins:
     echo:
       active: true
       prefix: "> "   # 插件自定义配置，会传入 Init
   ```
2. 将插件可执行文件放入 `plugin_dir`：
   - Linux/macOS: `<id>`（如 `echo`）
   - Windows: `<id>.exe`（如 `echo.exe`）
3. 插件 ID 必须匹配正则 `^[a-z][a-z0-9_]*$`。
4. 只有 `plugins.<id>.active: true` 的插件才会被加载。

## 最小插件示例

参考 `cmd/plugin-example/main.go`：

```go
package main

import (
    "github.com/hashicorp/go-plugin"
    seargoplugin "github.com/seargo/seargo/internal/plugin"
    "github.com/seargo/seargo/pkg/models"
)

type MyPlugin struct{}

func (m *MyPlugin) ID() string { return "myplugin" }
func (m *MyPlugin) Info() seargoplugin.PluginInfo {
    return seargoplugin.PluginInfo{ID: "myplugin", Name: "My Plugin"}
}
func (m *MyPlugin) Init(configSnapshot map[string]any) bool { return true }
func (m *MyPlugin) PreSearch(ctx seargoplugin.SearchContext) bool { return true }
func (m *MyPlugin) OnResult(ctx seargoplugin.SearchContext, r models.Result) bool { return true }
func (m *MyPlugin) PostSearch(ctx seargoplugin.SearchContext) []models.Result {
    return nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: seargoplugin.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "external_plugin": &seargoplugin.ExternalPluginPlugin{Impl: &MyPlugin{}},
        },
    })
}
```

## 接口说明

- `ID()` / `Info()`：标识与元数据。
- `Init(configSnapshot)`：仅接收该插件在 `settings.yml` 中的 `Extra` 配置（即 `plugins.<id>` 下除 `active` 外的字段），不会收到全局配置。
- `PreSearch(ctx)`：返回 `false` 可中止搜索。
- `OnResult(ctx, result)`：返回 `false` 可丢弃该结果。
- `PostSearch(ctx)`：返回要追加到结果列表的附加结果。

## 迁移说明（从旧 `.so` 插件）

1. 将插件改为 `package main` 的可执行文件。
2. 导入 `github.com/hashicorp/go-plugin` 和 `github.com/seargo/seargo/internal/plugin`。
3. 实现 `seargoplugin.ExternalPlugin` 接口。
4. 在 `main()` 中调用 `plugin.Serve(...)`。
5. 部署时去掉 `.so` 后缀，确保文件可执行。

## 故障排查

- 插件进程启动失败会在主程序日志中输出 warning，不影响其他插件。
- RPC 调用失败时该次 hook 会返回安全默认值（`PreSearch`/`OnResult` 返回 `true`，`PostSearch` 返回 `nil`），并标记插件为不可用。
- `SearchContext` 和 `models.Result` 通过 `gob` 编码传输，插件接收到的 `map[string]any` 值必须是 gob 可编码的基本类型。
