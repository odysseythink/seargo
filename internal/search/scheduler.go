package search

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/internal/cache"
	"github.com/seargo/seargo/internal/config"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/internal/logger"
	"github.com/seargo/seargo/internal/metrics"
	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/internal/search/processor"
	"github.com/seargo/seargo/internal/search/query"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

type Scheduler struct {
	processors           map[string]processor.Processor
	engineConfigs        map[string]config.EngineConfig
	engineWeights        map[string]float64
	engineShortcuts      map[string]string
	engineNames          []string
	allCategories        []models.Category
	workerPool           *ants.Pool
	cache                cache.Cache
	globalTimeout        time.Duration
	defaultEngineTimeout time.Duration
	suspension           *SuspensionTracker
	categoriesAsTabs     map[string]config.CategoryTabConfig
	pluginStorage        *plugin.PluginStorage
	answererStorage      *answerer.AnswererStorage
}

// isEngineEnabled 判断引擎是否启用。Enabled 优先于 Disabled。
func isEngineEnabled(ec config.EngineConfig) bool {
	if ec.Enabled {
		return true
	}
	return !ec.Disabled
}

// engineKey 返回引擎在 map 中的 key。
func engineKey(ec config.EngineConfig) string {
	if ec.Engine != "" {
		return ec.Engine
	}
	return ec.Name
}

func NewScheduler(cfg *config.Config, c cache.Cache, client *httpx.Client, pluginStorage *plugin.PluginStorage, answererStorage *answerer.AnswererStorage) (*Scheduler, error) {
	pool, err := ants.NewPool(50)
	if err != nil {
		return nil, err
	}

	// Build engine configs, weights, shortcuts, names
	engineCfgs := make(map[string]config.EngineConfig, len(cfg.Engines)*2)
	engineWeights := make(map[string]float64)
	engineShortcuts := make(map[string]string)
	engineNames := make([]string, 0, len(cfg.Engines))

	for _, ec := range cfg.Engines {
		key := engineKey(ec)
		engineCfgs[key] = ec
		if ec.Name != "" && ec.Name != key {
			engineCfgs[ec.Name] = ec
		}
		engineWeights[key] = ec.Weight
		if ec.Shortcut != "" {
			engineShortcuts[ec.Shortcut] = key
		}
		engineNames = append(engineNames, key)
	}

	// Build categories list for bang parser
	var allCategories []models.Category
	for catStr := range cfg.CategoriesAsTabs {
		allCategories = append(allCategories, models.Category(catStr))
	}

	// Build processors
	processors := make(map[string]processor.Processor)
	suspension := NewSuspensionTracker(cfg.Search)

	for _, ec := range cfg.Engines {
		if !isEngineEnabled(ec) {
			continue
		}
		lookupName := engineKey(ec)
		eng, ok := engine.Get(lookupName)
		if !ok {
			logger.Warn("Engine not found", "engine", lookupName)
			continue
		}
		proc, err := processor.NewProcessorFromConfig(eng, ec, suspension, client)
		if err != nil {
			logger.Error("Failed to create processor", "engine", lookupName, "error", err)
			continue
		}
		processors[lookupName] = proc
		logger.Info("Engine registered", "engine", lookupName)
	}

	// Compute global timeout
	globalTimeout := time.Duration(cfg.Outgoing.RequestTimeout) * time.Second

	metrics.EngineReloadsTotal.WithLabelValues("all", "success").Inc()

	return &Scheduler{
		processors:           processors,
		engineConfigs:        engineCfgs,
		engineWeights:        engineWeights,
		engineShortcuts:      engineShortcuts,
		engineNames:          engineNames,
		allCategories:        allCategories,
		workerPool:           pool,
		cache:                c,
		globalTimeout:        globalTimeout,
		defaultEngineTimeout: 8 * time.Second,
		suspension:           suspension,
		categoriesAsTabs:     cfg.CategoriesAsTabs,
		pluginStorage:        pluginStorage,
		answererStorage:      answererStorage,
	}, nil
}

func toModelCategories(cats []string) []models.Category {
	result := make([]models.Category, len(cats))
	for i, c := range cats {
		result[i] = models.Category(c)
	}
	return result
}

