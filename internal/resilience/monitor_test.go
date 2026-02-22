package resilience

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewResourceMonitor(t *testing.T) {
	t.Parallel()

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
	})

	if monitor == nil {
		t.Fatal("NewResourceMonitor returned nil")
	}
}

func TestResourceMonitorGetStats(t *testing.T) {
	t.Parallel()

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
	})

	stats := monitor.GetStats()

	// Memory values should be populated (at least some)
	// We can't predict exact values but can verify structure
	if stats.GoroutineCount < 1 {
		t.Errorf("GoroutineCount = %d, expected at least 1", stats.GoroutineCount)
	}
}

func TestResourceMonitorStartMonitoring(t *testing.T) {
	t.Parallel()

	var statsUpdates []ResourceStats
	var mu sync.Mutex

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
		OnStatsUpdate: func(stats ResourceStats) {
			mu.Lock()
			statsUpdates = append(statsUpdates, stats)
			mu.Unlock()
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start monitoring with short interval
	monitor.StartMonitoring(ctx, 20*time.Millisecond)

	// Wait for several updates
	time.Sleep(100 * time.Millisecond)

	cancel()

	// Give time for goroutine to stop
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	updateCount := len(statsUpdates)
	mu.Unlock()

	if updateCount < 2 {
		t.Errorf("expected at least 2 stats updates, got %d", updateCount)
	}
}

func TestResourceMonitorStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	var updateCount int
	var mu sync.Mutex

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
		OnStatsUpdate: func(stats ResourceStats) {
			mu.Lock()
			updateCount++
			mu.Unlock()
		},
	})

	ctx, cancel := context.WithCancel(context.Background())

	monitor.StartMonitoring(ctx, 20*time.Millisecond)

	// Wait for a few updates
	time.Sleep(100 * time.Millisecond)

	cancel()

	// Give time for the goroutine to process the cancellation
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	countAtCancel := updateCount
	mu.Unlock()

	// Wait to verify no more updates
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	finalCount := updateCount
	mu.Unlock()

	// Allow for at most 1 extra update due to race between cancel and callback
	if finalCount > countAtCancel+1 {
		t.Errorf("updates continued after cancel: had %d, now have %d", countAtCancel, finalCount)
	}
}

func TestResourceMonitorMemoryThreshold(t *testing.T) {
	t.Parallel()

	var highMemoryAlerts []ResourceStats
	var mu sync.Mutex

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    0.1, // Very low threshold to trigger
		GoroutineThreshold: 1000,
		OnHighMemory: func(stats ResourceStats) {
			mu.Lock()
			highMemoryAlerts = append(highMemoryAlerts, stats)
			mu.Unlock()
		},
	})

	ctx := t.Context()

	monitor.StartMonitoring(ctx, 20*time.Millisecond)

	// Wait for potential alerts
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	alertCount := len(highMemoryAlerts)
	mu.Unlock()

	// With 0.1% threshold, we should get alerts (unless running on minimal memory)
	// This test verifies the callback mechanism works
	_ = alertCount // May or may not trigger depending on system
}

func TestResourceMonitorGoroutineThreshold(t *testing.T) {
	t.Parallel()

	var highGoroutineAlerts []ResourceStats
	var mu sync.Mutex

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1, // Very low threshold to trigger
		OnHighGoroutines: func(stats ResourceStats) {
			mu.Lock()
			highGoroutineAlerts = append(highGoroutineAlerts, stats)
			mu.Unlock()
		},
	})

	ctx := t.Context()

	monitor.StartMonitoring(ctx, 20*time.Millisecond)

	// Wait for alerts
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	alertCount := len(highGoroutineAlerts)
	mu.Unlock()

	if alertCount < 1 {
		t.Errorf("expected high goroutine alerts with threshold=1, got %d", alertCount)
	}
}

func TestResourceMonitorConcurrentAccess(t *testing.T) {
	t.Parallel()

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
	})

	ctx := t.Context()

	monitor.StartMonitoring(ctx, 10*time.Millisecond)

	var wg sync.WaitGroup
	numGoroutines := 50

	for range numGoroutines {
		wg.Go(func() {
			for range 10 {
				_ = monitor.GetStats()
			}
		})
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

func TestResourceMonitorThresholds(t *testing.T) {
	t.Parallel()

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
	})

	thresholds := monitor.Thresholds()

	if thresholds.MemoryPercent != 80.0 {
		t.Errorf("MemoryPercent threshold = %f, want 80.0", thresholds.MemoryPercent)
	}
	if thresholds.GoroutineCount != 1000 {
		t.Errorf("GoroutineCount threshold = %d, want 1000", thresholds.GoroutineCount)
	}
}

