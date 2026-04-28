package loop

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// mockBridge is a test double for gopls.Bridge.
// The result of GetDiagnostics calls can be preconfigured.
type mockBridge struct {
	diagnostics []gopls.Diagnostic
	err         error
	called      bool
}

func (m *mockBridge) GetDiagnostics(ctx context.Context, path string) ([]gopls.Diagnostic, error) {
	m.called = true
	return m.diagnostics, m.err
}

// TestGoFeedbackGenerator_NilBridge verifies that the existing behavior is preserved
// when bridge is nil (backward compatibility).
func TestGoFeedbackGenerator_NilBridge(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	// Create with bridge=nil
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, nil)

	if gen == nil {
		t.Fatal("NewGoFeedbackGeneratorWithBridge() must not return nil")
	}
}

// TestGoFeedbackGenerator_NilBridgeCollect verifies that Collect() leaves the Diagnostics
// field empty and returns normally when bridge is nil.
func TestGoFeedbackGenerator_NilBridgeCollect(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v, expected: nil", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil Feedback")
	}
	// Diagnostics must be nil when bridge is nil.
	if fb.Diagnostics != nil {
		t.Errorf("Diagnostics = %v, expected: nil (because bridge is nil)", fb.Diagnostics)
	}
}

// TestGoFeedbackGenerator_WithBridge_PopulatesDiagnostics verifies that
// GetDiagnostics results are populated into Feedback.Diagnostics when bridge is present.
func TestGoFeedbackGenerator_WithBridge_PopulatesDiagnostics(t *testing.T) {
	t.Parallel()

	expectedDiags := []gopls.Diagnostic{
		{
			Severity: gopls.SeverityError,
			Message:  "undeclared name: foo",
			Source:   "compiler",
		},
		{
			Severity: gopls.SeverityWarning,
			Message:  "unused variable: bar",
			Source:   "staticcheck",
		},
	}

	mock := &mockBridge{diagnostics: expectedDiags}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v, expected: nil", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil Feedback")
	}

	if !mock.called {
		t.Error("bridge.GetDiagnostics() was not called")
	}

	if len(fb.Diagnostics) != len(expectedDiags) {
		t.Errorf("Diagnostics count = %d, expected %d", len(fb.Diagnostics), len(expectedDiags))
	}

	for i, d := range fb.Diagnostics {
		if d.Severity != expectedDiags[i].Severity {
			t.Errorf("[%d] Severity = %d, expected %d", i, d.Severity, expectedDiags[i].Severity)
		}
		if d.Message != expectedDiags[i].Message {
			t.Errorf("[%d] Message = %q, expected %q", i, d.Message, expectedDiags[i].Message)
		}
	}
}

// TestGoFeedbackGenerator_WithBridge_ErrorIsSilent verifies that Collect() returns
// empty Diagnostics without error even when bridge.GetDiagnostics returns an error.
func TestGoFeedbackGenerator_WithBridge_ErrorIsSilent(t *testing.T) {
	t.Parallel()

	mock := &mockBridge{err: context.DeadlineExceeded}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() must not propagate bridge errors. error = %v", err)
	}
	if fb == nil {
		t.Fatal("Collect() returned nil Feedback")
	}
	// On error, Diagnostics must be nil (not a partial result).
	if len(fb.Diagnostics) != 0 {
		t.Errorf("Diagnostics must be empty on bridge error, got %v", fb.Diagnostics)
	}
}

// TestGoFeedbackGenerator_BackwardCompat verifies that NewGoFeedbackGenerator (original signature)
// still works (backward compatibility).
func TestGoFeedbackGenerator_BackwardCompat(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	// The original constructor must still compile.
	gen := NewGoFeedbackGenerator(projectRoot)
	if gen == nil {
		t.Fatal("NewGoFeedbackGenerator() must not return nil")
	}
}

// TestGoFeedbackGenerator_WithBridge_EmptyDiagnostics verifies that Diagnostics is empty
// when bridge returns an empty slice.
func TestGoFeedbackGenerator_WithBridge_EmptyDiagnostics(t *testing.T) {
	t.Parallel()

	mock := &mockBridge{diagnostics: []gopls.Diagnostic{}}
	projectRoot := t.TempDir()
	gen := NewGoFeedbackGeneratorWithBridge(projectRoot, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fb, err := gen.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect() error = %v, expected: nil", err)
	}
	if len(fb.Diagnostics) != 0 {
		t.Errorf("Diagnostics = %v, expected: empty slice", fb.Diagnostics)
	}
}
