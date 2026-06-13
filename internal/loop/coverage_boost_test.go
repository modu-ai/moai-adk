package loop

import (
	"context"
	"errors"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// NOTE: TestParseIntFromString / TestMustParseInt moved to internal/measure/measure_test.go
// when the parsers + transitive helpers (parseIntFromString / mustParseInt / mustParseIntErr)
// were extracted to the internal/measure leaf package (SPEC-HARNESS-REGRESSION-GATE-001 M1/M2).
// The loop package now delegates parsing to internal/measure; the helpers no longer live here.

// TestFilterGoOnlyDiagnostics covers filterGoOnlyDiagnostics edge cases.
func TestFilterGoOnlyDiagnostics(t *testing.T) {
	t.Parallel()

	// Empty input returns nil.
	if got := filterGoOnlyDiagnostics(nil); got != nil {
		t.Errorf("filterGoOnlyDiagnostics(nil) = %v, want nil", got)
	}
	if got := filterGoOnlyDiagnostics([]lsp.Diagnostic{}); got != nil {
		t.Errorf("filterGoOnlyDiagnostics([]) = %v, want nil", got)
	}

	diags := []lsp.Diagnostic{
		{Severity: lsp.SeverityError, Message: "err"},
		{Severity: lsp.SeverityWarning, Message: "warn"},
	}
	got := filterGoOnlyDiagnostics(diags)
	if len(got) != len(diags) {
		t.Errorf("filterGoOnlyDiagnostics(%d diags) = %d, want %d", len(diags), len(got), len(diags))
	}
}

// TestGoFeedbackGenerator_LegacyBridge_StillWorks ensures gopls.Bridge path compiles
// and populates fb.Diagnostics (not fb.LSPDiagnostics).
func TestGoFeedbackGenerator_LegacyBridge_StillWorks(t *testing.T) {
	t.Parallel()

	expectedDiags := []gopls.Diagnostic{
		{Severity: gopls.SeverityError, Message: "legacy error"},
	}
	mock := &mockBridge{diagnostics: expectedDiags}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if !mock.called {
		t.Error("bridge.GetDiagnostics must be called")
	}
	if len(fb.Diagnostics) != 1 {
		t.Errorf("Diagnostics count = %d, want 1", len(fb.Diagnostics))
	}
	// Legacy bridge must NOT populate LSPDiagnostics.
	if len(fb.LSPDiagnostics) != 0 {
		t.Errorf("LSPDiagnostics must be empty when using legacy bridge, got %v", fb.LSPDiagnostics)
	}
}

// TestGoFeedbackGenerator_AggregatorWithError covers the error-is-silent path.
func TestGoFeedbackGenerator_AggregatorWithError_LSPEmpty(t *testing.T) {
	t.Parallel()

	agg := &mockDiagnosticsAggregator{err: errors.New("network error")}
	gen := NewGoFeedbackGeneratorWithAggregator(t.TempDir(), agg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() must not propagate aggregator error: %v", err)
	}
	if len(fb.LSPDiagnostics) != 0 {
		t.Errorf("on error, LSPDiagnostics must be empty, got %v", fb.LSPDiagnostics)
	}
}

// TestFeedbackChannel_CapacityEnforced verifies bounded channel capacity.
func TestFeedbackChannel_CapacityEnforced(t *testing.T) {
	t.Parallel()

	ch := NewFeedbackChannel(3)

	// Send 5 events to a channel of capacity 3.
	for i := 0; i < 5; i++ {
		ch.Send(Feedback{TestsFailed: i})
	}

	// After 5 sends to capacity-3, we should have exactly 3 events
	// (oldest 2 dropped, last 3 kept — or similar bounded behavior).
	count := ch.Len()
	if count > 3 {
		t.Errorf("channel exceeded capacity: len = %d, capacity = 3", count)
	}
}

// TestIsStagnantWithDiagnostics_BothHaveEmptyDiags covers path where both feedbacks
// have empty LSPDiagnostics slices (not nil).
func TestIsStagnantWithDiagnostics_BothHaveEmptyDiags(t *testing.T) {
	t.Parallel()

	prev := &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 90.0, LSPDiagnostics: []lsp.Diagnostic{}}
	curr := &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 90.0, LSPDiagnostics: []lsp.Diagnostic{}}

	// Same integer metrics + same (0) diagnostic count: stagnant.
	if !IsStagnantWithDiagnostics(prev, curr) {
		t.Error("empty LSPDiagnostics with identical metrics must signal stagnation")
	}
}