// Search 执行完整的搜索流程。
func (s *Scheduler) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	start := time.Now()

	// 1. Parse query
	rtq := query.NewRawTextQuery(req.Query)
	parsed, err := rtq.Parse(s.engineShortcuts, s.engineNames, s.allCategories, nil)
	if err != nil {
		return nil, err
	}

	// 2. Cache check
	if s.cache != nil {
		key := s.cacheKey(parsed, req)
		if cached, ok := s.cache.Get(key); ok {
			cached.ResponseTimeMs = time.Since(start).Milliseconds()
			return cached, nil
		}
	}

	// 3. External bang redirect
	if parsed.ExternalBang != "" {
		if redirectURL, ok := externalBangURL(parsed.ExternalBang, parsed.Terms); ok {
			resp := &models.Response{
				RedirectURL: redirectURL,
			}
			if s.cache != nil {
				s.cache.Set(s.cacheKey(parsed, req), resp, s.cacheTTL(req.Category))
			}
			return resp, nil
		}
	}

	// Build search context (used by hooks and processors)
	searchCtx := s.buildSearchContext(parsed, req)

	// Hook: pre_search â plugins can abort the search before engine execution.
	if s.pluginStorage != nil {
		if !s.pluginStorage.PreSearch(searchCtx) {
			resp := &models.Response{
				Query:   req.Query,
				Results: []models.Result{},
			}
			if s.cache != nil {
				s.cache.Set(s.cacheKey(parsed, req), resp, s.cacheTTL(req.Category))
			}
			return resp, nil
		}
	}

	// Hook: answerer_ask â instant answers from answerer keywords.
	var answererResults []models.Result
	if s.answererStorage != nil {
		actx := &answerer.AnswerContext{
			Query: req.Query,
		}
		answererResults = s.answererStorage.Ask(actx)
	}

	// 4. Select processors
	procs := s.selectProcessors(parsed, req.Category)
	if len(procs) == 0 {
		return &models.Response{
			Query:   req.Query,
			Results: []models.Result{},
		}, nil
	}

	// 5. Compute timeout
	timeout := s.computeTimeout(parsed, procs)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 6. Execute processors (concurrent)
	container := NewTypedResultContainer(s.engineWeights)
	s.executeProcessors(ctx, procs, parsed, req.Page, container, searchCtx)

	// Hook: post_search â plugins append additional results.
	if s.pluginStorage != nil && searchCtx != nil {
		pluginResults := s.pluginStorage.PostSearch(searchCtx)
		container.AddPluginResults(answererResults)
		container.AddPluginResults(pluginResults)
	} else if len(answererResults) > 0 {
		container.AddPluginResults(answererResults)
	}

	container.Close()

	results := container.Results()
	suggestions := container.GetSuggestions()
	answers := container.GetAnswers()
	corrections := container.GetCorrections()
	infoboxes := container.GetInfoboxes()
	engineData := container.GetEngineData()
	unresponsive := container.GetUnresponsive()

	// 7. All engines failed
	if len(results) == 0 && len(unresponsive) > 0 && len(unresponsive) == len(procs) {
		return nil, fmt.Errorf("all engines failed")
	}

	// 8. Paginate
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	window, total := paginate(results, req.Page, pageSize)

	// 9. Build response
	response := &models.Response{
		Query:          req.Query,
		Category:       req.Category,
		Results:        window,
		Suggestions:    suggestions,
		Answers:        answers,
		Corrections:    corrections,
		Infoboxes:      infoboxes,
		EngineData:     engineData,
		Total:          total,
		Page:           req.Page,
		PageSize:       pageSize,
		EnginesUsed:    container.GetEnginesUsed(),
		EnginesFailed:  container.GetEnginesFailed(),
		ResponseTimeMs: time.Since(start).Milliseconds(),
	}

	// 10. Record metrics
	s.recordMetrics(response)
	metrics.SearchResultsTotal.WithLabelValues(string(req.Category)).Add(float64(len(response.Results)))

	// 11. Write cache
	if s.cache != nil {
		s.cache.Set(s.cacheKey(parsed, req), response, s.cacheTTL(req.Category))
	}

	return response, nil
}

