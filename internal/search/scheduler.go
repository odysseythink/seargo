package search

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/metrics"
	"github.com/seargo/seargo/pkg/models"
)

type Scheduler struct {
	engines              map[string]engine.Engine
	engineConfigs        map[string]config.EngineConfig
	workerPool           *ants.Pool
	cache                cache.Cache
	globalTimeout        time.Duration
	defaultEngineTimeout time.Duration
	suspension           *SuspensionTracker
}

func NewScheduler(cfg *config.Config, c cache.Cache) (*Scheduler, error) {
	pool, err := ants.NewPool(50)
	if err != nil {
		return nil, err
	}

	engineCfgs := make(map[string]config.EngineConfig, len(cfg.Engines))
	for _, ec := range cfg.Engines {
		key := ec.Engine
		if key == "" {
			key = ec.Name
		}
		engineCfgs[key] = ec
	}

	return &Scheduler{
		engines:              make(map[string]engine.Engine),
		engineConfigs:        engineCfgs,
		workerPool:           pool,
		cache:                c,
		globalTimeout:        time.Duration(cfg.Outgoing.RequestTimeout) * time.Second,
		defaultEngineTimeout: 8 * time.Second,
		suspension:           NewSuspensionTracker(cfg.Search),
	}, nil
}

func (s *Scheduler) RegisterEngine(name string, e engine.Engine) {
	s.engines[name] = e
}

func (s *Scheduler) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	start := time.Now()

	// 1. Cache check
	if s.cache != nil {
		if cached, ok := s.cache.Get(req.CacheKey()); ok {
			cached.ResponseTimeMs = time.Since(start).Milliseconds()
			return cached, nil
		}
	}

	// 2. Select engines
	selected := s.selectEngines(req.Category)
	if len(selected) == 0 {
		return &models.Response{
			Query:   req.Query,
			Results: []models.Result{},
		}, nil
	}

	// 3. Global timeout
	ctx, cancel := context.WithTimeout(ctx, s.globalTimeout)
	defer cancel()

	// 4. Concurrent query
	results, enginesUsed, enginesFailed := s.queryEngines(ctx, req, selected)

	// 5. Post-process
	response := s.postProcess(results, req)
	response.EnginesUsed = enginesUsed
	response.EnginesFailed = enginesFailed
	response.ResponseTimeMs = time.Since(start).Milliseconds()

	metrics.SearchResultsTotal.WithLabelValues(string(req.Category)).Add(float64(len(response.Results)))

	// 6. Write cache
	if s.cache != nil {
		s.cache.Set(req.CacheKey(), response, s.cacheTTL(req.Category))
	}

	return response, nil
}

func (s *Scheduler) selectEngines(cat models.Category) []engine.Engine {
	var selected []engine.Engine
	for _, e := range s.engines {
		name := e.Name()
		cfg, ok := s.engineConfigs[name]
		if !ok || cfg.Disabled {
			continue
		}
		// Check suspension
		if s.suspension != nil && s.suspension.IsSuspended(name) {
			continue
		}
		for _, c := range e.Categories() {
			if c == cat {
				selected = append(selected, e)
				break
			}
		}
	}
	return selected
}

func (s *Scheduler) queryEngines(ctx context.Context, req *models.Request, engines []engine.Engine) ([]models.Result, []string, []string) {
	var wg sync.WaitGroup
	resultCh := make(chan []models.Result, len(engines))
	var usedMu, failedMu sync.Mutex
	enginesUsed := make([]string, 0, len(engines))
	enginesFailed := make([]string, 0, len(engines))

	for _, e := range engines {
		wg.Add(1)
		eng := e // capture loop variable
		s.workerPool.Submit(func() {
			defer wg.Done()

			engineStart := time.Now()
			timeout := s.getEngineTimeout(eng.Name())
			engineCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resp, err := eng.Search(engineCtx, req)
			if err != nil {
				metrics.EngineQueriesTotal.WithLabelValues(eng.Name(), "failed").Inc()
				logger.Warn("engine failed", "engine", eng.Name(), "error", err)
				failedMu.Lock()
				enginesFailed = append(enginesFailed, eng.Name())
				failedMu.Unlock()
				return
			}

			metrics.EngineQueriesTotal.WithLabelValues(eng.Name(), "success").Inc()
			metrics.EngineQueryDuration.WithLabelValues(eng.Name()).Observe(time.Since(engineStart).Seconds())

			usedMu.Lock()
			enginesUsed = append(enginesUsed, eng.Name())
			usedMu.Unlock()
			resultCh <- resp.Results
		})
	}

	go func() { wg.Wait(); close(resultCh) }()

	var allResults []models.Result
	for r := range resultCh {
		allResults = append(allResults, r...)
	}
	return allResults, enginesUsed, enginesFailed
}

func (s *Scheduler) getEngineTimeout(name string) time.Duration {
	if cfg, ok := s.engineConfigs[name]; ok && cfg.Timeout > 0 {
		return time.Duration(cfg.Timeout * float64(time.Second))
	}
	return s.defaultEngineTimeout
}

func (s *Scheduler) postProcess(results []models.Result, req *models.Request) *models.Response {
	deduped := deduplicate(results)
	sort.Slice(deduped, func(i, j int) bool {
		return s.score(deduped[i]) > s.score(deduped[j])
	})

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if len(deduped) > pageSize {
		deduped = deduped[:pageSize]
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  deduped,
		Total:    len(deduped),
		Page:     req.Page,
		PageSize: pageSize,
	}
}

func deduplicate(results []models.Result) []models.Result {
	seen := make(map[string]bool)
	var out []models.Result
	for _, r := range results {
		if seen[r.URL] {
			continue
		}
		seen[r.URL] = true
		out = append(out, r)
	}
	return out
}

func (s *Scheduler) score(r models.Result) float64 {
	cfg, ok := s.engineConfigs[r.Engine]
	if !ok {
		return r.Score
	}
	return r.Score * cfg.Weight
}

// SuspensionTracker tracks engine suspension state.
// Full implementation in Task 13.
type SuspensionTracker struct{}

func NewSuspensionTracker(cfg config.SearchConfig) *SuspensionTracker {
	return nil // stub
}

func (st *SuspensionTracker) IsSuspended(name string) bool {
	return false // stub
}

func (s *Scheduler) cacheTTL(cat models.Category) time.Duration {
	switch cat {
	case models.CategoryImages:
		return 2 * time.Minute
	case models.CategoryNews:
		return 15 * time.Second
	case models.CategoryVideos:
		return 2 * time.Minute
	default:
		return 30 * time.Second
	}
}
