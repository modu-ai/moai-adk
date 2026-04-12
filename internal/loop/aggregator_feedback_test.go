package loop

import (
	"context"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// mockDiagnosticsAggregator implements DiagnosticsAggregator for tests.
type mockDiagnosticsAggregator struct {
	diags map[string][]lsp.Diagnostic
	err   error
	calls []string
}

func (m *mockDiagnosticsAggregator) GetDiagnostics(ctx context.Context, path string) ([]lsp.Diagnostic, error) {
	m.calls = append(m.calls, path)
	if m.err != nil {
		return nil, m.err
	}
	if diags, ok := m.diags[path]; ok {
		return diags, nil
	}
	return []lsp.Diagnostic{}, nil
}

// TestGoFeedbackGenerator_WithAggregator_PopulatesLSPDiagnostics verifies REQ-LL-002:
// GoFeedbackGenerator with an Aggregator populates fb.LSPDiagnostics from Go source files.
func TestGoFeedbackGenerator_WithAggregator_PopulatesLSPDiagnostics(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	expectedDiags := []lsp.Diagnostic{
		{Severity: lsp.SeverityError, Source: "compiler", Message: "undeclared name: foo"},
		{Severity: lsp.SeverityWarning, Source: "staticcheck", Message: "deprecated use"},
	}

	agg := &mockDiagnosticsAggregator{
		diags: map[string][]lsp.Diagnostic{
			projectRoot: expectedDiags,
		},
	}

	gen := NewGoFeedbackGeneratorWithAggregator(projectRoot, agg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil Feedback")
	}

	if len(fb.LSPDiagnostics) != len(expectedDiags) {
		t.Errorf("LSPDiagnostics count = %d, want %d", len(fb.LSPDiagnostics), len(expectedDiags))
	}

	for i, d := range fb.LSPDiagnostics {
		if d.Severity != expectedDiags[i].Severity {
			t.Errorf("[%d] Severity = %d, want %d", i, d.Severity, expectedDiags[i].Severity)
		}
		if d.Source != expectedDiags[i].Source {
			t.Errorf("[%d] Source = %q, want %q", i, d.Source, expectedDiags[i].Source)
		}
	}
}

// TestGoFeedbackGenerator_WithAggregator_ErrorIsSilent verifies that aggregator
// errors do not propagate as Collect() errors (graceful degradation).
func TestGoFeedbackGenerator_WithAggregator_ErrorIsSilent(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	agg := &mockDiagnosticsAggregator{err: context.DeadlineExceeded}

	gen := NewGoFeedbackGeneratorWithAggregator(projectRoot, agg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() must not propagate aggregator error, got %v", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil Feedback")
	}
	if len(fb.LSPDiagnostics) != 0 {
		t.Errorf("on aggregator error, LSPDiagnostics must be empty, got %v", fb.LSPDiagnostics)
	}
}

// TestGoFeedbackGenerator_WithAggregator_NilAggregator verifies that nil aggregator
// falls back to go-toolchain-only feedback (REQ-LL-008).
func TestGoFeedbackGenerator_WithAggregator_NilAggregator(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithAggregator(projectRoot, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil Feedback")
	}
	// nil aggregator: LSPDiagnostics stays nil.
	if fb.LSPDiagnostics != nil {
		t.Errorf("nil aggregator: LSPDiagnostics = %v, want nil", fb.LSPDiagnostics)
	}
}

// TestDiagnosticsAggregator_Interface verifies DiagnosticsAggregator interface
// is declared in the loop package.
func TestDiagnosticsAggregator_Interface(t *testing.T) {
	t.Parallel()

	// Compile-time check: mockDiagnosticsAggregator satisfies DiagnosticsAggregator.
	var _ DiagnosticsAggregator = (*mockDiagnosticsAggregator)(nil)
}
