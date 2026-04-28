package session

import (
	"testing"
	"time"
)

func TestNewBlockerReport(t *testing.T) {
	provenance := ProvenanceTag{
		Source: "session",
		Origin: "cli",
		Loaded: time.Now(),
	}

	report := NewBlockerReport(
		"missing_input",
		"SPEC ID is required",
		"provide_spec_id",
		provenance,
	)

	if report.Kind != "missing_input" {
		t.Errorf("Kind = %v, want missing_input", report.Kind)
	}
	if report.Message != "SPEC ID is required" {
		t.Errorf("Message = %v, want SPEC ID is required", report.Message)
	}
	if report.RequestedAction != "provide_spec_id" {
		t.Errorf("RequestedAction = %v, want provide_spec_id", report.RequestedAction)
	}
	if report.Resolved {
		t.Errorf("Resolved = %v, want false", report.Resolved)
	}
	if report.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestBlockerReportResolve(t *testing.T) {
	report := &BlockerReport{
		Kind:            "error",
		Message:         "Test failed",
		RequestedAction: "fix_test",
		Resolved:        false,
	}

	report.Resolve("fixed the test")

	if !report.Resolved {
		t.Errorf("Resolved = %v, want true", report.Resolved)
	}
	if report.Resolution != "fixed the test" {
		t.Errorf("Resolution = %v, want fixed the test", report.Resolution)
	}
}

func TestBlockerReportContext(t *testing.T) {
	report := &BlockerReport{
		Kind:    "quality_gate",
		Message: "Lint errors found",
		Context: map[string]any{
			"file":     "main.go",
			"line":     42,
			"errors":   3,
			"warnings": 1,
		},
	}

	if report.Context == nil {
		t.Fatal("Context should not be nil")
	}
	if report.Context["file"] != "main.go" {
		t.Errorf("Context[\"file\"] = %v, want main.go", report.Context["file"])
	}
	if report.Context["line"] != 42 {
		t.Errorf("Context[\"line\"] = %v, want 42", report.Context["line"])
	}
}
