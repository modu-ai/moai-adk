package loop

import (
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// TestIsStagnantWithDiagnostics_DiagCountIncreasing verifies REQ-LL-010:
// when LSPDiagnostics count trends upward, stagnation is signaled.
func TestIsStagnantWithDiagnostics_DiagCountIncreasing(t *testing.T) {
	t.Parallel()

	prev := &Feedback{
		TestsFailed:    0,
		LintErrors:     0,
		Coverage:       85.0,
		LSPDiagnostics: []lsp.Diagnostic{{Severity: lsp.SeverityError}},
	}
	curr := &Feedback{
		TestsFailed: 0,
		LintErrors:  0,
		Coverage:    85.0,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityError},
			{Severity: lsp.SeverityError},
			{Severity: lsp.SeverityError},
		},
	}

	// Increasing diagnostic count is a stagnation signal.
	if !IsStagnantWithDiagnostics(prev, curr) {
		t.Error("increasing LSPDiagnostics count must signal stagnation")
	}
}

// TestIsStagnantWithDiagnostics_DiagCountDecreasing verifies REQ-LL-010:
// when LSPDiagnostics count decreases, it is NOT stagnation (improvement).
func TestIsStagnantWithDiagnostics_DiagCountDecreasing(t *testing.T) {
	t.Parallel()

	prev := &Feedback{
		TestsFailed: 0,
		LintErrors:  0,
		Coverage:    85.0,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityError},
			{Severity: lsp.SeverityError},
		},
	}
	curr := &Feedback{
		TestsFailed:    0,
		LintErrors:     0,
		Coverage:       85.0,
		LSPDiagnostics: []lsp.Diagnostic{{Severity: lsp.SeverityWarning}},
	}

	if IsStagnantWithDiagnostics(prev, curr) {
		t.Error("decreasing LSPDiagnostics count must NOT signal stagnation")
	}
}

// TestIsStagnantWithDiagnostics_NoDiagnostics verifies REQ-LL-010:
// falls back to IsStagnant behavior when no LSPDiagnostics present.
func TestIsStagnantWithDiagnostics_NoDiagnostics(t *testing.T) {
	t.Parallel()

	prev := &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0}
	curr := &Feedback{TestsFailed: 2, LintErrors: 1, Coverage: 80.0}

	// Same integer metrics with no diagnostics: stagnant.
	if !IsStagnantWithDiagnostics(prev, curr) {
		t.Error("identical integer metrics with no diagnostics must signal stagnation")
	}
}

// TestIsStagnantWithDiagnostics_NilPointers verifies nil handling.
func TestIsStagnantWithDiagnostics_NilPointers(t *testing.T) {
	t.Parallel()

	if IsStagnantWithDiagnostics(nil, &Feedback{}) {
		t.Error("nil prev must return false")
	}
	if IsStagnantWithDiagnostics(&Feedback{}, nil) {
		t.Error("nil curr must return false")
	}
}

// TestIsStagnantWithDiagnostics_SameCount verifies REQ-LL-010:
// identical diagnostic counts do not additionally indicate improvement.
func TestIsStagnantWithDiagnostics_SameCount(t *testing.T) {
	t.Parallel()

	diag := lsp.Diagnostic{Severity: lsp.SeverityError}
	prev := &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 85.0, LSPDiagnostics: []lsp.Diagnostic{diag}}
	curr := &Feedback{TestsFailed: 0, LintErrors: 0, Coverage: 85.0, LSPDiagnostics: []lsp.Diagnostic{diag}}

	// Same count + same integer metrics: stagnant.
	if !IsStagnantWithDiagnostics(prev, curr) {
		t.Error("identical diagnostic counts with identical integer metrics must signal stagnation")
	}
}
