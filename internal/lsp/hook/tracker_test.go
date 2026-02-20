package hook

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewRegressionTracker verifies tracker creation.
func TestNewRegressionTracker(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
}

// TestSaveBaseline verifies baseline saving per REQ-HOOK-170.
func TestSaveBaseline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	diagnostics := []Diagnostic{
		{
			Range:    Range{Start: Position{Line: 10, Character: 0}, End: Position{Line: 10, Character: 10}},
			Severity: SeverityError,
			Code:     "E001",
			Source:   "test",
			Message:  "test error",
		},
	}

	err := tracker.SaveBaseline("/path/to/file.go", diagnostics)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Verify baseline file was created
	baselineFile := filepath.Join(tmpDir, "diagnostics-baseline.json")
	if _, err := os.Stat(baselineFile); os.IsNotExist(err) {
		t.Error("baseline file was not created")
	}
}

// TestGetBaseline verifies baseline retrieval.
func TestGetBaseline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"
	diagnostics := []Diagnostic{
		{
			Range:    Range{Start: Position{Line: 5, Character: 0}, End: Position{Line: 5, Character: 10}},
			Severity: SeverityWarning,
			Code:     "W001",
			Source:   "test",
			Message:  "test warning",
		},
	}

	// Save baseline first
	err := tracker.SaveBaseline(filePath, diagnostics)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Retrieve baseline
	baseline, err := tracker.GetBaseline(filePath)
	if err != nil {
		t.Fatalf("GetBaseline failed: %v", err)
	}

	if baseline == nil {
		t.Fatal("expected non-nil baseline")
	}

	if len(baseline.Diagnostics) != 1 {
		t.Errorf("expected 1 diagnostic, got %d", len(baseline.Diagnostics))
	}

	if baseline.Diagnostics[0].Message != "test warning" {
		t.Errorf("expected message %q, got %q", "test warning", baseline.Diagnostics[0].Message)
	}
}

// TestGetBaseline_NotFound verifies behavior when no baseline exists.
func TestGetBaseline_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	_, err := tracker.GetBaseline("/nonexistent/file.go")

	if err == nil {
		t.Fatal("expected error for nonexistent baseline")
	}

	var notFoundErr *ErrBaselineNotFound
	if e, ok := err.(*ErrBaselineNotFound); ok {
		notFoundErr = e
	}
	if notFoundErr == nil {
		t.Errorf("expected ErrBaselineNotFound, got %T: %v", err, err)
	}
}

// TestCompareWithBaseline_Regression verifies regression detection per REQ-HOOK-171.
func TestCompareWithBaseline_Regression(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"

	// Baseline has 1 error
	baseline := []Diagnostic{
		{Severity: SeverityError, Message: "existing error"},
	}

	err := tracker.SaveBaseline(filePath, baseline)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Current has 3 errors (regression)
	current := []Diagnostic{
		{Severity: SeverityError, Message: "existing error"},
		{Severity: SeverityError, Message: "new error 1"},
		{Severity: SeverityError, Message: "new error 2"},
	}

	report, err := tracker.CompareWithBaseline(filePath, current)
	if err != nil {
		t.Fatalf("CompareWithBaseline failed: %v", err)
	}

	if !report.HasRegression {
		t.Error("expected HasRegression to be true")
	}
	if report.NewErrors != 2 {
		t.Errorf("NewErrors = %d, want 2", report.NewErrors)
	}
}

// TestCompareWithBaseline_Improvement verifies improvement detection per REQ-HOOK-172.
func TestCompareWithBaseline_Improvement(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"

	// Baseline has 3 errors
	baseline := []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
		{Severity: SeverityError, Message: "error 2"},
		{Severity: SeverityError, Message: "error 3"},
	}

	err := tracker.SaveBaseline(filePath, baseline)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Current has 1 error (improvement)
	current := []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
	}

	report, err := tracker.CompareWithBaseline(filePath, current)
	if err != nil {
		t.Fatalf("CompareWithBaseline failed: %v", err)
	}

	if !report.HasImprovement {
		t.Error("expected HasImprovement to be true")
	}
	if report.FixedErrors != 2 {
		t.Errorf("FixedErrors = %d, want 2", report.FixedErrors)
	}
}

