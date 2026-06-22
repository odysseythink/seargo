package search

import (
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

// TypedResultContainer manages typed results with per-kind bucketing, dedup, and merging.
type TypedResultContainer struct {
	mu            sync.Mutex
	closed        bool
	buckets       map[string]map[string]*models.Result // kind -> dedupKey -> result
	answers       map[string]*models.Answer
	suggestions   map[string]string
	corrections   map[string]string
	infoboxes     map[string]*models.Infobox
	engineData    map[string]any
	engineWeights map[string]float64
	unresponsive  []UnresponsiveEngine
}

// NewTypedResultContainer creates a new typed result container.
func NewTypedResultContainer(engineWeights map[string]float64) *TypedResultContainer {
	return &TypedResultContainer{
		buckets:       make(map[string]map[string]*models.Result),
		answers:       make(map[string]*models.Answer),
		suggestions:   make(map[string]string),
		corrections:   make(map[string]string),
		infoboxes:     make(map[string]*models.Infobox),
		engineData:    make(map[string]any),
		engineWeights: engineWeights,
	}
}

// typedDedupKey — normalize URL then key on host|path|query|thumbnail.
func typedDedupKey(r models.Result) string {
	normURL := normalizeURL(r.URL)
	u, err := url.Parse(normURL)
	if err != nil {
		return r.URL + "|" + r.ThumbnailURL
	}
	return u.Host + "|" + u.Path + "|" + u.RawQuery + "|" + r.ThumbnailURL
}

// Extend adds results from an engine into the container.
func (c *TypedResultContainer) Extend(engineName string, apiResults []models.Result, positionBase int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}

	for i, r := range apiResults {
		pos := positionBase + i + 1 // 1-based
		kind := r.Kind
		if kind == "" {
			kind = "main"
		}

		switch kind {
		case "answer":
			c.mergeAnswer(engineName, r, pos)
		case "infobox":
			c.mergeInfobox(engineName, r, pos)
		case "suggestion":
			c.mergeSuggestion(r)
		case "correction":
			c.mergeCorrection(r)
		default:
			c.mergeResult(engineName, kind, r, pos)
		}
	}
}

func (c *TypedResultContainer) mergeResult(engineName, kind string, r models.Result, pos int) {
	bucket, ok := c.buckets[kind]
	if !ok {
		bucket = make(map[string]*models.Result)
		c.buckets[kind] = bucket
	}

	key := kind + "|" + typedDedupKey(r)
	existing, ok := bucket[key]
	if !ok {
		r.Engine = engineName
		r.Engines = []string{engineName}
		r.Positions = []int{pos}
		if r.Domain == "" {
			r.Domain = extractDomain(r.URL)
		}
		bucket[key] = &r
		return
	}

	// Merge
	existing.Engines = appendUniqueStr(existing.Engines, engineName)
	existing.Positions = append(existing.Positions, pos)
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
	// Merge extra fields
	if r.Extra != nil {
		if existing.Extra == nil {
			existing.Extra = make(map[string]any)
		}
		for k, v := range r.Extra {
			if _, set := existing.Extra[k]; !set {
				existing.Extra[k] = v
			}
		}
	}
}

func (c *TypedResultContainer) mergeAnswer(engineName string, r models.Result, pos int) {
	answerText := ""
	if r.Extra != nil {
		if t, ok := r.Extra["answer"]; ok {
			answerText, _ = t.(string)
		}
	}
	if answerText == "" {
		answerText = r.Content
	}
	if answerText == "" {
		return
	}
	key := strings.ToLower(answerText)
	if existing, ok := c.answers[key]; ok {
		existing.Engine = engineName
		return
	}
	c.answers[key] = &models.Answer{
		Answer:  answerText,
		URL:     r.URL,
		Content: r.Content,
		Engine:  engineName,
	}
}

func (c *TypedResultContainer) mergeInfobox(engineName string, r models.Result, pos int) {
	id := infoboxIDFromResult(r)
	if id == "" {
		id = "infobox:" + r.Title
	}
	key := normalizeInfoboxID(id)

	if existing, ok := c.infoboxes[key]; ok {
		incoming := buildInfoboxFromResult(engineName, r)
		mergeTwoInfoboxes(existing, incoming)
		return
	}

	c.infoboxes[key] = buildInfoboxFromResult(engineName, r)
}

