package ops

import (
	"sync"
	"testing"
	"time"
)

func TestNewStatsTracker(t *testing.T) {
	st := NewStatsTracker()
	if st == nil {
		t.Fatal("NewStatsTracker returned nil")
	}

	stats := st.GetStats()
	if stats.Operations.Total != 0 {
		t.Errorf("initial Total = %d, want 0", stats.Operations.Total)
	}
}

func TestStatsTracker_RecordOperation(t *testing.T) {
	st := NewStatsTracker()

	st.RecordOperation(100*time.Millisecond, false, false)
	st.RecordOperation(200*time.Millisecond, true, false)
	st.RecordOperation(150*time.Millisecond, false, true)

	stats := st.GetStats()

	if stats.Operations.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Operations.Total)
	}
	if stats.Operations.CacheHits != 1 {
		t.Errorf("CacheHits = %d, want 1", stats.Operations.CacheHits)
	}
	if stats.Operations.CacheMisses != 2 {
		t.Errorf("CacheMisses = %d, want 2", stats.Operations.CacheMisses)
	}
	if stats.Operations.Errors != 1 {
		t.Errorf("Errors = %d, want 1", stats.Operations.Errors)
	}
}

func TestStatsTracker_CacheHitRate(t *testing.T) {
	st := NewStatsTracker()

	// 3 cache hits, 7 cache misses = 30% hit rate
	for range 3 {
		st.RecordOperation(10*time.Millisecond, true, false)
	}
	for range 7 {
		st.RecordOperation(10*time.Millisecond, false, false)
	}

	stats := st.GetStats()

	expectedRate := 0.3
	if stats.Operations.CacheHitRate != expectedRate {
		t.Errorf("CacheHitRate = %f, want %f", stats.Operations.CacheHitRate, expectedRate)
	}
}

func TestStatsTracker_CacheHitRate_ZeroOperations(t *testing.T) {
	st := NewStatsTracker()

	stats := st.GetStats()

	if stats.Operations.CacheHitRate != 0 {
		t.Errorf("CacheHitRate with zero ops = %f, want 0", stats.Operations.CacheHitRate)
	}
}

func TestStatsTracker_AvgExecutionTime(t *testing.T) {
	st := NewStatsTracker()

	st.RecordOperation(100*time.Millisecond, false, false)
	st.RecordOperation(200*time.Millisecond, false, false)
	st.RecordOperation(300*time.Millisecond, false, false)

	stats := st.GetStats()

	// Average: (100+200+300)/3 = 200ms
	expectedAvg := float64(200 * time.Millisecond)
	if stats.Operations.AvgExecutionTime != expectedAvg {
		t.Errorf("AvgExecutionTime = %f, want %f", stats.Operations.AvgExecutionTime, expectedAvg)
	}
}

func TestStatsTracker_TotalTime(t *testing.T) {
	st := NewStatsTracker()

	st.RecordOperation(100*time.Millisecond, false, false)
	st.RecordOperation(200*time.Millisecond, false, false)

	stats := st.GetStats()

	// Total: 100+200 = 300ms
	expectedTotal := int64(300 * time.Millisecond)
	if stats.Operations.TotalTime != expectedTotal {
		t.Errorf("TotalTime = %d, want %d", stats.Operations.TotalTime, expectedTotal)
	}
}

func TestStatsTracker_SetPending(t *testing.T) {
	st := NewStatsTracker()

	st.SetPending(5)
	stats := st.GetStats()
	if stats.Queue.Pending != 5 {
		t.Errorf("Pending = %d, want 5", stats.Queue.Pending)
	}

	st.SetPending(0)
	stats = st.GetStats()
	if stats.Queue.Pending != 0 {
		t.Errorf("Pending = %d, want 0", stats.Queue.Pending)
	}
}

func TestStatsTracker_IncrDecrPending(t *testing.T) {
	st := NewStatsTracker()

	st.IncrPending()
	st.IncrPending()
	st.IncrPending()

	stats := st.GetStats()
	if stats.Queue.Pending != 3 {
		t.Errorf("Pending after 3 incr = %d, want 3", stats.Queue.Pending)
	}

	st.DecrPending()

	stats = st.GetStats()
	if stats.Queue.Pending != 2 {
		t.Errorf("Pending after 1 decr = %d, want 2", stats.Queue.Pending)
	}
}

func TestStatsTracker_DecrPending_NoNegative(t *testing.T) {
	st := NewStatsTracker()

	st.DecrPending()
	st.DecrPending()

	stats := st.GetStats()
	if stats.Queue.Pending != 0 {
		t.Errorf("Pending should not go negative, got %d", stats.Queue.Pending)
	}
}

func TestStatsTracker_SetCacheStats(t *testing.T) {
	st := NewStatsTracker()

	cacheStats := CacheStats{
		Size:        50,
		SizeLimit:   100,
		Utilization: 0.5,
	}
	st.SetCacheStats(cacheStats)

	stats := st.GetStats()
	if stats.Cache.Size != 50 {
		t.Errorf("Cache.Size = %d, want 50", stats.Cache.Size)
	}
	if stats.Cache.Utilization != 0.5 {
		t.Errorf("Cache.Utilization = %f, want 0.5", stats.Cache.Utilization)
	}
}

func TestStatsTracker_Reset(t *testing.T) {
	st := NewStatsTracker()

	st.RecordOperation(100*time.Millisecond, true, false)
	st.RecordOperation(100*time.Millisecond, false, true)
	st.SetPending(10)

	st.Reset()

	stats := st.GetStats()
	if stats.Operations.Total != 0 {
		t.Errorf("Total after reset = %d, want 0", stats.Operations.Total)
	}
	if stats.Operations.CacheHits != 0 {
		t.Errorf("CacheHits after reset = %d, want 0", stats.Operations.CacheHits)
	}
	if stats.Operations.Errors != 0 {
		t.Errorf("Errors after reset = %d, want 0", stats.Operations.Errors)
	}
	if stats.Queue.Pending != 0 {
		t.Errorf("Pending after reset = %d, want 0", stats.Queue.Pending)
	}
}

func TestStatsTracker_ConcurrentAccess(t *testing.T) {
	st := NewStatsTracker()
	var wg sync.WaitGroup

	// Concurrent record operations
	for i := range 100 {
		wg.Go(func() {
			st.RecordOperation(10*time.Millisecond, i%2 == 0, false)
		})
	}

	// Concurrent pending updates
	for range 50 {
		wg.Go(func() {
			st.IncrPending()
		})
	}

	// Concurrent reads
	for range 50 {
		wg.Go(func() {
			_ = st.GetStats()
		})
	}

	wg.Wait()

	stats := st.GetStats()
	if stats.Operations.Total != 100 {
		t.Errorf("Total = %d, want 100", stats.Operations.Total)
	}
}
