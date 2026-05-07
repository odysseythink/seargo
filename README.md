# SearGo

A lightweight, self-hosted meta search engine written in Go, featuring a modern React frontend with a beautiful dark theme.

## Features

- **Multi-Engine Search** — Aggregates results from Google, Bing, DuckDuckGo, Brave, Wikipedia, and Yahoo
- **Dark Theme UI** — Modern card-based interface with glow effects, engine badges, and animations
- **Multi-Level Cache** — Local in-memory cache + Redis for optimal performance
- **Metrics & Monitoring** — Built-in Prometheus metrics endpoint
- **Graceful Shutdown** — Clean shutdown with timeout handling
- **Configurable** — YAML-based configuration for engines, cache, and server settings
- **Single Binary** — Frontend embedded into the Go binary via `embed.FS`

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────────┐
│   React     │────▶│   Go API    │────▶│ Search Engines  │
│  Frontend   │◄────│   Server    │◄────│ (Google/Bing/   │
│  (Dark UI)  │     │  (Gin/GORM) │     │  DDG/Brave/...) │
└─────────────┘     └──────┬──────┘     └─────────────────┘
                           │
                    ┌──────┴──────┐
                    │  Scheduler  │
                    │  + Cache    │
                    └─────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 20+ (for frontend development)
- Redis (optional, for distributed caching)

### Build & Run

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build Go binary
make build

# Run with default config
./bin/seargo -config configs/settings.yml
```

The server will start on `http://localhost:8080`.

### Configuration

Edit `configs/settings.yml` to customize:

```yaml
server:
  port: 8080

search:
  default_lang: "zh-CN"
  max_results: 10

engines:
  - name: google
    enabled: true
  - name: bing
    enabled: true
  # ... more engines

cache:
  enabled: true
  redis_addr: "localhost:6379"
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Search page (React app) |
| GET | `/api/search?q=...` | Search API |
| GET | `/api/autocomplete?q=...` | Autocomplete suggestions |
| GET | `/api/engines` | List available engines |
| GET | `/metrics` | Prometheus metrics |

## Development

### Frontend

```bash
cd web
npm run dev        # Development server
npm run build      # Production build
npm run lint       # ESLint check
```

### Backend

```bash
go run ./cmd/seargo -config configs/settings.yml
```

### Add a New Search Engine

Implement the `engine.Engine` interface in `engines/<name>/`:

```go
type Engine interface {
    Name() string
    Init(config map[string]any) error
    Search(ctx context.Context, query string, opts ...SearchOption) (*Result, error)
}
```

Register in `cmd/seargo/main.go`:

```go
_ "github.com/seargo/seargo/engines/<name>"
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Gin, Resty |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS |
| Cache | ristretto (local), Redis |
| Metrics | Prometheus client |
| Logging | mlog |

## License

MIT