func infoboxIDFromResult(r models.Result) string {
	if r.Extra != nil {
		if v, ok := r.Extra["infobox_id"]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	if r.URL != "" {
		return r.URL
	}
	return ""
}

func buildInfoboxFromResult(engineName string, r models.Result) *models.Infobox {
	id := infoboxIDFromResult(r)
	if id == "" {
		id = "infobox:" + r.Title
	}
	infobox := &models.Infobox{
		InfoboxID: id,
		Title:     r.Title,
		URL:       r.URL,
		Content:   r.Content,
		Engine:    engineName,
		Engines:   []string{engineName},
	}
	if r.Extra != nil {
		if v, ok := r.Extra["img_src"]; ok {
			infobox.ImgSrc, _ = v.(string)
		}
		if v, ok := r.Extra["attributes"]; ok {
			if attrs, ok := v.([]results.InfoboxAttribute); ok {
				for _, a := range attrs {
					infobox.Attributes = append(infobox.Attributes, models.InfoboxAttribute{
						Value: a.Value,
						Label: a.Label,
						URL:   a.URL,
					})
				}
			}
		}
		if v, ok := r.Extra["urls"]; ok {
			if urls, ok := v.([]results.InfoboxURL); ok {
				for _, u := range urls {
					infobox.URLs = append(infobox.URLs, models.InfoboxURL{
						Title: u.Title,
						URL:   u.URL,
					})
				}
			}
		}
		if v, ok := r.Extra["related_topics"]; ok {
			if topics, ok := v.([]string); ok {
				infobox.RelatedTopics = topics
			}
		}
	}
	return infobox
}

func normalizeInfoboxID(id string) string {
	id = strings.ReplaceAll(id, "http://", "https://")
	id = strings.ReplaceAll(id, "https://www.wikidata.org/entity/", "")

	// Strip any language-specific Wikipedia article prefix.
	if strings.HasPrefix(id, "https://") {
		if u, err := url.Parse(id); err == nil {
			host := strings.ToLower(u.Host)
			if strings.HasSuffix(host, ".wikipedia.org") {
				id = strings.TrimPrefix(u.Path, "/wiki/")
			}
		}
	}

	return strings.Trim(id, "_/")
}

func mergeTwoInfoboxes(a, b *models.Infobox) {
	if a.Title == "" && b.Title != "" {
		a.Title = b.Title
	}
	if a.Content == "" && b.Content != "" {
		a.Content = b.Content
	}
	if a.URL == "" && b.URL != "" {
		a.URL = b.URL
	}
	if a.ImgSrc == "" && b.ImgSrc != "" {
		a.ImgSrc = b.ImgSrc
	}

	seen := make(map[string]bool)
	for _, attr := range a.Attributes {
		seen[attr.Label] = true
	}
	for _, attr := range b.Attributes {
		if !seen[attr.Label] {
			a.Attributes = append(a.Attributes, attr)
			seen[attr.Label] = true
		}
	}

	seenURL := make(map[string]bool)
	for _, u := range a.URLs {
		seenURL[u.URL] = true
	}
	for _, u := range b.URLs {
		if !seenURL[u.URL] {
			a.URLs = append(a.URLs, u)
			seenURL[u.URL] = true
		}
	}

	a.Engines = appendUniqueStrSlice(a.Engines, b.Engines...)
}

func appendUniqueStrSlice(slice []string, items ...string) []string {
	for _, item := range items {
		slice = appendUniqueStr(slice, item)
	}
	return slice
}

func (c *TypedResultContainer) mergeSuggestion(r models.Result) {
	key := strings.ToLower(r.Title)
	if _, ok := c.suggestions[key]; !ok {
		c.suggestions[key] = r.Title
	}
}

func (c *TypedResultContainer) mergeCorrection(r models.Result) {
	key := strings.ToLower(r.Title)
	if _, ok := c.corrections[key]; !ok {
		c.corrections[key] = r.Title
	}
}

// AddSuggestions adds engine suggestions (case-insensitive dedup).
func (c *TypedResultContainer) AddSuggestions(engineName string, suggestions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for _, s := range suggestions {
		lower := strings.ToLower(s)
		if _, ok := c.suggestions[lower]; !ok {
			c.suggestions[lower] = s
		}
	}
}

// AddAnswers adds answer items.
func (c *TypedResultContainer) AddAnswers(engineName string, answers []models.Answer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for i := range answers {
		a := answers[i]
		key := strings.ToLower(a.Answer)
		if _, ok := c.answers[key]; !ok {
			c.answers[key] = &a
		}
	}
}

// AddCorrections adds spelling corrections.
func (c *TypedResultContainer) AddCorrections(engineName string, corrections []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for _, s := range corrections {
		lower := strings.ToLower(s)
		if _, ok := c.corrections[lower]; !ok {
			c.corrections[lower] = s
		}
	}
}

// AddInfoboxes adds infobox items.
func (c *TypedResultContainer) AddInfoboxes(engineName string, infoboxes []models.Infobox) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for i := range infoboxes {
		ib := infoboxes[i]
		id := ib.InfoboxID
		if id == "" {
			id = ib.URL
		}
		if id == "" {
			id = "infobox:" + ib.Title
		}
		key := normalizeInfoboxID(id)
		if existing, ok := c.infoboxes[key]; ok {
			ib.Engines = appendUniqueStr(ib.Engines, engineName)
			mergeTwoInfoboxes(existing, &ib)
			continue
		}
		ib.Engines = appendUniqueStr(ib.Engines, engineName)
		c.infoboxes[key] = &ib
	}
}

