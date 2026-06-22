package metrics

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestEngineStatsStoreRecordSuccess(t *testing.T) {
	ts := NewEngineStatsStore(100)
	ts.Record("google", 500*time.Millisecond, 300*time.Millisecond, 10, nil)
	snap := ts.Snapshot("google")
	if snap == nil {
		t.Fatal("expected snapshot for google")
	}
	if snap.Engine != "google" {
		t.Errorf("engine name mismatch: %s", snap.Engine)
	}
	if snap.RequestCount != 1 {
		t.Errorf("RequestCount expected 1, got %d", snap.RequestCount)
	}
	if snap.SuccessCount != 1 {
		t.Errorf("SuccessCount expected 1, got %d", snap.SuccessCount)
	}
	if snap.Reliability != 1.0 {
		t.Errorf("Reliability expected 1.0, got %v", snap.Reliability)
	}
	if snap.Time.Total.P50 <= 0 {
		t.Error("TotalTime P50 should be > 0")
	}
	if snap.Time.HTTP.P50 <= 0 {
		t.Error("HTTPTime P50 should be > 0")
	}
}

func TestEngineStatsStoreRecordError(t *testing.T) {
	ts := NewEngineStatsStore(100)
	err := errors.New("request timeout")
	ts.Record("bing", 2*time.Second, 2*time.Second, 0, err)
	snap := ts.Snapshot("bing")
	if snap.RequestCount != 1 {
		t.Errorf("RequestCount expected 1, got %d", snap.RequestCount)
	}
	if snap.SuccessCount != 0 {
		t.Errorf("SuccessCount expected 0, got %d", snap.SuccessCount)
	}
	if snap.Reliability != 0.0 {
		t.Errorf("Reliability expected 0.0 after error, got %v", snap.Reliability)
	}
	if snap.ErrorCounts["timeout"] != 1 {
		t.Errorf("timeout count expected 1, got %v", snap.ErrorCounts)
	}
}

func TestEngineStatsStoreReliabilityMixed(t *testing.T) {
	ts := NewEngineStatsStore(100)
	for i := 0; i < 7; i++ {
		ts.Record("ddg", 100*time.Millisecond, 80*time.Millisecond, 5, nil)
	}
	for i := 0; i < 3; i++ {
		ts.Record("ddg", 2*time.Second, 2*time.Second, 0, errors.New("captcha"))
	}
	snap := ts.Snapshot("ddg")
	if snap.Reliability < 0.69 || snap.Reliability > 0.71 {
		t.Errorf("Reliability expected ~0.7, got %v", snap.Reliability)
	}
}

func TestEngineStatsStoreScoreAverage(t *testing.T) {
	ts := NewEngineStatsStore(100)
	ts.RecordWithScore("wiki", 100*time.Millisecond, 50*time.Millisecond, 3, nil, 2.4)
	ts.RecordWithScore("wiki", 200*time.Millisecond, 100*time.Millisecond, 2, nil, 1.6)
	snap := ts.Snapshot("wiki")
	if snap.Score < 0.79 || snap.Score > 0.81 {
		t.Errorf("Score avg expected ~0.8 (4.0/5), got %v", snap.Score)
	}
}

func TestEngineStatsStoreSuspension(t *testing.T) {
	ts := NewEngineStatsStore(100)
	ts.SetSuspended("brave", true)
	if !ts.Snapshot("brave").Suspended {
		t.Error("expected suspended")
	}
	ts.SetSuspended("brave", false)
	if ts.Snapshot("brave").Suspended {
		t.Error("expected not suspended")
	}
}

func TestEngineStatsStoreUnknownEngine(t *testing.T) {
	ts := NewEngineStatsStore(100)
	snap := ts.Snapshot("nonexistent")
	if snap != nil {
		t.Error("expected nil for unknown engine")
	}
}

func TestEngineStatsStoreSnapshotAll(t *testing.T) {
	ts := NewEngineStatsStore(100)
	ts.Record("a", 100*time.Millisecond, 50*time.Millisecond, 3, nil)
	ts.Record("b", 200*time.Millisecond, 100*time.Millisecond, 5, nil)
	all := ts.SnapshotAll()
	if len(all) != 2 {
		t.Errorf("expected 2 engines, got %d", len(all))
	}
}

func TestEngineStatsStoreConcurrent(t *testing.T) {
	ts := NewEngineStatsStore(100)
	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				ts.Record("concurrent", 10*time.Millisecond, 5*time.Millisecond, 1, nil)
			}
		}(g)
	}
	wg.Wait()
	snap := ts.Snapshot("concurrent")
	if snap.RequestCount != 1000 {
		t.Errorf("concurrent RequestCount expected 1000, got %d", snap.RequestCount)
	}
}