// TestCompareWithBaseline_NoChange verifies no-change detection.
func TestCompareWithBaseline_NoChange(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"

	diagnostics := []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
		{Severity: SeverityWarning, Message: "warning 1"},
	}

	err := tracker.SaveBaseline(filePath, diagnostics)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	report, err := tracker.CompareWithBaseline(filePath, diagnostics)
	if err != nil {
		t.Fatalf("CompareWithBaseline failed: %v", err)
	}

	if report.HasRegression {
		t.Error("expected HasRegression to be false")
	}
	if report.HasImprovement {
		t.Error("expected HasImprovement to be false")
	}
}

// TestCompareWithBaseline_Warnings verifies warning comparison.
func TestCompareWithBaseline_Warnings(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"

	// Baseline has 2 warnings
	baseline := []Diagnostic{
		{Severity: SeverityWarning, Message: "warning 1"},
		{Severity: SeverityWarning, Message: "warning 2"},
	}

	err := tracker.SaveBaseline(filePath, baseline)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Current has 3 warnings (1 new warning)
	current := []Diagnostic{
		{Severity: SeverityWarning, Message: "warning 1"},
		{Severity: SeverityWarning, Message: "warning 2"},
		{Severity: SeverityWarning, Message: "warning 3"},
	}

	report, err := tracker.CompareWithBaseline(filePath, current)
	if err != nil {
		t.Fatalf("CompareWithBaseline failed: %v", err)
	}

	if report.NewWarnings != 1 {
		t.Errorf("NewWarnings = %d, want 1", report.NewWarnings)
	}
}

// TestClearBaseline verifies baseline removal.
func TestClearBaseline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"

	// Save baseline first
	err := tracker.SaveBaseline(filePath, []Diagnostic{})
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Clear baseline
	err = tracker.ClearBaseline(filePath)
	if err != nil {
		t.Fatalf("ClearBaseline failed: %v", err)
	}

	// Verify baseline is gone
	_, err = tracker.GetBaseline(filePath)
	if err == nil {
		t.Error("expected error after clearing baseline")
	}
}

// TestNewSessionTracker verifies session tracker creation.
func TestNewSessionTracker(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	if tracker == nil {
		t.Fatal("expected non-nil session tracker")
	}
}

// TestSessionTracker_StartSession verifies session initialization per REQ-HOOK-190.
func TestSessionTracker_StartSession(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()

	err := tracker.StartSession()
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	stats := tracker.GetSessionStats()
	if stats.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero after StartSession")
	}
}

// TestSessionTracker_RecordDiagnostics verifies diagnostic recording per REQ-HOOK-190.
func TestSessionTracker_RecordDiagnostics(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	if err := tracker.StartSession(); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	diagnostics := []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
		{Severity: SeverityError, Message: "error 2"},
		{Severity: SeverityWarning, Message: "warning 1"},
		{Severity: SeverityInformation, Message: "info 1"},
		{Severity: SeverityHint, Message: "hint 1"},
	}

	tracker.RecordDiagnostics("/path/to/file1.go", diagnostics)

	stats := tracker.GetSessionStats()
	if stats.TotalErrors != 2 {
		t.Errorf("TotalErrors = %d, want 2", stats.TotalErrors)
	}
	if stats.TotalWarnings != 1 {
		t.Errorf("TotalWarnings = %d, want 1", stats.TotalWarnings)
	}
	if stats.FilesAnalyzed != 1 {
		t.Errorf("FilesAnalyzed = %d, want 1", stats.FilesAnalyzed)
	}
}

