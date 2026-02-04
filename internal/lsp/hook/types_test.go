package hook

import (
	"testing"
	"time"
)

func TestDiagnosticSeverityConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity DiagnosticSeverity
		expected string
	}{
		{"error severity", SeverityError, "error"},
		{"warning severity", SeverityWarning, "warning"},
		{"information severity", SeverityInformation, "information"},
		{"hint severity", SeverityHint, "hint"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.severity) != tt.expected {
				t.Errorf("got %q, want %q", tt.severity, tt.expected)
			}
		})
	}
}

func TestSeverityCountsTotal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		counts SeverityCounts
		want   int
	}{
		{
			name:   "all zeros",
			counts: SeverityCounts{},
			want:   0,
		},
		{
			name:   "only errors",
			counts: SeverityCounts{Errors: 5},
			want:   5,
		},
		{
			name:   "mixed counts",
			counts: SeverityCounts{Errors: 2, Warnings: 3, Information: 1, Hints: 4},
			want:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.counts.Total(); got != tt.want {
				t.Errorf("Total() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPositionStruct(t *testing.T) {
	t.Parallel()

	pos := Position{Line: 10, Character: 5}

	if pos.Line != 10 {
		t.Errorf("Line = %d, want 10", pos.Line)
	}
	if pos.Character != 5 {
		t.Errorf("Character = %d, want 5", pos.Character)
	}
}

func TestRangeStruct(t *testing.T) {
	t.Parallel()

	r := Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: 0, Character: 10},
	}

	if r.Start.Line != 0 {
		t.Errorf("Start.Line = %d, want 0", r.Start.Line)
	}
	if r.End.Character != 10 {
		t.Errorf("End.Character = %d, want 10", r.End.Character)
	}
}

func TestDiagnosticStruct(t *testing.T) {
	t.Parallel()

	diag := Diagnostic{
		Range: Range{
			Start: Position{Line: 5, Character: 0},
			End:   Position{Line: 5, Character: 20},
		},
		Severity: SeverityError,
		Code:     "E001",
		Source:   "gopls",
		Message:  "undeclared name: foo",
	}

	if diag.Severity != SeverityError {
		t.Errorf("Severity = %v, want %v", diag.Severity, SeverityError)
	}
	if diag.Code != "E001" {
		t.Errorf("Code = %q, want %q", diag.Code, "E001")
	}
	if diag.Source != "gopls" {
		t.Errorf("Source = %q, want %q", diag.Source, "gopls")
	}
	if diag.Message != "undeclared name: foo" {
		t.Errorf("Message = %q, want %q", diag.Message, "undeclared name: foo")
	}
}

func TestRegressionReportStruct(t *testing.T) {
	t.Parallel()

	report := RegressionReport{
		HasRegression:  true,
		HasImprovement: false,
		NewErrors:      3,
		FixedErrors:    0,
		NewWarnings:    1,
		FixedWarnings:  2,
	}

	if !report.HasRegression {
		t.Error("HasRegression should be true")
	}
	if report.HasImprovement {
		t.Error("HasImprovement should be false")
	}
	if report.NewErrors != 3 {
		t.Errorf("NewErrors = %d, want 3", report.NewErrors)
	}
}

func TestQualityGateStruct(t *testing.T) {
	t.Parallel()

	gate := QualityGate{
		MaxErrors:      0,
		MaxWarnings:    10,
		BlockOnError:   true,
		BlockOnWarning: false,
	}

	if gate.MaxErrors != 0 {
		t.Errorf("MaxErrors = %d, want 0", gate.MaxErrors)
	}
	if gate.MaxWarnings != 10 {
		t.Errorf("MaxWarnings = %d, want 10", gate.MaxWarnings)
	}
	if !gate.BlockOnError {
		t.Error("BlockOnError should be true")
	}
	if gate.BlockOnWarning {
		t.Error("BlockOnWarning should be false")
	}
}

func TestFileBaselineStruct(t *testing.T) {
	t.Parallel()

	now := time.Now()
	baseline := FileBaseline{
		Path:        "/path/to/file.go",
		Hash:        "abc123",
		Diagnostics: []Diagnostic{},
		UpdatedAt:   now,
	}

	if baseline.Path != "/path/to/file.go" {
		t.Errorf("Path = %q, want %q", baseline.Path, "/path/to/file.go")
	}
	if baseline.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", baseline.Hash, "abc123")
	}
}

func TestDiagnosticsBaselineStruct(t *testing.T) {
	t.Parallel()

	baseline := DiagnosticsBaseline{
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
		Files:     make(map[string]FileBaseline),
	}

	if baseline.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", baseline.Version, "1.0.0")
	}
	if baseline.Files == nil {
		t.Error("Files should not be nil")
	}
}

func TestSessionStatsStruct(t *testing.T) {
	t.Parallel()

	stats := SessionStats{
		TotalErrors:      5,
		TotalWarnings:    10,
		TotalInformation: 2,
		TotalHints:       3,
		FilesAnalyzed:    4,
		StartedAt:        time.Now(),
	}

	if stats.TotalErrors != 5 {
		t.Errorf("TotalErrors = %d, want 5", stats.TotalErrors)
	}
	if stats.FilesAnalyzed != 4 {
		t.Errorf("FilesAnalyzed = %d, want 4", stats.FilesAnalyzed)
	}
}

func TestErrDiagnosticsUnavailable(t *testing.T) {
	t.Parallel()

	t.Run("without reason", func(t *testing.T) {
		t.Parallel()
		err := &ErrDiagnosticsUnavailable{Language: "python"}
		expected := "diagnostics unavailable for python"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("with reason", func(t *testing.T) {
		t.Parallel()
		err := &ErrDiagnosticsUnavailable{
			Language: "python",
			Reason:   "ruff not installed",
		}
		expected := "diagnostics unavailable for python: ruff not installed"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})
}

func TestErrBaselineNotFound(t *testing.T) {
	t.Parallel()

	err := &ErrBaselineNotFound{FilePath: "/path/to/file.go"}
	expected := "baseline not found for /path/to/file.go"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
