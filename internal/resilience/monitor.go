package resilience

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ResourceMonitorConfig holds configuration for a resource monitor.
type ResourceMonitorConfig struct {
	// MemoryThreshold is the memory usage percentage that triggers high memory alerts.
	// Default: 80.0
	MemoryThreshold float64

	// GoroutineThreshold is the goroutine count that triggers high goroutine alerts.
	// Default: 1000
	GoroutineThreshold int

	// OnStatsUpdate is called when new stats are collected.
	OnStatsUpdate func(stats ResourceStats)

	// OnHighMemory is called when memory usage exceeds the threshold.
	OnHighMemory func(stats ResourceStats)

	// OnHighGoroutines is called when goroutine count exceeds the threshold.
	OnHighGoroutines func(stats ResourceStats)
}

// ResourceMonitor tracks system resource usage.
type ResourceMonitor struct {
	config ResourceMonitorConfig

	mu         sync.RWMutex
	thresholds ResourceThresholds
	isRunning  bool
	cancelFunc context.CancelFunc
}

// NewResourceMonitor creates a new ResourceMonitor with the given configuration.
func NewResourceMonitor(config ResourceMonitorConfig) *ResourceMonitor {
	// Apply defaults
	if config.MemoryThreshold <= 0 {
		config.MemoryThreshold = 80.0
	}
	if config.GoroutineThreshold <= 0 {
		config.GoroutineThreshold = 1000
	}

	return &ResourceMonitor{
		config: config,
		thresholds: ResourceThresholds{
			MemoryPercent:  config.MemoryThreshold,
			GoroutineCount: config.GoroutineThreshold,
		},
	}
}

// GetStats returns the current system resource statistics.
func (rm *ResourceMonitor) GetStats() ResourceStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return ResourceStats{
		MemoryUsedMB:   memStats.Alloc / 1024 / 1024,
		MemoryTotalMB:  memStats.Sys / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
		CPUPercent:     0.0, // CPU percentage requires more complex measurement
	}
}

// StartMonitoring begins periodic resource monitoring.
// The monitoring continues until the context is cancelled.
func (rm *ResourceMonitor) StartMonitoring(ctx context.Context, interval time.Duration) {
	rm.mu.Lock()
	if rm.isRunning {
		rm.mu.Unlock()
		return
	}
	rm.isRunning = true
	ctx, rm.cancelFunc = context.WithCancel(ctx)
	rm.mu.Unlock()

	go rm.runMonitoring(ctx, interval)
}

// Stop stops the resource monitoring.
func (rm *ResourceMonitor) Stop() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.cancelFunc != nil {
		rm.cancelFunc()
		rm.cancelFunc = nil
	}
	rm.isRunning = false
}

// Thresholds returns the current threshold settings.
func (rm *ResourceMonitor) Thresholds() ResourceThresholds {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.thresholds
}

// SetThresholds updates the threshold settings.
func (rm *ResourceMonitor) SetThresholds(thresholds ResourceThresholds) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.thresholds = thresholds
}

// runMonitoring runs the monitoring loop.
func (rm *ResourceMonitor) runMonitoring(ctx context.Context, interval time.Duration) {
	// Perform initial collection
	rm.collectAndNotify()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			rm.mu.Lock()
			rm.isRunning = false
			rm.mu.Unlock()
			return
		case <-ticker.C:
			rm.collectAndNotify()
		}
	}
}

// collectAndNotify collects stats and triggers notifications.
func (rm *ResourceMonitor) collectAndNotify() {
	stats := rm.GetStats()

	rm.mu.RLock()
	thresholds := rm.thresholds
	onStatsUpdate := rm.config.OnStatsUpdate
	onHighMemory := rm.config.OnHighMemory
	onHighGoroutines := rm.config.OnHighGoroutines
	rm.mu.RUnlock()

	// Notify stats update
	if onStatsUpdate != nil {
		onStatsUpdate(stats)
	}

	// Check thresholds and notify
	if onHighMemory != nil && stats.IsMemoryHigh(thresholds.MemoryPercent) {
		onHighMemory(stats)
	}

	if onHighGoroutines != nil && stats.IsGoroutineHigh(thresholds.GoroutineCount) {
		onHighGoroutines(stats)
	}
}