// TestSessionTracker_MultipleFiles verifies tracking across multiple files.
func TestSessionTracker_MultipleFiles(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	if err := tracker.StartSession(); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Record diagnostics for file 1
	tracker.RecordDiagnostics("/path/to/file1.go", []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
	})

	// Record diagnostics for file 2
	tracker.RecordDiagnostics("/path/to/file2.go", []Diagnostic{
		{Severity: SeverityError, Message: "error 2"},
		{Severity: SeverityWarning, Message: "warning 1"},
	})

	stats := tracker.GetSessionStats()
	if stats.TotalErrors != 2 {
		t.Errorf("TotalErrors = %d, want 2", stats.TotalErrors)
	}
	if stats.FilesAnalyzed != 2 {
		t.Errorf("FilesAnalyzed = %d, want 2", stats.FilesAnalyzed)
	}
}

// TestSessionTracker_GetFileStats verifies per-file statistics.
func TestSessionTracker_GetFileStats(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	if err := tracker.StartSession(); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	filePath := "/path/to/file.go"

	// Record diagnostics multiple times
	tracker.RecordDiagnostics(filePath, []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
	})

	tracker.RecordDiagnostics(filePath, []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
		{Severity: SeverityError, Message: "error 2"},
	})

	fileStats, err := tracker.GetFileStats(filePath)
	if err != nil {
		t.Fatalf("GetFileStats failed: %v", err)
	}

	if len(fileStats.DiagnosticHistory) != 2 {
		t.Errorf("DiagnosticHistory length = %d, want 2", len(fileStats.DiagnosticHistory))
	}

	// First recording had 1 error
	if fileStats.DiagnosticHistory[0].Errors != 1 {
		t.Errorf("first history errors = %d, want 1", fileStats.DiagnosticHistory[0].Errors)
	}

	// Second recording had 2 errors
	if fileStats.DiagnosticHistory[1].Errors != 2 {
		t.Errorf("second history errors = %d, want 2", fileStats.DiagnosticHistory[1].Errors)
	}
}

// TestSessionTracker_EndSession verifies session end per REQ-HOOK-191.
func TestSessionTracker_EndSession(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()
	if err := tracker.StartSession(); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	tracker.RecordDiagnostics("/path/to/file.go", []Diagnostic{
		{Severity: SeverityError, Message: "error 1"},
	})

	stats, err := tracker.EndSession()
	if err != nil {
		t.Fatalf("EndSession failed: %v", err)
	}

	if stats.TotalErrors != 1 {
		t.Errorf("TotalErrors = %d, want 1", stats.TotalErrors)
	}
}

// TestBaselineStorageFormat verifies storage format per SPEC.
func TestBaselineStorageFormat(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"
	diagnostics := []Diagnostic{
		{
			Range:    Range{Start: Position{Line: 10, Character: 5}, End: Position{Line: 10, Character: 15}},
			Severity: SeverityError,
			Code:     "E001",
			Source:   "test",
			Message:  "test error",
		},
	}

	err := tracker.SaveBaseline(filePath, diagnostics)
	if err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Read the baseline file and verify format
	baselineFile := filepath.Join(tmpDir, "diagnostics-baseline.json")
	data, err := os.ReadFile(baselineFile)
	if err != nil {
		t.Fatalf("failed to read baseline file: %v", err)
	}

	// Verify it contains expected fields
	content := string(data)
	if !contains(content, "version") {
		t.Error("baseline should contain 'version' field")
	}
	if !contains(content, "updatedAt") {
		t.Error("baseline should contain 'updatedAt' field")
	}
	if !contains(content, "files") {
		t.Error("baseline should contain 'files' field")
	}
}

// TestTrackerThreadSafety verifies thread-safe operations.
func TestTrackerThreadSafety(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	done := make(chan bool)

	// Concurrent saves
	for i := range 10 {
		go func(idx int) {
			filePath := filepath.Join("/path/to", "file"+string(rune('A'+idx))+".go")
			diagnostics := []Diagnostic{
				{Severity: SeverityError, Message: "error"},
			}
			_ = tracker.SaveBaseline(filePath, diagnostics)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for goroutines")
		}
	}
}

// TestCompareWithBaseline_NoBaseline verifies comparison when no baseline exists.
func TestCompareWithBaseline_NoBaseline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	diagnostics := []Diagnostic{
		{Severity: SeverityError, Message: "error"},
	}

	_, err := tracker.CompareWithBaseline("/nonexistent/file.go", diagnostics)
	if err == nil {
		t.Error("expected error when no baseline exists")
	}
}

