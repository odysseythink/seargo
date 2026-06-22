package metrics

import (
	"math"
	"sort"
	"sync"
)

// HistogramWindow is a thread-safe sliding window of float64 values
// with fixed capacity. When full, new records overwrite oldest entries.
type HistogramWindow struct {
	mu       sync.RWMutex
	buf      []float64
	capacity int
	pos      int  // next write position
	full     bool
}

func NewHistogramWindow(capacity int) *HistogramWindow {
	if capacity <= 0 {
		capacity = 100
	}
	return &HistogramWindow{
		buf:      make([]float64, capacity),
		capacity: capacity,
	}
}

func (w *HistogramWindow) Record(v float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf[w.pos] = v
	w.pos++
	if w.pos >= w.capacity {
		w.pos = 0
		w.full = true
	}
}

func (w *HistogramWindow) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.full {
		return w.capacity
	}
	return w.pos
}

// snapshot returns a sorted copy of current values. Caller must hold read lock.
func (w *HistogramWindow) snapshot() []float64 {
	n := w.pos
	if w.full {
		n = w.capacity
	}
	if n == 0 {
		return nil
	}
	s := make([]float64, n)
	copy(s, w.buf[:n])
	sort.Float64s(s)
	return s
}

func (w *HistogramWindow) Percentiles() (p50, p80, p95 float64) {
	w.mu.RLock()
	s := w.snapshot()
	w.mu.RUnlock()
	if len(s) == 0 {
		return 0, 0, 0
	}
	return percentile(s, 0.50), percentile(s, 0.80), percentile(s, 0.95)
}

func (w *HistogramWindow) Min() float64 {
	w.mu.RLock()
	s := w.snapshot()
	w.mu.RUnlock()
	if len(s) == 0 {
		return 0
	}
	return s[0]
}

func (w *HistogramWindow) Max() float64 {
	w.mu.RLock()
	s := w.snapshot()
	w.mu.RUnlock()
	if len(s) == 0 {
		return 0
	}
	return s[len(s)-1]
}

// Avg returns the arithmetic mean of values in the window.
func (w *HistogramWindow) Avg() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	n := w.pos
	if w.full {
		n = w.capacity
	}
	if n == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += w.buf[i]
	}
	return sum / float64(n)
}

// percentile returns the value at the given percentile (0..1) in a sorted slice.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := p * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := lo + 1
	frac := idx - float64(lo)
	if hi >= len(sorted) {
		return sorted[lo]
	}
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}
