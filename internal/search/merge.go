package search

import (
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/seargo/seargo/pkg/models"
)

// trackingParams 是需要从 URL 中移除的已知追踪参数。
var trackingParams = map[string]bool{
	"utm_source":   true,
	"utm_medium":   true,
	"utm_campaign": true,
	"utm_term":     true,
	"utm_content":  true,
	"fbclid":       true,
	"gclid":        true,
	"ref":          true,
	"ref_src":      true,
	"ref_url":      true,
}

// normalizeURL 对 URL 做规范化：
// - scheme 和 host 转小写
// - 去除 www. 前缀（非 www2/www3 等子域名）
// - 去除 trailing slash
// - 去除已知追踪参数
// - 去除 fragment
func normalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// 去掉 www. 前缀（仅当 host 是 www.xxx 格式时）
	if strings.HasPrefix(u.Host, "www.") && !strings.HasPrefix(u.Host, "www2.") &&
		!strings.HasPrefix(u.Host, "www3.") {
		u.Host = u.Host[4:]
	}

	u.Path = strings.TrimSuffix(u.Path, "/")

	// 移除追踪参数
	q := u.Query()
	for param := range trackingParams {
		q.Del(param)
	}
	u.RawQuery = q.Encode()

	u.Fragment = ""

	return u.String()
}

// UnresponsiveEngine 记录一个无响应的引擎及其原因。
type UnresponsiveEngine struct {
	Name   string
	Reason string
}

// ResultContainer 管理跨引擎搜索结果的并发写入、去重合并、排序和分组。
type ResultContainer struct {
	mu            sync.Mutex
	closed        bool
	results       map[string]*models.Result // key = dedupKey
	answers       []models.Answer
	suggestions   []string
	suggestionSet map[string]bool
	corrections   []string
	infoboxes     []models.Infobox
	engineData    map[string]any
	unresponsive  []UnresponsiveEngine
	engineWeights map[string]float64
}

// NewResultContainer 创建结果容器。
func NewResultContainer(engineWeights map[string]float64) *ResultContainer {
	return &ResultContainer{
		results:       make(map[string]*models.Result),
		suggestionSet: make(map[string]bool),
		engineData:    make(map[string]any),
		engineWeights: engineWeights,
	}
}

// dedupKey 生成去重用的唯一键：template|host|path|query|thumbnail。
func dedupKey(r models.Result) string {
	normURL := normalizeURL(r.URL)
	u, err := url.Parse(normURL)
	if err != nil {
		return r.Template + "|" + r.URL + "|" + r.ThumbnailURL
	}
	return r.Template + "|" + u.Host + "|" + u.Path + "|" + u.RawQuery + "|" + r.ThumbnailURL
}

// preferHTTPS 判断两个 URL 之间是否应优选 HTTPS 版本。
func preferHTTPS(newURL, oldURL string) bool {
	return strings.HasPrefix(newURL, "https://") && !strings.HasPrefix(oldURL, "https://")
}

// Extend 将单个引擎的结果并入容器。positionBase 是该引擎结果起始位置（0-based）。
func (c *ResultContainer) Extend(engineName string, results []models.Result, positionBase int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}

	for i, r := range results {
		key := dedupKey(r)
		position := positionBase + i + 1 // 1-based

		existing, ok := c.results[key]
		if !ok {
			r.Engine = engineName
			r.Engines = []string{engineName}
			r.Positions = []int{position}
			if r.Domain == "" {
				r.Domain = extractDomain(r.URL)
			}
			c.results[key] = &r
			continue
		}

		// 合并已存在的结果
		existing.Engines = appendUniqueStr(existing.Engines, engineName)
		existing.Positions = append(existing.Positions, position)
		if len(r.Title) > len(existing.Title) {
			existing.Title = r.Title
		}
		if len(r.Content) > len(existing.Content) {
			existing.Content = r.Content
		}
		if preferHTTPS(r.URL, existing.URL) {
			existing.URL = r.URL
		}
		if r.ThumbnailURL != "" && existing.ThumbnailURL == "" {
			existing.ThumbnailURL = r.ThumbnailURL
		}
	}
}

// extractDomain 从 URL 提取域名。
func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

func appendUniqueStr(slice []string, item string) []string {
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}

// Close 标记容器写入完成，计算分数和排序。
func (c *ResultContainer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true

	for _, r := range c.results {
		c.calculateScore(r)
	}
}

// calculateScore 计算 SearXNG 风格分数：Σ (weight / position) × count。
func (c *ResultContainer) calculateScore(r *models.Result) {
	score := 0.0
	for i, pos := range r.Positions {
		engineName := ""
		if i < len(r.Engines) {
			engineName = r.Engines[i]
		}
		weight := c.engineWeights[engineName]
		if weight == 0 {
			weight = 1.0
		}
		score += weight / float64(pos)
	}
	score *= float64(len(r.Positions))
	r.Score = score
}

type resultGroup struct {
	key     string
	items   []*models.Result
	lastIdx int
}

