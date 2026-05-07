# SearGo 设计文档

> SearXNG 的 Go 渐进式重构方案
>
> 日期：2026-05-07

---

## 1. 概述

SearGo 是 SearXNG 元搜索引擎的 Go 语言重构版本。采用**渐进式迁移**策略，从核心引擎层开始自底向上重写，最终目标是完全替代原有 Python 后端，同时保留并现代化前端体验。

### 1.1 核心目标

- **性能提升**：利用 Go 的高并发能力（goroutine + channel）替代 Python 的线程模型
- **简化部署**：单二进制文件 + 嵌入前端产物，部署极简
- **插件化引擎**：清晰的 `Engine` 接口，后续引擎开发门槛低
- **兼容配置**：复用 SearXNG 的 `settings.yml`，降低现有用户迁移成本
- **前后端分离**：React + TypeScript 前端，RESTful API 后端

### 1.2 非目标（YAGNI）

- 不追求一次性迁移全部 222 个搜索引擎
- 不做复杂的微服务拆分（保持单体）
- 不做实时搜索引擎（WebSocket/SSE）
- 不做用户账户系统（保持匿名搜索）

---

## 2. 架构决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 迁移策略 | 渐进式 + 自底向上 | 风险可控，逐步验证 |
| 引擎迁移策略 | 核心引擎 + 插件化框架 | 先建立模式，再扩展 |
| Web 框架 | Gin | 生态成熟，性能优秀 |
| HTTP 客户端 | req / go-resty | 减少样板代码，支持高级特性 |
| 日志库 | mlog (`odysseythink/mlog`) | 用户指定，不复造轮子 |
| 配置格式 | YAML（兼容 SearXNG） | 降低迁移成本 |
| 前端架构 | 前后端分离 | 现代化，独立迭代 |
| 前端技术栈 | React 18 + Vite + TypeScript | 生态丰富，类型安全 |
| API 风格 | RESTful | 简单通用，调试方便 |
| 缓存策略 | 多级缓存（内存 + Redis） | 兼顾性能和分布式 |
| 并发控制 | 有限 goroutine 池 | 资源可控 |
| 超时策略 | 每个引擎独立超时 + 全局兜底 | 灵活且安全 |
| 失败处理 | 静默忽略失败引擎 | 高可用 |
| 结果合并 | 全部收集后统一排序 | 简单可靠 |
| 部署方式 | Docker + 二进制 | 两种都支持 |

---

## 3. 项目结构

```
seargo/
├── cmd/seargo/            # 主入口 main.go
├── internal/              # 私有包（不可外部导入）
│   ├── server/            # Gin HTTP 服务、路由、中间件
│   ├── search/            # 搜索调度器（goroutine 池、超时、合并）
│   ├── engine/            # Engine 接口定义 + 注册中心
│   ├── config/            # YAML 配置解析
│   ├── cache/             # 多级缓存（ristretto + Redis）
│   └── logger/            # mlog 封装适配
├── engines/               # 搜索引擎实现
│   ├── google/
│   ├── bing/
│   ├── duckduckgo/
│   └── ...                # 后续引擎逐个添加
├── pkg/models/            # 共享数据结构
├── web/                   # React + Vite + TypeScript 前端
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── services/
│   │   ├── stores/
│   │   └── types/
│   ├── package.json
│   └── vite.config.ts
├── configs/
│   └── settings.yml       # 配置文件示例
├── docs/                  # 文档
├── tests/                 # 集成测试 / E2E 测试
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

---

## 4. 核心模块设计

### 4.1 引擎接口（Engine Interface）

引擎是系统的核心抽象。每个搜索引擎是一个实现 `Engine` 接口的 Go 包。

```go
package engine

// Engine 是所有搜索引擎必须实现的接口
type Engine interface {
    Name() string
    Categories() []models.Category
    Capabilities() Capabilities
    Init(cfg map[string]any) error
    Search(ctx context.Context, req *models.Request) (*models.Response, error)
}

type Capabilities struct {
    SupportsSafeSearch bool
    SupportsLanguage   bool
    SupportsTimeRange  bool
    SupportsPagination bool
    RequiresAPIKey     bool
}

// 注册中心
var registry = make(map[string]Engine)

func Register(name string, e Engine) {
    registry[name] = e
}

