package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// writeTelemetryRecord is a test helper that writes UsageRecords as JSONL.
func writeTelemetryRecord(t *testing.T, dir string, records []telemetry.UsageRecord) {
	t.Helper()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		return
	}
	dayKey := records[0].Timestamp.UTC().Format("2006-01-02")
	path := filepath.Join(telDir, "usage-"+dayKey+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("writeTelemetryRecord: %v", err)
	}
	defer f.Close()
	for _, r := range records {
		line, _ := json.Marshal(r)
		f.Write(append(line, '\n'))
	}
}

// TestTelemetryReportCmd_EmptyProject verifies that the report command handles
// a project without telemetry data gracefully.
func TestTelemetryReportCmd_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	buf := new(bytes.Buffer)
	telemetryReportCmd.SetOut(buf)
	telemetryReportCmd.SetErr(buf)
	// Reset flags to defaults before each test
	if err := telemetryReportCmd.Flags().Set("days", "30"); err != nil {
		t.Fatal(err)
	}

	// Call RunE directly (avoids cobra's global state for SetArgs)
	if err := telemetryReportCmd.RunE(telemetryReportCmd, []string{}); err != nil {
		t.Fatalf("RunE() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Skill Usage Report") {
		t.Errorf("expected 'Skill Usage Report' header, got: %s", output)
	}
}

// TestTelemetryReportCmd_WithData verifies that the report command shows skill
// data when telemetry records exist.
func TestTelemetryReportCmd_WithData(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	now := time.Now().UTC()
	writeTelemetryRecord(t, dir, []telemetry.UsageRecord{
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: telemetry.OutcomeSuccess, SessionID: "s1"},
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: telemetry.OutcomeSuccess, SessionID: "s2"},
	})

	buf := new(bytes.Buffer)
	telemetryReportCmd.SetOut(buf)
	telemetryReportCmd.SetErr(buf)
	if err := telemetryReportCmd.Flags().Set("days", "30"); err != nil {
		t.Fatal(err)
	}

	if err := telemetryReportCmd.RunE(telemetryReportCmd, []string{}); err != nil {
		t.Fatalf("RunE() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "moai-workflow-tdd") {
		t.Errorf("expected skill ID in output, got: %s", output)
	}
}

// TestTelemetryCmd_Exists verifies that the telemetry command is registered.
func TestTelemetryCmd_Exists(t *testing.T) {
	t.Parallel()

	if telemetryCmd == nil {
		t.Fatal("telemetryCmd should not be nil")
	}
	if telemetryCmd.Use != "telemetry" {
		t.Errorf("telemetryCmd.Use = %q, want %q", telemetryCmd.Use, "telemetry")
	}
}

// TestTelemetryReportCmd_Exists verifies that the report subcommand is registered.
func TestTelemetryReportCmd_Exists(t *testing.T) {
	t.Parallel()

	if telemetryReportCmd == nil {
		t.Fatal("telemetryReportCmd should not be nil")
	}
	if telemetryReportCmd.Use != "report" {
		t.Errorf("telemetryReportCmd.Use = %q, want %q", telemetryReportCmd.Use, "report")
	}
}

// TestTelemetryReportCmd_InvalidDays verifies that non-positive --days values
// return an appropriate error.
func TestTelemetryReportCmd_InvalidDays(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	buf := new(bytes.Buffer)
	telemetryReportCmd.SetOut(buf)
	telemetryReportCmd.SetErr(buf)
	if err := telemetryReportCmd.Flags().Set("days", "0"); err != nil {
		t.Fatal(err)
	}

	err := telemetryReportCmd.RunE(telemetryReportCmd, []string{})
	if err == nil {
		t.Error("expected error for --days=0, got nil")
	}
}
