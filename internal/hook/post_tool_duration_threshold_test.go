package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// @MX:NOTE: [AUTO] post_tool_duration_threshold_test.go — REQ-CC2122-HOOK-002 yaml read
// @MX:SPEC: SPEC-CC2122-HOOK-002 REQ-001/002/003

// writeObservabilityYAML writes a test observability.yaml.
func writeObservabilityYAML(t *testing.T, projectRoot, content string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("config directory create failed: %v", err)
	}
	path := filepath.Join(dir, "observability.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("observability.yaml write failed: %v", err)
	}
}

// TestLoadSlowHookThreshold_Default verifies the default (5000ms) is returned when the yaml file is absent.
// REQ-CC2122-HOOK-002-002
func TestLoadSlowHookThreshold_Default(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	got := loadSlowHookThreshold(dir)
	if got != defaultSlowHookThresholdMs {
		t.Errorf("loadSlowHookThreshold = %d, want %d (default)", got, defaultSlowHookThresholdMs)
	}
}

// TestLoadSlowHookThreshold_CustomValue verifies that a positive integer defined in the yaml is returned with priority.
// REQ-CC2122-HOOK-002-001
func TestLoadSlowHookThreshold_CustomValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeObservabilityYAML(t, dir, `
observability:
  enabled: true
  hook_metrics:
    slow_hook_threshold_ms: 7500
    output_path: .moai/observability/hook-metrics.jsonl
`)

	got := loadSlowHookThreshold(dir)
	if got != 7500 {
		t.Errorf("loadSlowHookThreshold = %d, want 7500", got)
	}
}

// TestLoadSlowHookThreshold_MalformedYAML verifies the fallback to the default when yaml parsing fails.
// REQ-CC2122-HOOK-002-002
func TestLoadSlowHookThreshold_MalformedYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeObservabilityYAML(t, dir, "this: is: not: valid: yaml: ::")

	got := loadSlowHookThreshold(dir)
	if got != defaultSlowHookThresholdMs {
		t.Errorf("loadSlowHookThreshold = %d, want %d (default fallback for malformed yaml)", got, defaultSlowHookThresholdMs)
	}
}

// TestLoadSlowHookThreshold_InvalidValue verifies the fallback to the default for zero or negative values.
// REQ-CC2122-HOOK-002-003
func TestLoadSlowHookThreshold_InvalidValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		content string
	}{
		{
			name: "zero",
			content: `
observability:
  hook_metrics:
    slow_hook_threshold_ms: 0
`,
		},
		{
			name: "negative",
			content: `
observability:
  hook_metrics:
    slow_hook_threshold_ms: -100
`,
		},
		{
			name: "missing_key",
			content: `
observability:
  hook_metrics:
    output_path: .moai/observability/hook-metrics.jsonl
`,
		},
		{
			name:    "empty_file",
			content: ``,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			writeObservabilityYAML(t, dir, tc.content)

			got := loadSlowHookThreshold(dir)
			if got != defaultSlowHookThresholdMs {
				t.Errorf("loadSlowHookThreshold = %d, want %d (default fallback)", got, defaultSlowHookThresholdMs)
			}
		})
	}
}

// TestLoadSlowHookThreshold_EmptyProjectRoot verifies the default is returned when projectRoot is empty.
func TestLoadSlowHookThreshold_EmptyProjectRoot(t *testing.T) {
	t.Parallel()

	got := loadSlowHookThreshold("")
	if got != defaultSlowHookThresholdMs {
		t.Errorf("loadSlowHookThreshold(\"\") = %d, want %d", got, defaultSlowHookThresholdMs)
	}
}

// TestPostToolDuration_HonorsCustomThreshold integration-verifies that the yaml-defined
// threshold (2000ms) takes precedence over the hardcoded default (5000ms).
// When duration_ms=3000:
//   - default(5000) basis: skip
//   - custom(2000) basis: write
//
// Therefore the metric must be recorded if the yaml read works correctly.
// REQ-CC2122-HOOK-002-001 (E2E verification)
func TestPostToolDuration_HonorsCustomThreshold(t *testing.T) {
	dir := t.TempDir()

	obsDir := filepath.Join(dir, ".moai", "observability")
	if err := os.MkdirAll(obsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// custom threshold = 2000ms (below the default)
	writeObservabilityYAML(t, dir, `
observability:
  enabled: true
  hook_metrics:
    slow_hook_threshold_ms: 2000
    output_path: .moai/observability/hook-metrics.jsonl
`)

	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// 3000ms: skip with default 5000, write with custom 2000
	raw, _ := json.Marshal(map[string]any{"success": true, "duration_ms": int64(3000)})
	input := &HookInput{
		SessionID:     "sess-custom-threshold-001",
		ToolName:      "Bash",
		HookEventName: "PostToolUse",
		CWD:           dir,
		ToolResponse:  raw,
	}

	h := NewPostToolHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil")
	}

	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatal("custom threshold (2000ms) was not applied — metric not written even though duration_ms=3000")
	}

	entries := readMetricsJSONL(t, metricsPath)
	if len(entries) != 1 {
		t.Fatalf("metric entry count = %d, want 1 (must be written when custom threshold is applied)", len(entries))
	}
	if entries[0].DurationMs != 3000 {
		t.Errorf("duration_ms = %v, want 3000", entries[0].DurationMs)
	}
}
