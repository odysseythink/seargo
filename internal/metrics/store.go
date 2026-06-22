package metrics

import (
	"sync"
	"time"
)

// TimeStats holds p50/p80/p95 values for a timing dimension.
type TimeStats struct {
	P50 float64 `json:"p50"`
	P80 float64 `json:"p80"`
	P95 float64 `json:"p95"`
}

// CountStats holds min/max/avg for a counter dimension.
type CountStats struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
}

// EngineSnapshot is a point-in-time read of per-engine stats.
type EngineSnapshot struct {
	Engine       string             `json:"engine"`
	Reliability  float64            `json:"reliability"`
	Score        float64            `json:"score"`
	Time         EngineTimeSnapshot `json:"time"`
	ResultCount  CountStats         `json:"result_count"`
	RequestCount uint64             `json:"request_count"`
	SuccessCount uint64             `json:"success_count"`
	ErrorCounts  map[string]uint64  `json:"error_counts"`
	Suspended    bool               `json:"suspended"`
	LastErrorAt  *time.Time         `json:"last_error_at,omitempty"`
}

// EngineTimeSnapshot holds timing snapshots for each dimension.
type EngineTimeSnapshot struct {
	Total      TimeStats `json:"total"`
	HTTP       TimeStats `json:"http"`
	Processing TimeStats `json:"processing"`
}

// engineStats is the internal mutable per-engine data.
type engineStats struct {
	mu             sync.Mutex
	totalTime      *HistogramWindow
	httpTime       *HistogramWindow
	processingTime *HistogramWindow
	resultCount    *HistogramWindow
	requestCount   uint64
	successCount   uint64
	errorCounts    map[string]uint64
	scoreSum       float64
	scoreCount     uint64
	suspended      bool
	lastErrorAt    *time.Time
}

// EngineStatsStore is a thread-safe in-memory store of per-engine statistics.
type EngineStatsStore struct {
	windowSize int
	mu         sync.RWMutex
	engines    map[string]*engineStats
}

// NewEngineStatsStore creates a store with the given sliding window size per engine.
func NewEngineStatsStore(windowSize int) *EngineStatsStore {
	if windowSize <= 0 {
		windowSize = 100
	}
	return &EngineStatsStore{
		windowSize: windowSize,
		engines:    make(map[string]*engineStats),
	}
}

func (s *EngineStatsStore) ensureEngine(name string) *engineStats {
	s.mu.RLock()
	es, ok := s.engines[name]
	s.mu.RUnlock()
	if ok {
		return es
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if es, ok = s.engines[name]; ok {
		return es
	}
	es = &engineStats{
		totalTime:      NewHistogramWindow(s.windowSize),
		httpTime:       NewHistogramWindow(s.windowSize),
		processingTime: NewHistogramWindow(s.windowSize),
		resultCount:    NewHistogramWindow(s.windowSize),
		errorCounts:    make(map[string]uint64),
	}
	s.engines[name] = es
	return es
}

// lookupEngine returns the engine stats without creating a new entry.
func (s *EngineStatsStore) lookupEngine(name string) *engineStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engines[name]
}

// Record logs a completed engine search with timing and error info.
func (s *EngineStatsStore) Record(engine string, totalDur, httpDur time.Duration, resultCount int, err error) {
	ts := totalDur.Seconds()
	hs := httpDur.Seconds()
	ps := ts - hs
	if ps < 0 {
		ps = 0
	}
	es := s.ensureEngine(engine)
	es.totalTime.Record(ts)
	es.httpTime.Record(hs)
	es.processingTime.Record(ps)
	es.resultCount.Record(float64(resultCount))
	es.mu.Lock()
	es.requestCount++
	if err != nil {
		class := ClassifyError(err)
		if class != "" {
			es.errorCounts[class]++
		}
		t := time.Now()
		es.lastErrorAt = &t
	} else {
		es.successCount++
	}
	es.mu.Unlock()
}

// RecordWithScore is like Record but also records result score metrics.
func (s *EngineStatsStore) RecordWithScore(engine string, totalDur, httpDur time.Duration, resultCount int, err error, scoreSum float64) {
	s.Record(engine, totalDur, httpDur, resultCount, err)
	es := s.ensureEngine(engine)
	es.mu.Lock()
	es.scoreSum += scoreSum
	es.scoreCount += uint64(resultCount)
	es.mu.Unlock()
}

// SetSuspended updates the suspension state of an engine.
func (s *EngineStatsStore) SetSuspended(engine string, suspended bool) {
	es := s.ensureEngine(engine)
	es.mu.Lock()
	es.suspended = suspended
	es.mu.Unlock()
}

// Snapshot returns a read-only snapshot of engine stats, or nil if the engine has not been created yet.
func (s *EngineStatsStore) Snapshot(name string) *EngineSnapshot {
	es := s.lookupEngine(name)
	if es == nil {
		return nil
	}
	es.mu.Lock()
	totalP50, totalP80, totalP95 := es.totalTime.Percentiles()
	httpP50, httpP80, httpP95 := es.httpTime.Percentiles()
	procP50, procP80, procP95 := es.processingTime.Percentiles()
	resultMin := es.resultCount.Min()
	resultMax := es.resultCount.Max()
	resultAvg := es.resultCount.Avg()
	reqCount := es.requestCount
	succCount := es.successCount
	errCounts := copyErrorCounts(es.errorCounts)
	totalErr := totalErrors(es)
	susp := es.suspended
	lastErr := es.lastErrorAt
	scoreSum := es.scoreSum
	scoreCnt := es.scoreCount
	es.mu.Unlock()
	var reliability float64
	total := float64(succCount + totalErr)
	if total > 0 {
		reliability = float64(succCount) / total
	} else {
		reliability = 1.0
	}
	var avgScore float64
	if scoreCnt > 0 {
		avgScore = scoreSum / float64(scoreCnt)
	}
	return &EngineSnapshot{
		Engine:      name,
		Reliability: reliability,
		Score:       avgScore,
		Time: EngineTimeSnapshot{
			Total:      TimeStats{P50: totalP50, P80: totalP80, P95: totalP95},
			HTTP:       TimeStats{P50: httpP50, P80: httpP80, P95: httpP95},
			Processing: TimeStats{P50: procP50, P80: procP80, P95: procP95},
		},
		ResultCount: CountStats{
			Min: resultMin,
			Max: resultMax,
			Avg: resultAvg,
		},
		RequestCount: reqCount,
		SuccessCount: succCount,
		ErrorCounts:  errCounts,
		Suspended:    susp,
		LastErrorAt:  lastErr,
	}
}

// SnapshotAll returns snapshots for all engines that have data.
func (s *EngineStatsStore) SnapshotAll() []*EngineSnapshot {
	s.mu.RLock()
	names := make([]string, 0, len(s.engines))
	for n := range s.engines {
		names = append(names, n)
	}
	s.mu.RUnlock()
	result := make([]*EngineSnapshot, 0, len(names))
	for _, n := range names {
		if snap := s.Snapshot(n); snap != nil {
			result = append(result, snap)
		}
	}
	return result
}

func totalErrors(es *engineStats) uint64 {
	var total uint64
	for _, c := range es.errorCounts {
		total += c
	}
	return total
}

func copyErrorCounts(src map[string]uint64) map[string]uint64 {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]uint64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
