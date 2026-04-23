package mx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// TestValidateFiles_UnboundedConcurrency_PreFix is a characterization test that pins
// the pre-fix behavior for IMP-V3U-007 (unbounded goroutine fan-out).
// CHARACTERIZATION: pins pre-fix behavior
// Pre-fix: no semaphore → all N goroutines launch simultaneously → max concurrent > NumCPU*2.
// Post-fix: semaphore caps concurrency at runtime.NumCPU()*2 → this test PASSES.
func TestValidateFiles_UnboundedConcurrency_PreFix(t *testing.T) {
	// CHARACTERIZATION: pins pre-fix behavior
	t.Parallel()

	dir := t.TempDir()
	numFiles := 50 // exceeds NumCPU*2 on any development machine

	var files []string
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf(`package testpkg

// Func%d does something.
func Func%d() {}
`, i, i)
		path := writeGoFile(t, dir, fmt.Sprintf("file%d.go", i), content)
		files = append(files, path)
	}

	var maxConcurrent int64
	var current int64

	v := &mxValidator{
		analyzer:       nil,
		projectRoot:    dir,
		fanInThreshold: 3,
		testOnWorkerStart: func() {
			c := atomic.AddInt64(&current, 1)
			for {
				old := atomic.LoadInt64(&maxConcurrent)
				if c <= old {
					break
				}
				if atomic.CompareAndSwapInt64(&maxConcurrent, old, c) {
					break
				}
			}
		},
		testOnWorkerDone: func() {
			atomic.AddInt64(&current, -1)
		},
	}

	_, err := v.ValidateFiles(context.Background(), files)
	if err != nil {
		t.Fatalf("ValidateFiles() error = %v", err)
	}

	limit := int64(runtime.NumCPU() * 2)
	// Post-fix: semaphore enforces the cap.
	// Pre-fix: this assertion FAILS (max concurrent == numFiles > limit).
	if maxConcurrent > limit {
		t.Errorf("observed max concurrent goroutines = %d, exceeds limit = %d (runtime.NumCPU()*2): "+
			"semaphore not enforced (IMP-V3U-007)", maxConcurrent, limit)
	}
	t.Logf("max_concurrent=%d limit=%d", maxConcurrent, limit)
}

// TestValidateFiles_BoundedByNumCPUTimes2 verifies AC-UTIL-001-03:
// concurrent in-flight workers never exceed runtime.NumCPU()*2 under 200-file load.
func TestValidateFiles_BoundedByNumCPUTimes2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	numFiles := 200

	var files []string
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf(`package testpkg

// Worker%d runs background tasks.
func Worker%d() {
	go func() {}()
}
`, i, i)
		path := writeGoFile(t, dir, fmt.Sprintf("file%d.go", i), content)
		files = append(files, path)
	}

	var maxConcurrent int64
	var current int64

	v := &mxValidator{
		analyzer:       nil,
		projectRoot:    dir,
		fanInThreshold: 3,
		testOnWorkerStart: func() {
			c := atomic.AddInt64(&current, 1)
			for {
				old := atomic.LoadInt64(&maxConcurrent)
				if c <= old {
					break
				}
				if atomic.CompareAndSwapInt64(&maxConcurrent, old, c) {
					break
				}
			}
		},
		testOnWorkerDone: func() {
			atomic.AddInt64(&current, -1)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := v.ValidateFiles(ctx, files)
	if err != nil {
		t.Fatalf("ValidateFiles() error = %v", err)
	}

	if len(report.FileReports)+len(report.TimedOutFiles) != numFiles {
		t.Errorf("total processed = %d, want %d", len(report.FileReports)+len(report.TimedOutFiles), numFiles)
	}

	limit := int64(runtime.NumCPU() * 2)
	if maxConcurrent > limit {
		t.Errorf("max concurrent = %d, exceeded limit = %d (runtime.NumCPU()*2=%d) (AC-UTIL-001-03)",
			maxConcurrent, limit, limit)
	}
	t.Logf("AC-UTIL-001-03 passed: max_concurrent=%d limit=%d files=%d violations=%d",
		maxConcurrent, limit, len(report.FileReports), report.TotalViolations())
}

// TestValidateFiles_NoSendOnClosedChannel verifies that adding a semaphore does not
// cause "send on closed channel" panic under concurrent load (R8 risk mitigation).
func TestValidateFiles_NoSendOnClosedChannel(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	var files []string
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf(`package testpkg

func F%d() {}
`, i)
		files = append(files, writeGoFile(t, dir, fmt.Sprintf("f%d.go", i), content))
	}

	v := NewValidator(nil, dir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Must not panic.
	_, err := v.ValidateFiles(ctx, files)
	if err != nil {
		t.Fatalf("ValidateFiles() error = %v", err)
	}
}

// TestValidateFiles_1000Files_SemaphoreStress verifies semaphore correctness under
// 1000-file synthetic load (AC-UTIL-001-03 scale test).
func TestValidateFiles_1000Files_SemaphoreStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 1000-file stress test in short mode")
	}
	t.Parallel()

	dir := t.TempDir()
	numFiles := 1000
	var files []string

	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf(`package testpkg

// Stress%d is a test function.
func Stress%d() {
	_ = %d
}
`, i, i, i)
		name := fmt.Sprintf("stress%04d.go", i)
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		files = append(files, path)
	}

	var maxConcurrent int64
	var current int64

	v := &mxValidator{
		analyzer:       nil,
		projectRoot:    dir,
		fanInThreshold: 3,
		testOnWorkerStart: func() {
			c := atomic.AddInt64(&current, 1)
			for {
				old := atomic.LoadInt64(&maxConcurrent)
				if c <= old {
					break
				}
				if atomic.CompareAndSwapInt64(&maxConcurrent, old, c) {
					break
				}
			}
		},
		testOnWorkerDone: func() {
			atomic.AddInt64(&current, -1)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	report, err := v.ValidateFiles(ctx, files)
	if err != nil {
		t.Fatalf("ValidateFiles() error = %v", err)
	}

	limit := int64(runtime.NumCPU() * 2)
	if maxConcurrent > limit {
		t.Errorf("max concurrent = %d exceeded limit = %d (runtime.NumCPU()*2=%d)",
			maxConcurrent, limit, limit)
	}
	t.Logf("1000-file stress: max_concurrent=%d limit=%d files=%d violations=%d",
		maxConcurrent, limit, len(report.FileReports), report.TotalViolations())
}