func Get(name string) (Engine, bool) {
    e, ok := registry[name]
    return e, ok
}
```

每个引擎包在 `init()` 中自注册：

```go
// engines/google/google.go
func init() {
    engine.Register("google", &Google{})
}
```

**设计理由**：
- 接口只有 5 个方法，新增引擎门槛低
- `Capabilities` 自声明能力，调度器据此决定是否传某些参数
- `init()` 自注册，新增引擎只需 import 对应包

### 4.2 搜索调度器（Search Scheduler）

调度器负责并发查询多个引擎、收集结果、后处理。

```go
package search

type Scheduler struct {
    engines       map[string]engine.Engine
    workerPool    *ants.Pool
    cache         cache.Cache
    globalTimeout time.Duration
    engineTimeout time.Duration
    logger        *mlog.Logger
}

func (s *Scheduler) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
    // 1. 缓存检查
    if cached, ok := s.cache.Get(req.CacheKey()); ok {
        return cached, nil
    }
    
    // 2. 选择引擎
    selected := s.selectEngines(req.Category)
    
    // 3. 全局超时兜底
    ctx, cancel := context.WithTimeout(ctx, s.globalTimeout)
    defer cancel()
    
    // 4. 并发查询
    results := s.queryEngines(ctx, req, selected)
    
    // 5. 后处理（去重、排序、分页）
    response := s.postProcess(results, req)
    
    // 6. 写入缓存
    s.cache.Set(req.CacheKey(), response, s.cacheTTL(req.Category))
    
    return response, nil
}
```

**并发查询实现**：

```go
func (s *Scheduler) queryEngines(ctx context.Context, req *models.Request, engines []engine.Engine) []models.Result {
    var wg sync.WaitGroup
    resultCh := make(chan []models.Result, len(engines))
    
    for _, e := range engines {
        wg.Add(1)
        s.workerPool.Submit(func() {
            defer wg.Done()
            
            engineCtx, cancel := context.WithTimeout(ctx, s.getTimeout(e))
            defer cancel()
            
            resp, err := e.Search(engineCtx, req)
            if err != nil {
                s.logger.Warn("engine failed", "engine", e.Name(), "error", err)
                return // 静默忽略失败
            }
            resultCh <- resp.Results
        })
    }
    
    go func() { wg.Wait(); close(resultCh) }()
    
    var allResults []models.Result
    for r := range resultCh {
        allResults = append(allResults, r...)
    }
    return allResults
}
```

**去重与排序**：

```go
func (s *Scheduler) postProcess(results []models.Result, req *models.Request) *models.Response {
    deduped := deduplicate(results, s.engineWeights)
    sort.Slice(deduped, func(i, j int) bool {
        return score(deduped[i], s.engineWeights) > score(deduped[j], s.engineWeights)
    })
    if req.PageSize > 0 && len(deduped) > req.PageSize {
        deduped = deduped[:req.PageSize]
    }
    return &models.Response{Results: deduped, Total: len(deduped)}
}
```

**设计理由**：
- `ants.Pool` 限制并发 goroutine 数量
- 双层 context：全局兜底 + 每个引擎独立超时
- `sync.WaitGroup + channel` 收集结果，简洁可靠
- 基于 URL 去重，按引擎权重排序

### 4.3 数据模型

```go
package models

type Category string

const (
    CategoryGeneral Category = "general"
    CategoryImages  Category = "images"
    CategoryVideos  Category = "videos"
    CategoryNews    Category = "news"
)

type Request struct {
    Query      string
    Category   Category
    Language   string
    SafeSearch bool
    TimeRange  string
    Page       int
    PageSize   int
}

type Result struct {
    Title        string
    URL          string
    Content      string
    Engine       string
    Category     Category
    Score        float64
    ThumbnailURL string
    PublishedAt  *time.Time
}

type Response struct {
    Results     []Result
    Suggestions []string
    Corrections []string
    Total       int
}
```

---

## 5. RESTful API 设计

### 5.1 端点一览

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/search` | 搜索（核心） |
| `GET` | `/api/engines` | 可用引擎列表 |
| `GET` | `/api/categories` | 可用分类 |
| `GET` | `/api/config` | 服务端配置 |
| `GET` | `/health` | 健康检查 |
| `GET` | `/metrics` | Prometheus 指标 |