// AddEngineData adds engine passthrough data.
func (c *TypedResultContainer) AddEngineData(engineName string, data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for k, v := range data {
		c.engineData[engineName+"."+k] = v
	}
}

// MarkUnresponsive records an unresponsive engine.
func (c *TypedResultContainer) MarkUnresponsive(engineName, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unresponsive = append(c.unresponsive, UnresponsiveEngine{Name: engineName, Reason: reason})
}

// GetSuggestions returns the deduplicated suggestions list.
func (c *TypedResultContainer) GetSuggestions() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.suggestions) == 0 {
		return nil
	}
	result := make([]string, 0, len(c.suggestions))
	for _, v := range c.suggestions {
		result = append(result, v)
	}
	return result
}

// GetAnswers returns the answer list.
func (c *TypedResultContainer) GetAnswers() []models.Answer {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.answers) == 0 {
		return nil
	}
	result := make([]models.Answer, 0, len(c.answers))
	for _, v := range c.answers {
		result = append(result, *v)
	}
	return result
}

// GetCorrections returns the deduplicated corrections list.
func (c *TypedResultContainer) GetCorrections() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.corrections) == 0 {
		return nil
	}
	result := make([]string, 0, len(c.corrections))
	for _, v := range c.corrections {
		result = append(result, v)
	}
	return result
}

// GetInfoboxes returns the infobox list.
func (c *TypedResultContainer) GetInfoboxes() []models.Infobox {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.infoboxes) == 0 {
		return nil
	}
	result := make([]models.Infobox, 0, len(c.infoboxes))
	for _, v := range c.infoboxes {
		result = append(result, *v)
	}
	return result
}

// GetEngineData returns engine passthrough data.
func (c *TypedResultContainer) GetEngineData() map[string]any {
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

// GetUnresponsive returns the list of unresponsive engines.
func (c *TypedResultContainer) GetUnresponsive() []UnresponsiveEngine {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.unresponsive) == 0 {
		return nil
	}
	result := make([]UnresponsiveEngine, len(c.unresponsive))
	copy(result, c.unresponsive)
	return result
}

// GetEnginesUsed returns engine names that contributed results.
func (c *TypedResultContainer) GetEnginesUsed() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	seen := make(map[string]bool)
	var names []string
	for _, bucket := range c.buckets {
		for _, r := range bucket {
			for _, e := range r.Engines {
				if !seen[e] {
					seen[e] = true
					names = append(names, e)
				}
			}
		}
	}
	return names
}

// GetEnginesFailed returns failed engine names.
func (c *TypedResultContainer) GetEnginesFailed() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var names []string
	for _, ue := range c.unresponsive {
		names = append(names, ue.Name)
	}
	return names
}

// Close marks the container as closed and calculates scores for all results.
func (c *TypedResultContainer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true

	for _, bucket := range c.buckets {
		for _, r := range bucket {
			c.calculateScore(r)
		}
	}
}

// Results flattens all buckets, sorts by score, and applies category grouping.
func (c *TypedResultContainer) Results() []models.Result {
	c.mu.Lock()
	defer c.mu.Unlock()

	var all []*models.Result
	for _, bucket := range c.buckets {
		for _, r := range bucket {
			all = append(all, r)
		}
	}

	// Sort by Score desc, then URL asc
	sort.Slice(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		return all[i].URL < all[j].URL
	})

	// Apply category grouping
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

// calculateScore calculates the score using the SearXNG-style formula:
// Σ (weight / position) × count.
func (c *TypedResultContainer) calculateScore(r *models.Result) {
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

// AddPluginResults adds results from plugin post_search hooks into the container.
// Unlike Extend, these don't have an engine origin.
func (c *TypedResultContainer) AddPluginResults(results []models.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for i := range results {
		r := results[i]
		kind := r.Kind
		if kind == "" {
			kind = "main"
		}
		switch kind {
		case "answer":
			c.mergeAnswer("plugin", r, 0)
		case "infobox":
			c.mergeInfobox("plugin", r, 0)
		case "suggestion":
			c.mergeSuggestion(r)
		case "correction":
			c.mergeCorrection(r)
		default:
			c.mergeResult("plugin", kind, r, 0)
		}
	}
}
