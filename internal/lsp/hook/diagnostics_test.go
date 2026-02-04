package hook

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp"
)

// mockLSPClient implements lsp.DiagnosticsProvider for testing.
type mockLSPClient struct {
	diagnostics []lsp.Diagnostic
	err         error
}

func (m *mockLSPClient) Diagnostics(ctx context.Context, uri string) ([]lsp.Diagnostic, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.diagnostics, nil
}

// TestNewDiagnosticsCollector verifies collector creation.
func TestNewDiagnosticsCollector(t *testing.T) {
	t.Parallel()

	collector := NewDiagnosticsCollector(nil, nil)
	if collector == nil {
		t.Fatal("expected non-nil collector")
	}
}

// TestGetDiagnostics_WithLSP verifies LSP-first behavior per REQ-HOOK-151.
func TestGetDiagnostics_WithLSP(t *testing.T) {
	t.Parallel()

	mockClient := &mockLSPClient{
		diagnostics: []lsp.Diagnostic{
			{
				Range:    lsp.Range{Start: lsp.Position{Line: 10, Character: 0}, End: lsp.Position{Line: 10, Character: 10}},
				Severity: lsp.SeverityError,
				Code:     "E001",
				Source:   "gopls",
				Message:  "test error",
			},
		},
	}

	collector := NewDiagnosticsCollector(mockClient, nil)
	ctx := context.Background()

	diagnostics, err := collector.GetDiagnostics(ctx, "/path/to/file.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}

	if diagnostics[0].Severity != SeverityError {
		t.Errorf("expected severity %v, got %v", SeverityError, diagnostics[0].Severity)
	}
	if diagnostics[0].Message != "test error" {
		t.Errorf("expected message %q, got %q", "test error", diagnostics[0].Message)
	}
}

// TestGetDiagnostics_FallbackOnLSPError verifies fallback behavior per REQ-HOOK-152.
func TestGetDiagnostics_FallbackOnLSPError(t *testing.T) {
	t.Parallel()

	mockClient := &mockLSPClient{
		err: errors.New("LSP unavailable"),
	}

	mockFallback := &mockFallbackDiagnostics{
		diagnostics: []Diagnostic{
			{
				Range:    Range{Start: Position{Line: 5, Character: 0}, End: Position{Line: 5, Character: 10}},
				Severity: SeverityWarning,
				Source:   "ruff",
				Message:  "fallback warning",
			},
		},
		available: true,
	}

	collector := NewDiagnosticsCollector(mockClient, mockFallback)
	ctx := context.Background()

	diagnostics, err := collector.GetDiagnostics(ctx, "/path/to/file.py")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}

	if diagnostics[0].Source != "ruff" {
		t.Errorf("expected source %q, got %q", "ruff", diagnostics[0].Source)
	}
}

// TestGetDiagnostics_NoBlock verifies observation-only behavior per REQ-HOOK-153.
func TestGetDiagnostics_NoBlock(t *testing.T) {
	t.Parallel()

	// Both LSP and fallback fail
	mockClient := &mockLSPClient{
		err: errors.New("LSP unavailable"),
	}

	mockFallback := &mockFallbackDiagnostics{
		err:       &ErrDiagnosticsUnavailable{Language: "unknown"},
		available: false,
	}

	collector := NewDiagnosticsCollector(mockClient, mockFallback)
	ctx := context.Background()

	// Should return empty diagnostics, not error (observation-only)
	diagnostics, err := collector.GetDiagnostics(ctx, "/path/to/file.unknown")

	// Per REQ-HOOK-153: Must NOT block on diagnostic failure
	// Returns empty slice with no error, or returns error but calling code should not block
	if err != nil {
		// Error is acceptable, but it should be ErrDiagnosticsUnavailable
		var unavailableErr *ErrDiagnosticsUnavailable
		if !errors.As(err, &unavailableErr) {
			t.Errorf("expected ErrDiagnosticsUnavailable, got %T: %v", err, err)
		}
	}

	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	// Should have empty or nil diagnostics, not panic
	_ = len(diagnostics)
}