### 5.2 搜索端点

```
GET /api/search?q=go+programming&category=general&language=zh-CN&safesearch=1&page=1
```

**响应**：

```json
{
  "query": "go programming",
  "category": "general",
  "results": [
    {
      "title": "The Go Programming Language",
      "url": "https://go.dev/",
      "content": "Go is an open source programming language...",
      "engine": "google",
      "score": 0.95
    }
  ],
  "suggestions": ["go programming tutorial"],
  "total": 128,
  "page": 1,
  "page_size": 10,
  "engines_used": ["google", "bing", "duckduckgo"],
  "engines_failed": [],
  "response_time_ms": 420
}
```

### 5.3 错误响应

```json
{
  "error": {
    "code": "INVALID_CATEGORY",
    "message": "category 'unknown' is not supported",
    "details": { "available": ["general", "images", "videos", "news"] }
  }
}
```

HTTP 状态码：
- `200` 成功
- `400` 请求参数错误
- `429` 请求过于频繁
- `500` 服务端内部错误
- `503` 所有引擎不可用

---

## 6. 缓存设计

### 6.1 多级缓存架构

```
读：请求 → 本地缓存(ristretto) → Redis → 搜索引擎
      命中 ←────────────── 命中 ←────────────

写：搜索引擎结果 → 本地缓存 → Redis
```

### 6.2 接口与实现

```go
package cache

type Cache interface {
    Get(key string) (*models.Response, bool)
    Set(key string, value *models.Response, ttl time.Duration)
    Delete(key string)
}

type MultiLevel struct {
    local            *ristretto.Cache
    remote           redis.Client
    logger           *mlog.Logger
    defaultLocalTTL  time.Duration
    defaultRemoteTTL time.Duration
}
```

### 6.3 Cache Key 与 TTL

```go
func (r *models.Request) CacheKey() string {
    h := fnv.New64a()
    h.Write([]byte(r.Query))
    return fmt.Sprintf("search:%s:%s:%d:%s:%d:%x",
        r.Category, r.Language, boolToInt(r.SafeSearch),
        r.TimeRange, r.Page, h.Sum64())
}
```

TTL 策略：
- `general`：30 秒
- `images`：2 分钟
- `news`：15 秒

---

## 7. 配置系统

### 7.1 配置文件（兼容 SearXNG）

```yaml
server:
  port: 8080
  bind_address: "0.0.0.0"

search:
  safe_search: 1
  autocomplete: "google"
  default_lang: "zh-CN"
  default_category: "general"

engines:
  - name: google
    enabled: true
    weight: 1.0
    timeout: 10
    
  - name: bing
    enabled: true
    weight: 0.8
    timeout: 8

outgoing:
  request_timeout: 15
  useragent: "SearGo/1.0"

cache:
  enabled: true
  local_ttl: 30
  redis_ttl: 300
  redis_addr: "localhost:6379"
```

### 7.2 环境变量覆盖

| 配置路径 | 环境变量 |
|---------|---------|
| `server.secret_key` | `SEARGO_SERVER_SECRET_KEY` |
| `engines[*].api_key` | `SEARGO_ENGINE_{NAME}_API_KEY` |
| `cache.redis_addr` | `SEARGO_CACHE_REDIS_ADDR` |

---

## 8. 前端设计

### 8.1 技术栈

- React 18 + Vite + TypeScript
- Zustand（状态管理）
- Axios（API 客户端）
- React Router（路由）

### 8.2 目录结构

```
web/src/
├── components/           # 通用组件
│   ├── SearchBar.tsx
│   ├── ResultCard.tsx
│   ├── ResultList.tsx
│   ├── FilterPanel.tsx
│   └── EngineSelector.tsx
├── pages/
│   ├── SearchPage.tsx
│   └── SettingsPage.tsx
├── hooks/
│   ├── useSearch.ts
│   ├── useEngines.ts
│   └── useConfig.ts
├── services/
│   └── api.ts
├── stores/
│   └── searchStore.ts
├── types/
│   ├── search.ts
│   ├── engine.ts
│   └── config.ts
├── App.tsx
└── main.tsx
```

### 8.3 与后端集成

前端构建产物通过 Go 的 `embed` 嵌入二进制：

