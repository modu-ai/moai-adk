package loop

import (
	"context"
	"errors"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// TestParseIntFromString covers parseIntFromString / mustParseInt / mustParseIntErr.
func TestParseIntFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"42", 42, false},
		{"100", 100, false},
		{"", 0, false}, // empty string: no digit chars, loop skips, returns 0
		{"abc", 0, true},
		{"1a", 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := parseIntFromString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseIntFromString(%q) = %d, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseIntFromString(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseIntFromString(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestMustParseInt covers mustParseInt for valid inputs.
func TestMustParseInt(t *testing.T) {
	t.Parallel()

	if got := mustParseInt("7"); got != 7 {
		t.Errorf("mustParseInt(\"7\") = %d, want 7", got)
	}
	if got := mustParseInt("abc"); got != 0 {
		t.Errorf("mustParseInt(\"abc\") = %d, want 0 (error case)", got)
	}
}

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