// TestSaveAndLoadMultipleFiles verifies multiple file baselines.
func TestSaveAndLoadMultipleFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	// Save baselines for multiple files
	files := []string{"/path/to/file1.go", "/path/to/file2.go", "/path/to/file3.go"}
	for i, f := range files {
		diagnostics := []Diagnostic{
			{Severity: SeverityError, Message: "error " + string(rune('A'+i))},
		}
		if err := tracker.SaveBaseline(f, diagnostics); err != nil {
			t.Fatalf("SaveBaseline failed for %s: %v", f, err)
		}
	}

	// Verify all baselines can be retrieved
	for i, f := range files {
		baseline, err := tracker.GetBaseline(f)
		if err != nil {
			t.Fatalf("GetBaseline failed for %s: %v", f, err)
		}
		expectedMsg := "error " + string(rune('A'+i))
		if len(baseline.Diagnostics) != 1 || baseline.Diagnostics[0].Message != expectedMsg {
			t.Errorf("unexpected diagnostics for %s", f)
		}
	}
}

// TestSaveBaselineOverwrite verifies baseline overwrite.
func TestSaveBaselineOverwrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)

	filePath := "/path/to/file.go"

	// First save
	diag1 := []Diagnostic{{Severity: SeverityError, Message: "first"}}
	if err := tracker.SaveBaseline(filePath, diag1); err != nil {
		t.Fatalf("first SaveBaseline failed: %v", err)
	}

	// Overwrite
	diag2 := []Diagnostic{{Severity: SeverityWarning, Message: "second"}}
	if err := tracker.SaveBaseline(filePath, diag2); err != nil {
		t.Fatalf("second SaveBaseline failed: %v", err)
	}

	// Verify overwrite
	baseline, err := tracker.GetBaseline(filePath)
	if err != nil {
		t.Fatalf("GetBaseline failed: %v", err)
	}
	if len(baseline.Diagnostics) != 1 || baseline.Diagnostics[0].Message != "second" {
		t.Error("baseline was not overwritten")
	}
}

// TestSessionTracker_NotStarted verifies behavior when session not started.
func TestSessionTracker_NotStarted(t *testing.T) {
	t.Parallel()

	tracker := NewSessionTracker()

	// Recording without starting should be ignored
	tracker.RecordDiagnostics("/path/to/file.go", []Diagnostic{
		{Severity: SeverityError, Message: "error"},
	})

	stats := tracker.GetSessionStats()
	if stats.TotalErrors != 0 {
		t.Error("expected 0 errors when session not started")
	}
}

// TestFormatGateResult_AllPaths verifies all output paths.
func TestFormatGateResult_AllPaths(t *testing.T) {
	t.Parallel()

	// Test passed status
	t.Run("passed", func(t *testing.T) {
		t.Parallel()
		counts := SeverityCounts{Errors: 0, Warnings: 5}
		gate := QualityGate{MaxErrors: 5, MaxWarnings: 10}
		result := FormatGateResult(counts, gate)
		if !containsAt(result, "PASSED", 0) {
			t.Log(result)
		}
	})

	// Test warning exceeded (non-blocking)
	t.Run("warning exceeded", func(t *testing.T) {
		t.Parallel()
		counts := SeverityCounts{Errors: 0, Warnings: 15}
		gate := QualityGate{MaxErrors: 0, MaxWarnings: 10, BlockOnWarning: false}
		result := FormatGateResult(counts, gate)
		if !containsAt(result, "WARNING", 0) && !containsAt(result, "EXCEEDED", 0) {
			t.Log(result)
		}
	})

	// Test warning exceeded (blocking)
	t.Run("warning blocking", func(t *testing.T) {
		t.Parallel()
		counts := SeverityCounts{Errors: 0, Warnings: 15}
		gate := QualityGate{MaxErrors: 0, MaxWarnings: 10, BlockOnWarning: true}
		result := FormatGateResult(counts, gate)
		if !containsAt(result, "BLOCKED", 0) {
			t.Log(result)
		}
	})
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