// executeProcessors 并发执行所有 processor，将结果写入 container。
func (s *Scheduler) executeProcessors(ctx context.Context, procs []processor.Processor, parsed *query.ParsedQuery, page int, container *TypedResultContainer, searchCtx *plugin.SearchContext) {
	var wg sync.WaitGroup

	for _, p := range procs {
		wg.Add(1)
		proc := p
		s.workerPool.Submit(func() {
			defer wg.Done()

			engineStart := time.Now()
			result, err := proc.Search(ctx, parsed, page)

			if err != nil {
				metrics.EngineQueriesTotal.WithLabelValues(proc.Engine().Name(), "failed").Inc()
				errorClass := classifyError(err)
				metrics.EngineFailuresTotal.WithLabelValues(proc.Engine().Name(), errorClass).Inc()
				// Track parser failures (e.g. JSON/HTML parsing errors from engines)
				if strings.Contains(strings.ToLower(err.Error()), "parse") {
					metrics.EngineParserFailures.WithLabelValues(proc.Engine().Name()).Inc()
				}
				logger.Warn("engine failed", "engine", proc.Engine().Name(), "error", err)
				container.MarkUnresponsive(proc.Engine().Name(), err.Error())
				return
			}

			metrics.EngineQueriesTotal.WithLabelValues(proc.Engine().Name(), "success").Inc()
			metrics.EngineQueryDuration.WithLabelValues(proc.Engine().Name()).Observe(time.Since(engineStart).Seconds())
			metrics.EngineResults.WithLabelValues(proc.Engine().Name()).Observe(float64(len(result.Results)))

			if len(result.TypedResults) > 0 {
				apiResults := results.ToAPIResult(result.TypedResults)
				container.Extend(proc.Engine().Name(), apiResults, 0)
			} else if len(result.Results) > 0 {
				container.Extend(proc.Engine().Name(), result.Results, 0)
			}
			// Hook: on_result â filter results through plugins.
			if s.pluginStorage != nil && searchCtx != nil {
				filtered := result.Results[:0]
				for i := range result.Results {
					if s.pluginStorage.OnResult(searchCtx, &result.Results[i]) {
						filtered = append(filtered, result.Results[i])
					}
				}
				result.Results = filtered
			}
			if len(result.Suggestions) > 0 {
				container.AddSuggestions(proc.Engine().Name(), result.Suggestions)
			}
			if len(result.Answers) > 0 {
				container.AddAnswers(proc.Engine().Name(), result.Answers)
			}
			if len(result.Corrections) > 0 {
				container.AddCorrections(proc.Engine().Name(), result.Corrections)
			}
			if len(result.Infoboxes) > 0 {
				container.AddInfoboxes(proc.Engine().Name(), result.Infoboxes)
			}
			if len(result.EngineData) > 0 {
				container.AddEngineData(proc.Engine().Name(), result.EngineData)
			}
		})
	}

	wg.Wait()
}

// selectProcessors 根据 parsed query 选择 processor。
// 如果指定了引擎引用，只使用明确命名的引擎；否则按分类匹配。
func (s *Scheduler) selectProcessors(parsed *query.ParsedQuery, defaultCat models.Category) []processor.Processor {
	// If explicit engine refs, only use those
	if len(parsed.EngineRefs) > 0 {
		var selected []processor.Processor
		for _, ref := range parsed.EngineRefs {
			if proc, ok := s.processors[ref]; ok && !proc.Suspended() {
				selected = append(selected, proc)
			}
		}
		return selected
	}

	// Otherwise match by category from categoriesAsTabs
	cat := defaultCat
	if len(parsed.Categories) > 0 {
		cat = parsed.Categories[0]
	}

	cfg, ok := s.categoriesAsTabs[string(cat)]
	if !ok {
		return nil
	}
	var selected []processor.Processor
	for _, engineName := range cfg.Engines {
		if proc, ok := s.processors[engineName]; ok && !proc.Suspended() {
			selected = append(selected, proc)
		}
	}
	return selected
}