// GetOrderedResults 返回按 score 降序 + category grouping 重排后的结果列表。
func (c *ResultContainer) GetOrderedResults() []models.Result {
	all := make([]*models.Result, 0, len(c.results))
	for _, r := range c.results {
		all = append(all, r)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		return all[i].URL < all[j].URL
	})

	const groupWindow = 20
	const groupMaxSize = 8

	output := make([]models.Result, 0, len(all))
	groupMap := make(map[string]*resultGroup)

	for _, r := range all {
		gKey := string(r.Category) + "|" + r.Template
		if r.ThumbnailURL != "" {
			gKey += "|thumb"
		}

		grp, exists := groupMap[gKey]
		if exists {
			distance := len(output) - grp.lastIdx
			if len(grp.items) < groupMaxSize && distance < groupWindow {
				grp.items = append(grp.items, r)
				grp.lastIdx = len(output)
				output = insertAt(output, grp.lastIdx, *r)
				for _, g := range groupMap {
					if g != grp && g.lastIdx >= grp.lastIdx {
						g.lastIdx++
					}
				}
				continue
			}
		}

		newGrp := &resultGroup{key: gKey, items: []*models.Result{r}, lastIdx: len(output)}
		groupMap[gKey] = newGrp
		output = append(output, *r)
	}

	return output
}

func insertAt(slice []models.Result, idx int, item models.Result) []models.Result {
	if idx >= len(slice) {
		return append(slice, item)
	}
	slice = append(slice, models.Result{})
	copy(slice[idx+1:], slice[idx:])
	slice[idx] = item
	return slice
}

// AddSuggestions 添加引擎的建议列表（大小写去重）。
func (c *ResultContainer) AddSuggestions(engineName string, suggestions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for _, s := range suggestions {
		lower := strings.ToLower(s)
		if !c.suggestionSet[lower] {
			c.suggestionSet[lower] = true
			c.suggestions = append(c.suggestions, s)
		}
	}
}

// AddAnswers 添加答案列表。
func (c *ResultContainer) AddAnswers(engineName string, answers []models.Answer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.answers = append(c.answers, answers...)
}

// AddCorrections 添加拼写纠正建议。
func (c *ResultContainer) AddCorrections(engineName string, corrections []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.corrections = append(c.corrections, corrections...)
}

// AddInfoboxes 添加信息框。
func (c *ResultContainer) AddInfoboxes(engineName string, infoboxes []models.Infobox) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.infoboxes = append(c.infoboxes, infoboxes...)
}

// AddEngineData 添加引擎的透传数据。
func (c *ResultContainer) AddEngineData(engineName string, data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for k, v := range data {
		c.engineData[engineName+"."+k] = v
	}
}

// MarkUnresponsive 记录一个无响应的引擎。
func (c *ResultContainer) MarkUnresponsive(engineName, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unresponsive = append(c.unresponsive, UnresponsiveEngine{Name: engineName, Reason: reason})
}

// GetSuggestions 返回建议列表。
func (c *ResultContainer) GetSuggestions() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.suggestions) == 0 {
		return nil
	}
	result := make([]string, len(c.suggestions))
	copy(result, c.suggestions)
	return result
}

// GetAnswers 返回答案列表。
func (c *ResultContainer) GetAnswers() []models.Answer {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.answers) == 0 {
		return nil
	}
	result := make([]models.Answer, len(c.answers))
	copy(result, c.answers)
	return result
}

// GetCorrections 返回纠正列表。
func (c *ResultContainer) GetCorrections() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.corrections) == 0 {
		return nil
	}
	result := make([]string, len(c.corrections))
	copy(result, c.corrections)
	return result
}

// GetInfoboxes 返回信息框列表。
func (c *ResultContainer) GetInfoboxes() []models.Infobox {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.infoboxes) == 0 {
		return nil
	}
	result := make([]models.Infobox, len(c.infoboxes))
	copy(result, c.infoboxes)
	return result
}

// GetEngineData 返回引擎透传数据。
func (c *ResultContainer) GetEngineData() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.engineData) == 0 {
		return nil
	}
	result := make(map[string]any, len(c.engineData))
	for k, v := range c.engineData {
		result[k] = v
	}
	return result
}

// GetUnresponsive 返回无响应引擎列表。
func (c *ResultContainer) GetUnresponsive() []UnresponsiveEngine {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.unresponsive) == 0 {
		return nil
	}
	result := make([]UnresponsiveEngine, len(c.unresponsive))
	copy(result, c.unresponsive)
	return result
}

// GetEnginesUsed 返回有结果贡献的引擎名。
func (c *ResultContainer) GetEnginesUsed() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	seen := make(map[string]bool)
	var names []string
	for _, r := range c.results {
		for _, e := range r.Engines {
			if !seen[e] {
				seen[e] = true
				names = append(names, e)
			}
		}
	}
	return names
}

// GetEnginesFailed 返回失败引擎名列表。
func (c *ResultContainer) GetEnginesFailed() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var names []string
	for _, ue := range c.unresponsive {
		names = append(names, ue.Name)
	}
	return names
}
