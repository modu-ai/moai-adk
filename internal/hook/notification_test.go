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

func TestNotificationHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewNotificationHandler()

	if got := h.EventType(); got != EventNotification {
		t.Errorf("EventType() = %q, want %q", got, EventNotification)
	}
}

// newNotificationCfg builds a ConfigProvider with the given HOI master toggle
// and per-event whitelist state.
func newNotificationCfg(hoiEnabled bool, events []string) ConfigProvider {
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.OptIn.Enabled = hoiEnabled
	cfg.System.Hook.ObservabilityEvents = events
	return &auditConfigProvider{cfg: cfg}
}

func TestNotificationHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "notification with all fields",
			input: &HookInput{
				SessionID:        "sess-notif-1",
				Title:            "Build Complete",
				Message:          "All tests passed",
				NotificationType: "success",
				HookEventName:    "Notification",
			},
		},
		{
			name: "notification without title",
			input: &HookInput{
				SessionID:     "sess-notif-2",
				Message:       "Info msg",
				HookEventName: "Notification",
			},
		},
		{
			name: "empty notification",
			input: &HookInput{
				SessionID: "sess-notif-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNotificationHandler().(*notificationHandler)
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("Notification hook should not set hookSpecificOutput")
			}

			// No config → HOI gate is false → no goroutine.
			testutil.WaitForAsync(t, h.waitGroup(), 100*time.Millisecond)
		})
	}
}

// TestNotification_DualGateMatrix verifies REQ-HAE-004 dual-gate semantics.
// Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state.
func TestNotification_DualGateMatrix(t *testing.T) {
	tests := []struct {
		name           string
		hoiEnabled     bool
		obsMasterOn    bool
		rt006Whitelist []string
	}{
		{
			name:           "Q1: HOI=false, obsMaster=false",
			hoiEnabled:     false,
			obsMasterOn:    false,
			rt006Whitelist: []string{"notification"},
		},
		{
			name:           "Q2: HOI=false, obsMaster=true",
			hoiEnabled:     false,
			obsMasterOn:    true,
			rt006Whitelist: []string{"notification"},
		},
		{
			name:           "Q3: HOI=true, obsMaster=false",
			hoiEnabled:     true,
			obsMasterOn:    false,
			rt006Whitelist: []string{"notification"},
		},
		{
			name:           "Q4: HOI=true, obsMaster=true, whitelist present",
			hoiEnabled:     true,
			obsMasterOn:    true,
			rt006Whitelist: []string{"notification"},
		},
		{
			name:           "Q4-edge: HOI=true, obsMaster=true, whitelist empty",
			hoiEnabled:     true,
			obsMasterOn:    true,
			rt006Whitelist: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetObservabilityMasterForTesting()
			SetObservabilityMasterForTesting(tt.obsMasterOn)
			t.Cleanup(ResetObservabilityMasterForTesting)

			cfg := newNotificationCfg(tt.hoiEnabled, tt.rt006Whitelist)
			h := NewNotificationHandlerWithConfig(cfg).(*notificationHandler)

			out, err := h.Handle(context.Background(), &HookInput{Title: "x"})
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if out == nil {
				t.Fatal("nil output")
			}
			if out.HookSpecificOutput != nil {
				t.Error("Notification hook should not set hookSpecificOutput")
			}

			testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)
		})
	}
}

// TestNotification_ObservabilityDisabled_NoGoroutine verifies the bonus
// assertion in AC-HAE-005: zero-overhead path when obsMaster is false.
func TestNotification_ObservabilityDisabled_NoGoroutine(t *testing.T) {
	ResetObservabilityMasterForTesting()
	SetObservabilityMasterForTesting(false)
	t.Cleanup(ResetObservabilityMasterForTesting)

	cfg := newNotificationCfg(true, []string{"notification"})
	h := NewNotificationHandlerWithConfig(cfg).(*notificationHandler)

	out, err := h.Handle(context.Background(), &HookInput{Title: "no-goroutine"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("nil output")
	}

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

// BenchmarkNotification_AsyncReturn measures p95 latency under 10-concurrent
// load with all gates passing. AC-HAE-005 requires p95 ≤ 100ms.
func BenchmarkNotification_AsyncReturn(b *testing.B) {
	ResetObservabilityMasterForTesting()
	SetObservabilityMasterForTesting(true)
	b.Cleanup(ResetObservabilityMasterForTesting)

	cfg := newNotificationCfg(true, []string{"notification"})

	const concurrency = 10
	durations := make([]time.Duration, 0, b.N*concurrency)
	var mu sync.Mutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := NewNotificationHandlerWithConfig(cfg).(*notificationHandler)
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				input := &HookInput{
					SessionID: "bench",
					Title:     "t-bench",
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

	p95Ms := notifPercentileMillis(durations, 0.95)
	b.ReportMetric(p95Ms, "p95-ms")
	if p95Ms > 100 {
		b.Errorf("AC-HAE-005 violation: p95 = %.2f ms, want ≤ 100 ms", p95Ms)
	}
}

func notifPercentileMillis(durations []time.Duration, p float64) float64 {
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