```go
//go:embed dist
var webFS embed.FS

func (s *Server) setupStatic() {
    dist, _ := fs.Sub(webFS, "dist")
    s.router.StaticFS("/", http.FS(dist))
}
```

---

## 9. 基础设施

### 9.1 错误处理

三层防护：
1. **Gin Recovery**：捕获 panic，返回 500
2. **全局错误中间件**：统一错误格式和日志
3. **业务层**：使用结构化错误类型

```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
    Status  int    `json:"-"`
}
```

### 9.2 日志（mlog）

直接使用 `github.com/odysseythink/mlog`：

```go
package logger

import "github.com/odysseythink/mlog"

var defaultLogger *mlog.Logger

func Init(level string, output string) error {
    defaultLogger = mlog.New(mlog.Config{
        Level:  level,
        Output: output,
        Format: "json",
    })
    return nil
}

func WithContext(ctx context.Context) *mlog.Logger {
    if reqID := ctx.Value("request_id"); reqID != nil {
        return defaultLogger.With("request_id", reqID)
    }
    return defaultLogger
}
```

### 9.3 测试策略

**三层测试金字塔**：
- **单元测试（75%）**：Go `testing` + `testify`，Mock 引擎测试调度器
- **集成测试（20%）**：`httptest` 测试 API 端点
- **E2E 测试（5%）**：Playwright 测试完整搜索流程

覆盖率阈值：Go ≥ 60%，前端 ≥ 50%。

### 9.4 部署

**两种部署方式都支持**：

**A. Docker（推荐生产环境）**：
- 多阶段 Dockerfile（Node 构建前端 → Go 构建后端 → Alpine 运行）
- docker-compose.yml 包含 Redis 依赖

**B. 二进制（推荐开发/边缘部署）**：
- `make build` 产出单个二进制文件
- 直接运行：`./seargo -config configs/settings.yml`
- 环境变量覆盖敏感配置

---

## 10. 迁移路线图

### Phase 1：基础设施（MVP）
- [ ] 项目骨架（Go module + 目录结构）
- [ ] mlog 日志集成
- [ ] YAML 配置系统
- [ ] 多级缓存（ristretto + Redis）
- [ ] HTTP 客户端封装（req）
- [ ] Gin HTTP 服务 + 路由
- [ ] 错误处理中间件
- [ ] 健康检查端点

### Phase 2：核心搜索能力
- [ ] Engine 接口定义
- [ ] 引擎注册中心
- [ ] 搜索调度器（goroutine 池 + 超时 + 合并）
- [ ] Google 引擎实现
- [ ] Bing 引擎实现
- [ ] DuckDuckGo 引擎实现
- [ ] `/api/search` 端点
- [ ] `/api/engines` 端点

### Phase 3：前端
- [ ] React + Vite + TypeScript 项目初始化
- [ ] 搜索页面（SearchBar + ResultList）
- [ ] API 客户端封装
- [ ] Zustand 状态管理
- [ ] 前端构建产物嵌入 Go 二进制
- [ ] 设置页面

### Phase 4：完善与扩展
- [ ] 更多搜索引擎（Brave、Yahoo、Wikipedia 等）
- [ ] 图片/视频/新闻分类支持
- [ ] 自动补全
- [ ] Prometheus 指标
- [ ] 单元测试覆盖率达到 60%
- [ ] Docker 镜像优化
- [ ] 文档完善

---

## 11. 附录

### 11.1 依赖库清单

| 库 | 用途 |
|---|------|
| `github.com/gin-gonic/gin` | HTTP Web 框架 |
| `github.com/odysseythink/mlog` | 日志 |
| `github.com/go-resty/resty/v2` 或 `github.com/imroc/req/v3` | HTTP 客户端 |
| `github.com/dgraph-io/ristretto` | 内存缓存 |
| `github.com/redis/go-redis/v9` | Redis 客户端 |
| `github.com/panjf2000/ants/v2` | Goroutine 池 |
| `gopkg.in/yaml.v3` | YAML 解析 |
| `github.com/stretchr/testify` | 测试断言 |

### 11.2 相关文档

- SearXNG 官方文档：https://docs.searxng.org
- SearXNG 配置参考：`../searxng-master/docs/admin/settings/index.html`
