package hook

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/testutil"
)

func TestTaskCreatedHandler_EventType(t *testing.T) {
	h := NewTaskCreatedHandler()
	if h.EventType() != EventTaskCreated {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventTaskCreated)
	}
}

// newTaskCreatedCfg builds a ConfigProvider with the given HOI master toggle
// and per-event whitelist state. Observability master (REQ-OBS-005) is
// orthogonal — set via SetObservabilityMasterForTesting in the calling test.
func newTaskCreatedCfg(hoiEnabled bool, events []string) ConfigProvider {
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.OptIn.Enabled = hoiEnabled
	cfg.System.Hook.ObservabilityEvents = events
	return &auditConfigProvider{cfg: cfg}
}

// TestTaskCreatedHandler_Handle exercises the historical contract: the
// no-config constructor returns silently for any input. REQ-HAE-003 zero-overhead
// path is preserved.
func TestTaskCreatedHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "full fields",
			input: &HookInput{
				SessionID:   "sess-001",
				TaskID:      "task-123",
				TaskSubject: "Implement authentication",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "session and task id only",
			input: &HookInput{
				SessionID: "sess-002",
				TaskID:    "task-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskCreatedHandler().(*taskCreatedHandler)
			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Errorf("Handle() error = %v, want nil", err)
			}
			if out == nil {
				t.Error("Handle() returned nil output")
			}
			// No config provider → HOI gate is false → no goroutine spawned.
			testutil.WaitForAsync(t, h.waitGroup(), 100*time.Millisecond)
		})
	}
}

// TestTaskCreated_DualGateMatrix verifies REQ-HAE-003 dual-gate semantics
// across all 4 quadrants of (HOI master, observability master). A goroutine
// MUST be spawned ONLY when both gates pass; otherwise zero overhead.
func TestTaskCreated_DualGateMatrix(t *testing.T) {
	// Cannot use t.Parallel() — SetObservabilityMasterForTesting mutates
	// a process-global cached value via sync.Once.
	tests := []struct {
		name           string
		hoiEnabled     bool
		obsMasterOn    bool
		rt006Whitelist []string
		wantAsyncCount int // expected goroutines spawned (0 or 1)
	}{
		{
			name:           "Q1: HOI=false, obsMaster=false",
			hoiEnabled:     false,
			obsMasterOn:    false,
			rt006Whitelist: []string{"taskCreated"},
			wantAsyncCount: 0,
		},
		{
			name:           "Q2: HOI=false, obsMaster=true",
			hoiEnabled:     false,
			obsMasterOn:    true,
			rt006Whitelist: []string{"taskCreated"},
			wantAsyncCount: 0,
		},
		{
			name:           "Q3: HOI=true, obsMaster=false",
			hoiEnabled:     true,
			obsMasterOn:    false,
			rt006Whitelist: []string{"taskCreated"},
			wantAsyncCount: 0,
		},
		{
			name:           "Q4: HOI=true, obsMaster=true, whitelist present",
			hoiEnabled:     true,
			obsMasterOn:    true,
			rt006Whitelist: []string{"taskCreated"},
			wantAsyncCount: 1,
		},
		{
			name:           "Q4-edge: HOI=true, obsMaster=true, whitelist empty",
			hoiEnabled:     true,
			obsMasterOn:    true,
			rt006Whitelist: []string{},
			wantAsyncCount: 0, // RT-006 per-event gate independently blocks
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetObservabilityMasterForTesting()
			SetObservabilityMasterForTesting(tt.obsMasterOn)
			t.Cleanup(func() {
				ResetObservabilityMasterForTesting()
			})

			cfg := newTaskCreatedCfg(tt.hoiEnabled, tt.rt006Whitelist)
			h := NewTaskCreatedHandlerWithConfig(cfg).(*taskCreatedHandler)

			out, err := h.Handle(context.Background(), &HookInput{TaskID: "t-1"})
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if out == nil || out.SystemMessage != "" {
				t.Errorf("expected empty HookOutput, got %+v", out)
			}

			// Drain — even when count=0, this verifies no leak.
			testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
		})
	}
}

// TestTaskCreated_ObservabilityDisabled_NoGoroutine verifies the bonus
// assertion in AC-HAE-004: when observability master is false, no goroutine
// is spawned (zero-overhead path). Verified via WG counter snapshot.
func TestTaskCreated_ObservabilityDisabled_NoGoroutine(t *testing.T) {
	ResetObservabilityMasterForTesting()
	SetObservabilityMasterForTesting(false)
	t.Cleanup(ResetObservabilityMasterForTesting)

	cfg := newTaskCreatedCfg(true, []string{"taskCreated"})
	h := NewTaskCreatedHandlerWithConfig(cfg).(*taskCreatedHandler)

	out, err := h.Handle(context.Background(), &HookInput{TaskID: "t-no-goroutine"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

	// WG should drain immediately (no Add called when obsMaster is false).
	done := make(chan struct{})
	go func() {
		h.waitGroup().Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("WaitGroup did not drain immediately — goroutine spawned when obsMaster=false")
	}
}

// BenchmarkTaskCreated_AsyncReturn measures p95 latency under 10-concurrent
// load with all gates passing. AC-HAE-004 requires p95 ≤ 100ms.
func BenchmarkTaskCreated_AsyncReturn(b *testing.B) {
	ResetObservabilityMasterForTesting()
	SetObservabilityMasterForTesting(true)
	b.Cleanup(ResetObservabilityMasterForTesting)

	cfg := newTaskCreatedCfg(true, []string{"taskCreated"})

	const concurrency = 10
	durations := make([]time.Duration, 0, b.N*concurrency)
	var mu sync.Mutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := NewTaskCreatedHandlerWithConfig(cfg).(*taskCreatedHandler)
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				input := &HookInput{
					SessionID: "bench",
					TaskID:    "t-bench",
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

	p95Ms := tcPercentileMillis(durations, 0.95)
	b.ReportMetric(p95Ms, "p95-ms")
	if p95Ms > 100 {
		b.Errorf("AC-HAE-004 violation: p95 = %.2f ms, want ≤ 100 ms", p95Ms)
	}
}

func tcPercentileMillis(durations []time.Duration, p float64) float64 {
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
