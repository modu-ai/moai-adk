package cli

// Coverage tests for astgrep output functions and filterBySeverity.
// These were 0% or very low coverage.

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
	"github.com/spf13/cobra"
)

// makeTestCmd creates a minimal cobra.Command with a bytes.Buffer as output.
func makeTestCmd() (*cobra.Command, *bytes.Buffer) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	return cmd, &buf
}

// --- outputText ---

// TestOutputText_NoFindings prints "no findings" message.
func TestOutputText_NoFindings(t *testing.T) {
	t.Parallel()

	cmd, buf := makeTestCmd()
	outputText(cmd, nil)
	if buf.Len() == 0 {
		t.Error("expected output for empty findings")
	}
	// Should not panic.
}

// TestOutputText_WithFindings formats findings.
func TestOutputText_WithFindings(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{RuleID: "rule-1", Message: "test message", Severity: "error", File: "main.go", Line: 10},
		{RuleID: "rule-2", Message: "warning msg", Severity: "warning", File: "util.go", Line: 5, Note: "see docs"},
	}
	cmd, buf := makeTestCmd()
	outputText(cmd, findings)

	out := buf.String()
	if out == "" {
		t.Error("expected non-empty output")
	}
}

// TestOutputText_WithNote includes note in output.
func TestOutputText_WithNote(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{RuleID: "rule-1", Message: "msg", File: "file.go", Line: 1, Note: "important note"},
	}
	cmd, buf := makeTestCmd()
	outputText(cmd, findings)

	out := buf.String()
	if !strings.Contains(out, "important note") {
		t.Errorf("output should contain note text, got %q", out)
	}
}

// --- outputJSON ---

// TestOutputJSON_EmptyFindings outputs empty JSON array.
func TestOutputJSON_EmptyFindings(t *testing.T) {
	t.Parallel()

	cmd, buf := makeTestCmd()
	if err := outputJSON(cmd, nil); err != nil {
		t.Fatalf("error: %v", err)
	}

	var result []astgrep.Finding
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %v", result)
	}
}

// TestOutputJSON_WithFindings outputs valid JSON array.
func TestOutputJSON_WithFindings(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{RuleID: "r1", Message: "msg1", Severity: "error"},
		{RuleID: "r2", Message: "msg2", Severity: "warning"},
	}
	cmd, buf := makeTestCmd()
	if err := outputJSON(cmd, findings); err != nil {
		t.Fatalf("error: %v", err)
	}

	var result []astgrep.Finding
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 findings, got %d", len(result))
	}
}

// --- filterBySeverity ---

// TestFilterBySeverity_ErrorLevel keeps only errors.
func TestFilterBySeverity_ErrorLevel(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{Severity: "error"},
		{Severity: "warning"},
		{Severity: "info"},
	}

	result := filterBySeverity(findings, "error")
	if len(result) != 1 {
		t.Errorf("expected 1 error finding, got %d", len(result))
	}
	if result[0].Severity != "error" {
		t.Errorf("expected error severity, got %q", result[0].Severity)
	}
}

// TestFilterBySeverity_WarningLevel keeps errors and warnings.
func TestFilterBySeverity_WarningLevel(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{Severity: "error"},
		{Severity: "warning"},
		{Severity: "info"},
	}

	result := filterBySeverity(findings, "warning")
	if len(result) != 2 {
		t.Errorf("expected 2 findings (error+warning), got %d", len(result))
	}
}

// TestFilterBySeverity_InfoLevel keeps all findings.
func TestFilterBySeverity_InfoLevel(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{Severity: "error"},
		{Severity: "warning"},
		{Severity: "info"},
	}

	result := filterBySeverity(findings, "info")
	if len(result) != 3 {
		t.Errorf("expected all 3 findings, got %d", len(result))
	}
}

// TestFilterBySeverity_UnknownLevel keeps all (default).
func TestFilterBySeverity_UnknownLevel(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{Severity: "error"},
		{Severity: "warning"},
	}

	result := filterBySeverity(findings, "unknown")
	if len(result) != 2 {
		t.Errorf("expected all 2 findings for unknown level, got %d", len(result))
	}
}

// TestFilterBySeverity_EmptyInput returns nil.
func TestFilterBySeverity_EmptyInput(t *testing.T) {
	t.Parallel()

	result := filterBySeverity(nil, "error")
	if len(result) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(result))
	}
}

// TestFilterBySeverity_CaseInsensitive handles uppercase severity.
func TestFilterBySeverity_CaseInsensitive(t *testing.T) {
	t.Parallel()

	findings := []astgrep.Finding{
		{Severity: "ERROR"},
		{Severity: "WARNING"},
	}

	result := filterBySeverity(findings, "ERROR")
	if len(result) != 1 {
		t.Errorf("expected 1 error (case insensitive), got %d", len(result))
	}
}