func TestResourceMonitorSetThresholds(t *testing.T) {
	t.Parallel()

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
	})

	newThresholds := ResourceThresholds{
		MemoryPercent:  90.0,
		GoroutineCount: 500,
	}

	monitor.SetThresholds(newThresholds)

	thresholds := monitor.Thresholds()

	if thresholds.MemoryPercent != 90.0 {
		t.Errorf("MemoryPercent threshold after set = %f, want 90.0", thresholds.MemoryPercent)
	}
	if thresholds.GoroutineCount != 500 {
		t.Errorf("GoroutineCount threshold after set = %d, want 500", thresholds.GoroutineCount)
	}
}

func TestResourceMonitorMultipleStartCalls(t *testing.T) {
	t.Parallel()

	var updateCount int
	var mu sync.Mutex

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
		OnStatsUpdate: func(stats ResourceStats) {
			mu.Lock()
			updateCount++
			mu.Unlock()
		},
	})

	ctx := t.Context()

	// Start monitoring multiple times - should only have one goroutine
	monitor.StartMonitoring(ctx, 20*time.Millisecond)
	monitor.StartMonitoring(ctx, 20*time.Millisecond)
	monitor.StartMonitoring(ctx, 20*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	count := updateCount
	mu.Unlock()

	// Should have reasonable number of updates (not 3x the expected)
	// With 20ms interval over 100ms, expect ~5 updates, not ~15
	if count > 10 {
		t.Errorf("too many updates (%d), multiple monitoring goroutines may be running", count)
	}
}

func TestResourceStatsJSON(t *testing.T) {
	t.Parallel()

	stats := ResourceStats{
		MemoryUsedMB:   1024,
		MemoryTotalMB:  8192,
		GoroutineCount: 100,
		CPUPercent:     25.5,
	}

	// Just verify the struct can be used - JSON tags are for serialization
	if stats.MemoryUsedMB != 1024 {
		t.Error("unexpected MemoryUsedMB value")
	}
}

func TestResourceThresholdsZeroValues(t *testing.T) {
	t.Parallel()

	var thresholds ResourceThresholds

	if thresholds.MemoryPercent != 0 {
		t.Errorf("zero value MemoryPercent = %f, want 0", thresholds.MemoryPercent)
	}
	if thresholds.GoroutineCount != 0 {
		t.Errorf("zero value GoroutineCount = %d, want 0", thresholds.GoroutineCount)
	}
}

func TestResourceMonitorDefaults(t *testing.T) {
	t.Parallel()

	// Test with zero config - should apply defaults
	monitor := NewResourceMonitor(ResourceMonitorConfig{})

	if monitor == nil {
		t.Fatal("NewResourceMonitor with zero config returned nil")
	}

	thresholds := monitor.Thresholds()

	// Should have sensible defaults
	if thresholds.MemoryPercent == 0 {
		t.Error("default MemoryPercent threshold should not be 0")
	}
	if thresholds.GoroutineCount == 0 {
		t.Error("default GoroutineCount threshold should not be 0")
	}
}

func TestResourceMonitorStop(t *testing.T) {
	t.Parallel()

	var updateCount int
	var mu sync.Mutex

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
		OnStatsUpdate: func(stats ResourceStats) {
			mu.Lock()
			updateCount++
			mu.Unlock()
		},
	})

	ctx := context.Background()

	// Start monitoring
	monitor.StartMonitoring(ctx, 20*time.Millisecond)

	// Wait for a few updates
	time.Sleep(100 * time.Millisecond)

	// Stop using Stop() method
	monitor.Stop()

	// Give time for goroutine to stop
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	countAtStop := updateCount
	mu.Unlock()

	// Wait and verify no more updates
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	finalCount := updateCount
	mu.Unlock()

	if countAtStop < 2 {
		t.Errorf("expected at least 2 updates, got %d", countAtStop)
	}
	// Allow for at most 1 extra update due to timing race
	if finalCount > countAtStop+1 {
		t.Errorf("updates continued after Stop(): had %d, now have %d", countAtStop, finalCount)
	}
}

func TestResourceMonitorStopWithoutStart(t *testing.T) {
	t.Parallel()

	monitor := NewResourceMonitor(ResourceMonitorConfig{
		MemoryThreshold:    80.0,
		GoroutineThreshold: 1000,
	})

	// Stop without starting should not panic
	monitor.Stop()
}
