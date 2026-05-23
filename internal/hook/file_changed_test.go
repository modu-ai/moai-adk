package hook

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/testutil"
)

func TestFileChangedHandler_EventType(t *testing.T) {
	h := NewFileChangedHandler()
	if h.EventType() != EventFileChanged {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventFileChanged)
	}
}

func TestFileChangedHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *HookInput
		createFile  bool
		fileContent string
	}{
		{
			name: "deleted file - skip",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "/tmp/deleted.go",
				ChangeType:    "deleted",
				HookEventName: "FileChanged",
			},
			createFile: false,
		},
		{
			name: "unsupported extension",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "/tmp/file.txt",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile: false,
		},
		{
			name: "supported Go file without tags",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "test.go",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:  true,
			fileContent: "package main\n\nfunc main() {}\n",
		},
		{
			name: "supported Go file with tags",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "test.go",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:  true,
			fileContent: "// @MX:NOTE: This is a note\npackage main\n",
		},
		{
			name: "Python file with tags",
			input: &HookInput{
				SessionID:     "test-session",
				FilePath:      "test.py",
				ChangeType:    "modified",
				HookEventName: "FileChanged",
			},
			createFile:  true,
			fileContent: "# @MX:NOTE: Python note\nprint('hello')\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewFileChangedHandler().(*fileChangedHandler)

			// Create temp file if needed
			if tt.createFile {
				tempDir := t.TempDir()
				filePath := tt.input.FilePath
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(tempDir, filepath.Base(filePath))
				}

				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				tt.input.FilePath = filePath
			}

			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}

			// REQ-HAE-001: async path emits empty HookOutput; SystemMessage
			// is logged at info level only.
			if out.SystemMessage != "" {
				t.Errorf("expected empty SystemMessage in async mode, got: %q", out.SystemMessage)
			}

			// Drain any spawned goroutines deterministically (REQ-HAE-006).
			testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
		})
	}
}

// TestFileChanged_AsyncReturn_Under100ms verifies REQ-HAE-001: main return
// path completes within ≤ 100 ms regardless of side-effect duration.
// AC-HAE-002 covers the formal p95 ≤ 100ms benchmark below; this is a
// per-call sanity check.
func TestFileChanged_AsyncReturn_Under100ms(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sample.go")
	if err := os.WriteFile(path, []byte("package main\n// @MX:NOTE: x\nfunc main(){}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	h := NewFileChangedHandler().(*fileChangedHandler)
	input := &HookInput{
		SessionID:     "t1",
		FilePath:      path,
		ChangeType:    "modified",
		HookEventName: "FileChanged",
		CWD:           tempDir,
	}

	start := time.Now()
	out, err := h.Handle(context.Background(), input)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("synchronous return took %v, want ≤ 100ms (REQ-HAE-001)", elapsed)
	}

	testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
}

// TestFileChanged_SideEffectsCompleted verifies the side-effect goroutine
// actually executes (sidecar updated) when WaitForAsync drains the WG.
func TestFileChanged_SideEffectsCompleted(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "tagged.go")
	if err := os.WriteFile(path, []byte("// @MX:ANCHOR: side\n// @MX:REASON: test\npackage main\nfunc main(){}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	h := NewFileChangedHandler().(*fileChangedHandler)
	input := &HookInput{
		SessionID:     "t2",
		FilePath:      path,
		ChangeType:    "modified",
		HookEventName: "FileChanged",
		CWD:           tempDir,
	}
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

	// Wait for the goroutine to complete its sidecar update.
	testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)

	// The sidecar index lives under <CWD>/.moai/state/. Its exact
	// filename depends on mx.Manager internals; verify the directory
	// exists as a proxy for "the async side-effect ran".
	stateDir := filepath.Join(tempDir, ".moai", "state")
	if _, err := os.Stat(stateDir); err != nil {
		t.Errorf("expected MX state dir %s to exist after async scan, got: %v", stateDir, err)
	}
}

// BenchmarkFileChanged_AsyncReturn measures the p95 latency of the main
// return path under 10 concurrent invocations. AC-HAE-002 requires p95 ≤ 100ms.
// The metric `p95-ms` is registered via b.ReportMetric for grep-able output.
func BenchmarkFileChanged_AsyncReturn(b *testing.B) {
	tempDir := b.TempDir()
	path := filepath.Join(tempDir, "bench.go")
	if err := os.WriteFile(path, []byte("// @MX:NOTE: bench\npackage main\nfunc main(){}\n"), 0644); err != nil {
		b.Fatalf("write: %v", err)
	}

	const concurrency = 10
	durations := make([]time.Duration, 0, b.N*concurrency)
	var mu sync.Mutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := NewFileChangedHandler().(*fileChangedHandler)
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				input := &HookInput{
					SessionID:     "bench",
					FilePath:      path,
					ChangeType:    "modified",
					HookEventName: "FileChanged",
					CWD:           tempDir,
				}
				start := time.Now()
				_, _ = h.Handle(context.Background(), input)
				d := time.Since(start)
				mu.Lock()
				durations = append(durations, d)
				mu.Unlock()
			}()
		}
		wg.Wait()
		// Drain async goroutines so the benchmark doesn't leak.
		testutil.WaitForAsync(b, h.waitGroup(), 5*time.Second)
	}
	b.StopTimer()

	p95Ms := percentileMillis(durations, 0.95)
	b.ReportMetric(p95Ms, "p95-ms")
	if p95Ms > 100 {
		b.Errorf("AC-HAE-002 violation: p95 = %.2f ms, want ≤ 100 ms", p95Ms)
	}

	// Sanity: no goroutine should leak after the benchmark.
	finalCount := atomic.LoadInt64(new(int64)) // placeholder for leak counter
	_ = finalCount
}

// percentileMillis computes the percentile of a duration slice in milliseconds.
// p ∈ [0, 1]. Returns 0 if the slice is empty.
func percentileMillis(durations []time.Duration, p float64) float64 {
	if len(durations) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return float64(sorted[idx].Microseconds()) / 1000.0
}

func TestFileChangedHandler_SupportedExtensions(t *testing.T) {
	tests := []struct {
		ext       string
		supported bool
	}{
		{".go", true},
		{".py", true},
		{".ts", true},
		{".js", true},
		{".rs", true},
		{".java", true},
		{".kt", true},
		{".cs", true},
		{".rb", true},
		{".php", true},
		{".ex", true},
		{".exs", true},
		{".cpp", true},
		{".cc", true},
		{".cxx", true},
		{".h", true},
		{".hpp", true},
		{".scala", true},
		{".r", true},
		{".dart", true},
		{".swift", true},
		{".txt", false},
		{".md", false},
		{".json", false},
		{".yaml", false},
		{".yml", false},
		{".toml", false},
		{".xml", false},
		{".html", false},
		{".css", false},
		{".sh", false},
		{".bash", false},
		{".zsh", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			if supportedExtensions[tt.ext] != tt.supported {
				t.Errorf("Extension %v: supported=%v, want %v", tt.ext, supportedExtensions[tt.ext], tt.supported)
			}
		})
	}
}
