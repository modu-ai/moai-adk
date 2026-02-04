package ops

import (
	"sync"
	"sync/atomic"
	"time"
)

// StatsTracker tracks performance and cache statistics for Git operations.
type StatsTracker struct {
	mu sync.RWMutex

	// Operation stats (atomic for high-frequency updates)
	totalOps    atomic.Int64
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64
	errors      atomic.Int64
	totalTime   atomic.Int64

	// Queue stats
	pending atomic.Int32

	// Cache stats (protected by mutex)
	cacheStats CacheStats
}

// NewStatsTracker creates a new statistics tracker.
func NewStatsTracker() *StatsTracker {
	return &StatsTracker{}
}

// RecordOperation records the completion of a Git operation.
func (st *StatsTracker) RecordOperation(duration time.Duration, cacheHit bool, hasError bool) {
	st.totalOps.Add(1)
	st.totalTime.Add(int64(duration))

	if cacheHit {
		st.cacheHits.Add(1)
	} else {
		st.cacheMisses.Add(1)
	}

	if hasError {
		st.errors.Add(1)
	}
}

// SetPending sets the current number of pending operations.
func (st *StatsTracker) SetPending(count int) {
	st.pending.Store(int32(count))
}

// IncrPending increments the pending operation count.
func (st *StatsTracker) IncrPending() {
	st.pending.Add(1)
}

// DecrPending decrements the pending operation count.
// The count will not go below zero.
func (st *StatsTracker) DecrPending() {
	for {
		current := st.pending.Load()
		if current <= 0 {
			return
		}
		if st.pending.CompareAndSwap(current, current-1) {
			return
		}
	}
}

// SetCacheStats updates the cache statistics.
func (st *StatsTracker) SetCacheStats(stats CacheStats) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.cacheStats = stats
}

// GetStats returns the current statistics snapshot.
func (st *StatsTracker) GetStats() Statistics {
	st.mu.RLock()
	cacheStats := st.cacheStats
	st.mu.RUnlock()

	total := st.totalOps.Load()
	hits := st.cacheHits.Load()
	misses := st.cacheMisses.Load()
	errs := st.errors.Load()
	totalTime := st.totalTime.Load()
	pending := st.pending.Load()

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	var avgTime float64
	if total > 0 {
		avgTime = float64(totalTime) / float64(total)
	}

	return Statistics{
		Operations: OperationStats{
			Total:            int(total),
			CacheHits:        int(hits),
			CacheMisses:      int(misses),
			CacheHitRate:     hitRate,
			Errors:           int(errs),
			AvgExecutionTime: avgTime,
			TotalTime:        totalTime,
		},
		Cache: cacheStats,
		Queue: QueueStats{
			Pending: int(pending),
		},
	}
}

// Reset resets all statistics to zero.
func (st *StatsTracker) Reset() {
	st.totalOps.Store(0)
	st.cacheHits.Store(0)
	st.cacheMisses.Store(0)
	st.errors.Store(0)
	st.totalTime.Store(0)
	st.pending.Store(0)

	st.mu.Lock()
	st.cacheStats = CacheStats{}
	st.mu.Unlock()
}
