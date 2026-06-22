package metrics

import (
	"math"
	"sync"
	"testing"
)

func TestHistogramWindowPercentiles(t *testing.T) {
	w := NewHistogramWindow(100)
	p50, p80, p95 := w.Percentiles()
	if p50 != 0 || p80 != 0 || p95 != 0 {
		t.Fatalf("empty window: expected (0,0,0), got (%v,%v,%v)", p50, p80, p95)
	}
	for i := 1.0; i <= 100; i++ {
		w.Record(i)
	}
	p50, p80, p95 = w.Percentiles()
	if math.Abs(p50-50) > 1 {
		t.Errorf("p50 expected ~50, got %v", p50)
	}
	if math.Abs(p80-80) > 1 {
		t.Errorf("p80 expected ~80, got %v", p80)
	}
	if math.Abs(p95-95) > 1 {
		t.Errorf("p95 expected ~95, got %v", p95)
	}
	for i := 101.0; i <= 150; i++ {
		w.Record(i)
	}
	p50, _, _ = w.Percentiles()
	if p50 < 100 {
		t.Errorf("after overflow p50 expected >100, got %v", p50)
	}
}

func TestHistogramWindowMinMax(t *testing.T) {
	w := NewHistogramWindow(10)
	if got := w.Min(); got != 0 {
		t.Errorf("empty window Min expected 0, got %v", got)
	}
	if got := w.Max(); got != 0 {
		t.Errorf("empty window Max expected 0, got %v", got)
	}
	w.Record(3.0)
	w.Record(7.0)
	if w.Min() != 3.0 {
		t.Errorf("Min expected 3, got %v", w.Min())
	}
	if w.Max() != 7.0 {
		t.Errorf("Max expected 7, got %v", w.Max())
	}
	if w.Count() != 2 {
		t.Errorf("Count expected 2, got %d", w.Count())
	}
}

func TestHistogramWindowConcurrent(t *testing.T) {
	w := NewHistogramWindow(1000)
	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w.Record(float64(i))
			}
		}()
	}
	wg.Wait()
	if w.Count() != 1000 {
		t.Errorf("concurrent Count expected 1000, got %d", w.Count())
	}
}
