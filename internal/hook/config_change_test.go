package hook

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/testutil"
)

// TestConfigChange_RT005ReloadIntegration verifies REQ-HAE-002 + SPEC-V3R2-RT-006
// REQ-011 compatibility: when ConfigChange fires for a valid YAML file, the
// main return path completes within ≤ 100ms with Continue=true, and the
// async goroutine completes successfully.
//
// Per SPEC-V3R6-HOOK-ASYNC-EXPAND-001 REQ-HAE-002, the SystemMessage reload
// confirmation is now logged via slog (async observability) rather than
// returned in the main response. Tests verify the async path completes.
func TestConfigChange_RT005ReloadIntegration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "quality.yaml")
	content := []byte("development_mode: tdd\ncoverage_target: 85\n")
	if err := os.WriteFile(yamlPath, content, 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	h := NewConfigChangeHandler().(*configChangeHandler)
	input := &HookInput{
		SessionID:      "sess-rt005",
		ConfigFilePath: yamlPath,
		HookEventName:  "ConfigChange",
	}

	start := time.Now()
	out, err := h.Handle(context.Background(), input)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
	// REQ-HAE-002: main response is non-blocking.
	if elapsed > 100*time.Millisecond {
		t.Errorf("synchronous return took %v, want ≤ 100ms (REQ-HAE-002)", elapsed)
	}
	// Continue=true preserves SDK protocol invariant for the success case.
	if !out.Continue {
		t.Errorf("expected Continue=true for valid config, got false")
	}
	// SystemMessage moves to async log per REQ-HAE-002 design.
	if out.SystemMessage != "" {
		t.Errorf("expected empty SystemMessage in async mode, got: %q", out.SystemMessage)
	}

	testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
}

// TestConfigChange_InvalidYAMLAsyncReject verifies REQ-HAE-002 + REQ-062
// compatibility: invalid YAML rejection is logged asynchronously (Warn level)
// — the main response stays Continue=true because the validation happens
// in the goroutine. This is the explicit async design tradeoff per SPEC §2.1.
//
// Old contract (sync): Continue=false + SystemMessage "Config reload rejected".
// New contract (async): main response Continue=true; rejection visible via slog Warn.
func TestConfigChange_InvalidYAMLAsyncReject(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "quality.yaml")
	// Deliberately invalid YAML.
	content := []byte(": bad yaml: :\n  key:\n")
	if err := os.WriteFile(yamlPath, content, 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	h := NewConfigChangeHandler().(*configChangeHandler)
	input := &HookInput{
		SessionID:      "sess-invalid",
		ConfigFilePath: yamlPath,
		HookEventName:  "ConfigChange",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
	// Main response is success (rejection moved to async log).
	if !out.Continue {
		t.Error("expected Continue=true (async design — validation runs in goroutine)")
	}

	// The async goroutine MUST complete (even though it logs a Warn and returns).
	testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
}

func TestConfigChangeHandler_EventType(t *testing.T) {
	h := NewConfigChangeHandler()
	if h.EventType() != EventConfigChange {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventConfigChange)
	}
}

func TestConfigChangeHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *HookInput
		createFile  bool
		fileContent string
	}{
		{
			name: "valid YAML config",
			input: &HookInput{
				SessionID:      "sess-001",
				ConfigFilePath: "quality.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:  true,
			fileContent: "development_mode: ddd\ncoverage_target: 85\n",
		},
		{
			name: "invalid YAML config (async reject)",
			input: &HookInput{
				SessionID:      "sess-002",
				ConfigFilePath: "invalid.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:  true,
			fileContent: "invalid: yaml: content:\n  - broken\n",
		},
		{
			name: "empty file",
			input: &HookInput{
				SessionID:      "sess-003",
				ConfigFilePath: "empty.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:  true,
			fileContent: "",
		},
		{
			name: "non-existent file (async warn)",
			input: &HookInput{
				SessionID:      "sess-004",
				ConfigFilePath: "/does/not/exist.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewConfigChangeHandler().(*configChangeHandler)

			// Create temp file if needed
			if tt.createFile {
				tempDir := t.TempDir()
				filePath := tt.input.ConfigFilePath
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(tempDir, filepath.Base(filePath))
				}

				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				tt.input.ConfigFilePath = filePath
			}

			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}

			// REQ-HAE-002: main response is Continue=true regardless of
			// validation outcome (validation runs in goroutine).
			if !out.Continue {
				t.Errorf("expected Continue=true (async design), got false")
			}
			// REQ-HAE-002: SystemMessage moves to async log.
			if out.SystemMessage != "" {
				t.Errorf("expected empty SystemMessage in async mode, got: %q", out.SystemMessage)
			}

			testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
		})
	}
}

func TestConfigChangeHandler_ValidateConfig(t *testing.T) {
	t.Parallel()

	h := &configChangeHandler{}

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "valid YAML",
			content:     "key: value\nnested:\n  item: 1\n",
			expectError: false,
		},
		{
			name:        "invalid YAML",
			content:     "key: value\n: badcolon\n",
			expectError: true,
		},
		{
			name:        "empty file",
			content:     "",
			expectError: false,
		},
		{
			name:        "YAML list",
			content:     "- item1\n- item2\n",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file
			tempFile, err := os.CreateTemp("", "config-test-*.yaml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tempFile.Name()) }()

			// Write content
			if _, err := tempFile.Write([]byte(tt.content)); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			_ = tempFile.Close()

			// Validate config
			err = h.validateConfig(tempFile.Name())
			if tt.expectError && err == nil {
				t.Error("expected error for invalid YAML")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// BenchmarkConfigChange_AsyncReturn measures p95 of the main return path
// under 10-concurrent load. AC-HAE-003 requires p95 ≤ 100ms.
func BenchmarkConfigChange_AsyncReturn(b *testing.B) {
	tempDir := b.TempDir()
	yamlPath := filepath.Join(tempDir, "quality.yaml")
	if err := os.WriteFile(yamlPath, []byte("development_mode: tdd\n"), 0644); err != nil {
		b.Fatalf("write: %v", err)
	}

	const concurrency = 10
	durations := make([]time.Duration, 0, b.N*concurrency)
	var mu sync.Mutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := NewConfigChangeHandler().(*configChangeHandler)
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				input := &HookInput{
					SessionID:      "bench",
					ConfigFilePath: yamlPath,
					HookEventName:  "ConfigChange",
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
		testutil.WaitForAsync(b, h.waitGroup(), 5*time.Second)
	}
	b.StopTimer()

	p95Ms := bencPercentileMillis(durations, 0.95)
	b.ReportMetric(p95Ms, "p95-ms")
	if p95Ms > 100 {
		b.Errorf("AC-HAE-003 violation: p95 = %.2f ms, want ≤ 100 ms", p95Ms)
	}
}

// bencPercentileMillis is the local benchmark percentile helper, mirroring
// percentileMillis from file_changed_test.go but with a distinct name to
// avoid duplicate symbol errors.
func bencPercentileMillis(durations []time.Duration, p float64) float64 {
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