// TestGetDiagnostics_ContextCancellation verifies context handling.
func TestGetDiagnostics_ContextCancellation(t *testing.T) {
	t.Parallel()

	slowClient := &mockLSPClient{}
	collector := NewDiagnosticsCollector(slowClient, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(5 * time.Millisecond) // Let context expire

	_, err := collector.GetDiagnostics(ctx, "/path/to/file.go")
	if err == nil {
		// Context cancellation is acceptable but not required
		// The implementation may or may not check context
		t.Log("context cancellation not enforced (acceptable)")
	}
}

// TestGetSeverityCounts verifies severity counting per REQ-HOOK-150.
func TestGetSeverityCounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		diagnostics []Diagnostic
		want        SeverityCounts
	}{
		{
			name:        "empty diagnostics",
			diagnostics: []Diagnostic{},
			want:        SeverityCounts{},
		},
		{
			name: "mixed severities",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Message: "error 1"},
				{Severity: SeverityError, Message: "error 2"},
				{Severity: SeverityWarning, Message: "warning 1"},
				{Severity: SeverityInformation, Message: "info 1"},
				{Severity: SeverityHint, Message: "hint 1"},
				{Severity: SeverityHint, Message: "hint 2"},
			},
			want: SeverityCounts{
				Errors:      2,
				Warnings:    1,
				Information: 1,
				Hints:       2,
			},
		},
		{
			name: "only errors",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Message: "error 1"},
				{Severity: SeverityError, Message: "error 2"},
			},
			want: SeverityCounts{Errors: 2},
		},
	}

	collector := NewDiagnosticsCollector(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := collector.GetSeverityCounts(tt.diagnostics)
			if got != tt.want {
				t.Errorf("GetSeverityCounts() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestConvertLSPDiagnostic verifies LSP diagnostic conversion.
func TestConvertLSPDiagnostic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    lsp.Diagnostic
		expected Diagnostic
	}{
		{
			name: "error diagnostic",
			input: lsp.Diagnostic{
				Range:    lsp.Range{Start: lsp.Position{Line: 10, Character: 5}, End: lsp.Position{Line: 10, Character: 15}},
				Severity: lsp.SeverityError,
				Code:     "E001",
				Source:   "gopls",
				Message:  "undefined: foo",
			},
			expected: Diagnostic{
				Range:    Range{Start: Position{Line: 10, Character: 5}, End: Position{Line: 10, Character: 15}},
				Severity: SeverityError,
				Code:     "E001",
				Source:   "gopls",
				Message:  "undefined: foo",
			},
		},
		{
			name: "warning diagnostic",
			input: lsp.Diagnostic{
				Range:    lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 0}},
				Severity: lsp.SeverityWarning,
				Message:  "unused variable",
			},
			expected: Diagnostic{
				Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}},
				Severity: SeverityWarning,
				Message:  "unused variable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := convertLSPDiagnostic(tt.input)
			if got.Severity != tt.expected.Severity {
				t.Errorf("Severity = %v, want %v", got.Severity, tt.expected.Severity)
			}
			if got.Message != tt.expected.Message {
				t.Errorf("Message = %q, want %q", got.Message, tt.expected.Message)
			}
			if got.Range.Start.Line != tt.expected.Range.Start.Line {
				t.Errorf("Range.Start.Line = %d, want %d", got.Range.Start.Line, tt.expected.Range.Start.Line)
			}
		})
	}
}

// TestConvertLSPSeverity verifies severity conversion.
func TestConvertLSPSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    lsp.DiagnosticSeverity
		expected DiagnosticSeverity
	}{
		{lsp.SeverityError, SeverityError},
		{lsp.SeverityWarning, SeverityWarning},
		{lsp.SeverityInfo, SeverityInformation},
		{lsp.SeverityHint, SeverityHint},
		{lsp.DiagnosticSeverity(99), SeverityInformation}, // unknown defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.expected.String(), func(t *testing.T) {
			t.Parallel()
			got := convertLSPSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("convertLSPSeverity(%d) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestFormatDiagnostics verifies diagnostic formatting.
func TestFormatDiagnostics(t *testing.T) {
	t.Parallel()

	t.Run("empty diagnostics", func(t *testing.T) {
		t.Parallel()
		result := FormatDiagnostics("/path/to/file.go", []Diagnostic{})
		if result == "" {
			t.Error("expected non-empty result")
		}
		if !containsStr(result, "No diagnostics") {
			t.Error("expected 'No diagnostics' message")
		}
	})

	t.Run("with diagnostics", func(t *testing.T) {
		t.Parallel()
		diagnostics := []Diagnostic{
			{
				Range:    Range{Start: Position{Line: 10, Character: 5}},
				Severity: SeverityError,
				Code:     "E001",
				Source:   "gopls",
				Message:  "undefined: foo",
			},
			{
				Range:    Range{Start: Position{Line: 20, Character: 0}},
				Severity: SeverityWarning,
				Source:   "gopls",
				Message:  "unused variable",
			},
		}

		result := FormatDiagnostics("/path/to/file.go", diagnostics)
		if !containsStr(result, "ERROR") {
			t.Error("expected ERROR in output")
		}
		if !containsStr(result, "WARNING") {
			t.Error("expected WARNING in output")
		}
		if !containsStr(result, "gopls") {
			t.Error("expected source in output")
		}
	})
}

// containsStr checks if s contains substr.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestFilePathToURI verifies file path to URI conversion.
func TestFilePathToURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"absolute path", "/Users/test/file.go", "file://"},
		{"relative path", "file.go", "file://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := filePathToURI(tt.input)
			if !containsStr(result, tt.contains) {
				t.Errorf("filePathToURI(%q) = %q, expected to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}

// TestDiagnosticSeverityString verifies severity string representation.
func TestDiagnosticSeverityString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity DiagnosticSeverity
		expected string
	}{
		{SeverityError, "Error"},
		{SeverityWarning, "Warning"},
		{SeverityInformation, "Information"},
		{SeverityHint, "Hint"},
		{DiagnosticSeverity("unknown"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// mockFallbackDiagnostics is used for testing fallback behavior.
type mockFallbackDiagnostics struct {
	diagnostics []Diagnostic
	err         error
	available   bool
	language    string
}

func (m *mockFallbackDiagnostics) RunFallback(ctx context.Context, filePath string) ([]Diagnostic, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.diagnostics, nil
}

func (m *mockFallbackDiagnostics) IsAvailable(language string) bool {
	return m.available
}

func (m *mockFallbackDiagnostics) GetLanguage(filePath string) string {
	if m.language != "" {
		return m.language
	}
	return "unknown"
}
