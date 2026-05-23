package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// hookMetricEntry is the structure of a single line in .moai/observability/hook-metrics.jsonl.
type hookMetricEntry struct {
	Ts         string  `json:"ts"`
	Hook       string  `json:"hook"`
	Tool       string  `json:"tool"`
	DurationMs float64 `json:"duration_ms"`
	SessionID  string  `json:"session_id"`
	Outcome    string  `json:"outcome,omitempty"`
}

// newDurationInput is a helper that builds a HookInput containing a duration_ms field.
func newDurationInput(sessionID, toolName string, durationMs int64, hookEventName string) *HookInput {
	raw, _ := json.Marshal(map[string]any{
		"success":     true,
		"duration_ms": durationMs,
	})
	return &HookInput{
		SessionID:     sessionID,
		ToolName:      toolName,
		HookEventName: hookEventName,
		ToolResponse:  raw,
	}
}

// readMetricsJSONL parses the JSONL file at the given path and returns a slice of entries.
func readMetricsJSONL(t *testing.T, path string) []hookMetricEntry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("metric file open failed: %v", err)
	}
	defer func() { _ = f.Close() }()

	var entries []hookMetricEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var e hookMetricEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("JSONL parse failed: %v (line=%q)", err, line)
		}
		entries = append(entries, e)
	}
	return entries
}

// TestPostToolDuration_SlowHookWritesMetric verifies that 1 line is written to
// .moai/observability/hook-metrics.jsonl when duration_ms exceeds the threshold (5000ms).
//
// RED: this test must fail because the PostToolUse handler has not yet implemented
// the duration_ms-based metric writing feature.
func TestPostToolDuration_SlowHookWritesMetric(t *testing.T) {
	// Cannot use t.Parallel because t.Setenv is used
	dir := t.TempDir()

	// Create the observability directory (precondition for enabling metric writes)
	obsDir := filepath.Join(dir, ".moai", "observability")
	if err := os.MkdirAll(obsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	input := newDurationInput("sess-slow-001", "Bash", 6000, "PostToolUse")
	input.CWD = dir

	h := NewPostToolHandler()
	ctx := context.Background()
	out, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}

	// The metric file must be created
	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatalf("hook-metrics.jsonl not created: %s", metricsPath)
	}

	entries := readMetricsJSONL(t, metricsPath)
	if len(entries) != 1 {
		t.Fatalf("metric entry count = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Hook != "handle-post-tool" {
		t.Errorf("hook = %q, want %q", e.Hook, "handle-post-tool")
	}
	if e.Tool != "Bash" {
		t.Errorf("tool = %q, want %q", e.Tool, "Bash")
	}
	if e.DurationMs != 6000 {
		t.Errorf("duration_ms = %v, want 6000", e.DurationMs)
	}
	if e.SessionID != "sess-slow-001" {
		t.Errorf("session_id = %q, want %q", e.SessionID, "sess-slow-001")
	}
	if e.Ts == "" {
		t.Error("ts field is not set")
	}
	// Verify ts is in ISO 8601 form
	if _, err := time.Parse(time.RFC3339, e.Ts); err != nil {
		t.Errorf("ts parse failed (expected ISO 8601): %v", err)
	}
}

// TestPostToolDuration_FastHookSkipsMetric verifies that no metric is recorded
// when duration_ms is at or below the threshold.
func TestPostToolDuration_FastHookSkipsMetric(t *testing.T) {
	dir := t.TempDir()

	// Create the observability directory
	obsDir := filepath.Join(dir, ".moai", "observability")
	if err := os.MkdirAll(obsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// 1000ms < 5000ms threshold
	input := newDurationInput("sess-fast-001", "Read", 1000, "PostToolUse")
	input.CWD = dir

	h := NewPostToolHandler()
	ctx := context.Background()
	_, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// The metric file must not be created
	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); !os.IsNotExist(err) {
		// If the file exists, verify its contents (it may be empty)
		if data, readErr := os.ReadFile(metricsPath); readErr == nil && len(strings.TrimSpace(string(data))) > 0 {
			t.Errorf("metric recorded for fast hook (expected: none): %s", string(data))
		}
	}
}

// TestPostToolDuration_NoObsDirSkipsMetricSilently verifies the hook does not fail
// and silently skips when the observability directory does not exist.
func TestPostToolDuration_NoObsDirSkipsMetricSilently(t *testing.T) {
	dir := t.TempDir()

	// Do not create the observability directory
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	input := newDurationInput("sess-nodir-001", "Write", 9000, "PostToolUse")
	input.CWD = dir

	h := NewPostToolHandler()
	ctx := context.Background()
	out, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error (no observability): %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}

	// The observability directory must not be newly created
	obsDir := filepath.Join(dir, ".moai", "observability")
	if _, err := os.Stat(obsDir); !os.IsNotExist(err) {
		t.Error("observability directory was auto-created (expected: not created)")
	}
}

// TestPostToolFailureDuration_SlowHookWritesFailureOutcome verifies that, on the
// PostToolUseFailure event, outcome: "failure" is recorded when duration_ms exceeds the threshold.
func TestPostToolFailureDuration_SlowHookWritesFailureOutcome(t *testing.T) {
	dir := t.TempDir()

	obsDir := filepath.Join(dir, ".moai", "observability")
	if err := os.MkdirAll(obsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// Input for PostToolUseFailure (the failure handler may receive a ToolResponse containing duration_ms)
	raw, _ := json.Marshal(map[string]any{
		"duration_ms": int64(7500),
	})
	input := &HookInput{
		SessionID:     "sess-fail-001",
		ToolName:      "Bash",
		HookEventName: "PostToolUseFailure",
		Error:         "exit status 1",
		CWD:           dir,
		ToolResponse:  raw,
	}

	h := NewPostToolUseFailureHandler()
	ctx := context.Background()
	out, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}

	// Verify the metric file
	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatalf("hook-metrics.jsonl not created: %s", metricsPath)
	}

	entries := readMetricsJSONL(t, metricsPath)
	if len(entries) != 1 {
		t.Fatalf("metric entry count = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Outcome != "failure" {
		t.Errorf("outcome = %q, want %q", e.Outcome, "failure")
	}
	if e.Hook != "handle-post-tool-failure" {
		t.Errorf("hook = %q, want %q", e.Hook, "handle-post-tool-failure")
	}
}