// computeTimeout 计算搜索超时时间：取引擎超时、用户指定超时、max_request_timeout 和全局超时的最小值。
func (s *Scheduler) computeTimeout(parsed *query.ParsedQuery, procs []processor.Processor) time.Duration {
	timeout := s.defaultEngineTimeout

	// User-specified timeout from query
	if parsed.Timeout > 0 {
		timeout = time.Duration(parsed.Timeout * float64(time.Second))
	}

	// Per-engine timeout (take the shortest)
	for _, p := range procs {
		name := p.Engine().Name()
		if cfg, ok := s.engineConfigs[name]; ok && cfg.Timeout > 0 {
			engineTimeout := time.Duration(cfg.Timeout * float64(time.Second))
			if engineTimeout < timeout {
				timeout = engineTimeout
			}
		}
	}

	// Global timeout cap
	if s.globalTimeout > 0 && timeout > s.globalTimeout {
		timeout = s.globalTimeout
	}

	return timeout
}

// cacheKey 生成缓存键，基于 terms + engine_refs + categories + category + safesearch + timerange + page + pagesize 的 FNV 哈希。
func (s *Scheduler) cacheKey(parsed *query.ParsedQuery, req *models.Request) string {
	h := fnv.New64a()
	for _, t := range parsed.Terms {
		h.Write([]byte(t))
	}
	for _, ref := range parsed.EngineRefs {
		h.Write([]byte(ref))
	}
	for _, cat := range parsed.Categories {
		h.Write([]byte(cat))
	}
	h.Write([]byte(req.Category))
	h.Write([]byte(strconv.Itoa(req.SafeSearch)))
	h.Write([]byte(req.TimeRange))
	h.Write([]byte(strconv.Itoa(req.Page)))
	h.Write([]byte(strconv.Itoa(req.PageSize)))
	return fmt.Sprintf("search:%x", h.Sum64())
}

// externalBangURL 返回外部搜索引擎跳转 URL。
// 内置映射：g→google, ddg→duckduckgo, bing, gh→github, so→stackoverflow, wiki, yt。
func externalBangURL(bang string, terms []string) (string, bool) {
	q := url.QueryEscape(strings.Join(terms, " "))

	mappings := map[string]string{
		"g":    "https://www.google.com/search?q=%s",
		"ddg":  "https://duckduckgo.com/?q=%s",
		"bing": "https://www.bing.com/search?q=%s",
		"gh":   "https://github.com/search?q=%s",
		"so":   "https://stackoverflow.com/search?q=%s",
		"wiki": "https://en.wikipedia.org/w/index.php?search=%s",
		"yt":   "https://www.youtube.com/results?search_query=%s",
	}

	template, ok := mappings[bang]
	if !ok {
		return "", false
	}
	return fmt.Sprintf(template, q), true
}

// recordMetrics 记录结果流指标。
func (s *Scheduler) recordMetrics(resp *models.Response) {
	metrics.ResultStreamTotal.WithLabelValues("results").Add(float64(len(resp.Results)))
	metrics.ResultStreamTotal.WithLabelValues("suggestions").Add(float64(len(resp.Suggestions)))
	if resp.Answers != nil {
		metrics.ResultStreamTotal.WithLabelValues("answers").Add(float64(len(resp.Answers)))
	}
	if resp.Corrections != nil {
		metrics.ResultStreamTotal.WithLabelValues("corrections").Add(float64(len(resp.Corrections)))
	}
	if resp.Infoboxes != nil {
		metrics.ResultStreamTotal.WithLabelValues("infoboxes").Add(float64(len(resp.Infoboxes)))
	}
}

// paginate returns a stable windowed slice and the total count before windowing.
// page is 1-based; page=0 defaults to 1. pageSize <= 0 defaults to 10.
func paginate(results []models.Result, page, pageSize int) ([]models.Result, int) {
	total := len(results)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []models.Result{}, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return results[start:end], total
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

// buildSearchContext creates a SearchContext from parsed query and request.
func (s *Scheduler) buildSearchContext(parsed *query.ParsedQuery, req *models.Request) *plugin.SearchContext {
	queryStr := strings.Join(parsed.Terms, " ")
	if queryStr == "" {
		queryStr = req.Query
	}
	return &plugin.SearchContext{
		Query:      queryStr,
		RawQuery:   req.Query,
		Lang:       parsed.Lang,
		SafeSearch: req.SafeSearch,
		PageNo:     req.Page,
		TimeRange:  req.TimeRange,
	}
}
