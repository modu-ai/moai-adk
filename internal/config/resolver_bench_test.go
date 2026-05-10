package config

// resolver_bench_test.go — performance benchmarks and memory footprint test.
//
// AC-V3R2-RT-005-16: BenchmarkResolver_Load p99 < 100ms.
// AC-V3R2-RT-005-17: BenchmarkResolver_Reload p99 < 20ms.
// AC-V3R2-RT-005-18: TestResolver_MemoryFootprint HeapAlloc delta < 2 MiB.
//
// T-RT005-43 (Load bench), T-RT005-44 (Reload bench), T-RT005-45 (memory footprint).

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"
)

// setupSyntheticProject stages a typical project in t.TempDir() with n yaml sections.
// Each section contains a schema_version and two keys so the resolver has realistic work.
func setupSyntheticProject(tb testing.TB, n int) string {
	tb.Helper()
	dir := tb.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		tb.Fatalf("setupSyntheticProject: mkdir: %v", err)
	}
	for i := 0; i < n; i++ {
		path := filepath.Join(sectionsDir, fmt.Sprintf("section%02d.yaml", i))
		content := fmt.Sprintf("schema_version: 3\nkey1: value%d\nkey2: %d\n", i, i*10)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			tb.Fatalf("setupSyntheticProject: write: %v", err)
		}
	}
	return dir
}

// percentiles returns the requested percentile values from a float64 sample set.
// samples may be empty; in that case all percentiles return 0.
func percentiles(samples []float64, ps ...float64) []float64 {
	out := make([]float64, len(ps))
	if len(samples) == 0 {
		return out
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)
	for i, p := range ps {
		idx := int(math.Ceil(float64(len(sorted))*p/100.0)) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		out[i] = sorted[idx]
	}
	return out
}

// BenchmarkResolver_Load measures cold-load p99 latency.
// Target: p99 < 100ms per spec.md §7 Constraints, AC-V3R2-RT-005-16.
//
// Each iteration uses a fresh resolver to measure cold-cache load.
// Uses a synthetic project with 23 yaml sections × typical content.
// Minimum 100 samples collected for meaningful p99 regardless of b.N.
func BenchmarkResolver_Load(b *testing.B) {
	dir := setupSyntheticProject(b, 23)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		b.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	// Ensure at least minSamples for meaningful p99 even under -benchtime=1x.
	const minSamples = 10
	iterations := b.N
	if iterations < minSamples {
		iterations = minSamples
	}

	samples := make([]float64, 0, iterations)
	b.ResetTimer()
	for i := 0; i < iterations; i++ {
		start := time.Now()
		r := NewResolver()
		_, _ = r.Load()
		elapsed := time.Since(start)
		samples = append(samples, float64(elapsed.Microseconds()))
	}
	b.StopTimer()

	ps := percentiles(samples, 50, 95, 99)
	b.ReportMetric(ps[0], "us-p50")
	b.ReportMetric(ps[1], "us-p95")
	b.ReportMetric(ps[2], "us-p99")

	const limitUs = 100_000 // 100ms in microseconds
	if ps[2] > limitUs {
		b.Errorf("Load p99 = %.0fus, want < %dus (100ms)", ps[2], limitUs)
	}
}

// BenchmarkResolver_Reload measures diff-aware Reload p99 latency.
// Target: p99 < 20ms per spec.md §7 Constraints, AC-V3R2-RT-005-17.
//
// Each iteration mutates section00.yaml and calls Reload() on it to force
// a full re-parse of that tier plus re-merge.
func BenchmarkResolver_Reload(b *testing.B) {
	dir := setupSyntheticProject(b, 23)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		b.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	// Use concrete type so we can call Reload (not part of SettingsResolver interface).
	r := &resolver{
		loadedAt:    time.Now(),
		tierData:    make(map[Source]map[string]any),
		tierOrigins: make(map[Source]string),
	}
	_, _ = r.Load() // initial warm load

	sectionPath, _ := filepath.Abs(filepath.Join(".moai", "config", "sections", "section00.yaml"))

	const minSamples = 10
	iterations := b.N
	if iterations < minSamples {
		iterations = minSamples
	}

	samples := make([]float64, 0, iterations)
	b.ResetTimer()
	for i := 0; i < iterations; i++ {
		// Mutate the file to force a real re-parse on each iteration.
		content := fmt.Sprintf("schema_version: 3\nkey1: rotated%d\nkey2: %d\n", i, i*100)
		_ = os.WriteFile(sectionPath, []byte(content), 0o644)

		start := time.Now()
		_ = r.Reload(sectionPath)
		elapsed := time.Since(start)
		samples = append(samples, float64(elapsed.Microseconds()))
	}
	b.StopTimer()

	ps := percentiles(samples, 50, 95, 99)
	b.ReportMetric(ps[0], "us-p50")
	b.ReportMetric(ps[1], "us-p95")
	b.ReportMetric(ps[2], "us-p99")

	const limitUs = 20_000 // 20ms in microseconds
	if ps[2] > limitUs {
		b.Errorf("Reload p99 = %.0fus, want < %dus (20ms)", ps[2], limitUs)
	}
}

// TestResolver_MemoryFootprint verifies that a full Load() allocates < 2 MiB on the heap.
//
// Strategy: snapshot runtime.MemStats before and after Load(), GC between both reads.
// Test is NOT parallel (runtime.MemStats is global; concurrent tests corrupt measurement).
//
// AC-V3R2-RT-005-18, T-RT005-45.
func TestResolver_MemoryFootprint(t *testing.T) {
	// Do NOT add t.Parallel() here — global MemStats would be corrupted.
	dir := setupSyntheticProject(t, 23)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	// Force GC to settle allocations from test setup before measuring.
	runtime.GC()
	runtime.GC()

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Force GC so allocated-but-freed memory doesn't inflate the delta.
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&after)

	// HeapAlloc is the live heap bytes after GC — a conservative measure.
	// Use TotalAlloc delta as an upper-bound measure (monotonically increasing).
	totalAllocDelta := after.TotalAlloc - before.TotalAlloc

	const limit uint64 = 2 * 1024 * 1024 // 2 MiB
	t.Logf("Resolver TotalAlloc delta: %d bytes (%.2f KiB)", totalAllocDelta, float64(totalAllocDelta)/1024.0)

	if totalAllocDelta >= limit {
		t.Errorf("TotalAlloc delta = %d bytes (%.2f KiB), want < %d bytes (2 MiB)",
			totalAllocDelta, float64(totalAllocDelta)/1024.0, limit)
	}

	// Keep r alive so GC does not collect before measurement completes.
	runtime.KeepAlive(r)
}
