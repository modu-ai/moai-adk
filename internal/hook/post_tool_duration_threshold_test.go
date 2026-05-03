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

// writeObservabilityYAML는 테스트용 observability.yaml을 작성한다.
func writeObservabilityYAML(t *testing.T, projectRoot, content string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("config 디렉토리 생성 실패: %v", err)
	}
	path := filepath.Join(dir, "observability.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("observability.yaml 작성 실패: %v", err)
	}
}

// TestLoadSlowHookThreshold_Default는 yaml 파일이 부재할 때 default(5000ms)를 반환하는지 검증한다.
// REQ-CC2122-HOOK-002-002
func TestLoadSlowHookThreshold_Default(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	got := loadSlowHookThreshold(dir)
	if got != defaultSlowHookThresholdMs {
		t.Errorf("loadSlowHookThreshold = %d, want %d (default)", got, defaultSlowHookThresholdMs)
	}
}

// TestLoadSlowHookThreshold_CustomValue는 yaml 에 정의된 양의 정수를 우선 반환하는지 검증한다.
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

// TestLoadSlowHookThreshold_MalformedYAML는 yaml 파싱 실패 시 default 로 fallback 하는지 검증한다.
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

// TestLoadSlowHookThreshold_InvalidValue는 0 이하 또는 음수 값일 때 default 로 fallback 하는지 검증한다.
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
			name: "empty_file",
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

// TestLoadSlowHookThreshold_EmptyProjectRoot는 projectRoot가 빈 문자열일 때 default를 반환하는지 검증한다.
func TestLoadSlowHookThreshold_EmptyProjectRoot(t *testing.T) {
	t.Parallel()

	got := loadSlowHookThreshold("")
	if got != defaultSlowHookThresholdMs {
		t.Errorf("loadSlowHookThreshold(\"\") = %d, want %d", got, defaultSlowHookThresholdMs)
	}
}

// TestPostToolDuration_HonorsCustomThreshold는 yaml 에 정의된 임계값(2000ms)이
// hardcoded default(5000ms)보다 우선 적용되는지 통합 검증한다.
// duration_ms=3000 일 때:
//   - default(5000) 기준: skip
//   - custom(2000) 기준: write
// 따라서 yaml read가 정상 동작하면 메트릭이 기록되어야 한다.
// REQ-CC2122-HOOK-002-001 (E2E 검증)
func TestPostToolDuration_HonorsCustomThreshold(t *testing.T) {
	dir := t.TempDir()

	obsDir := filepath.Join(dir, ".moai", "observability")
	if err := os.MkdirAll(obsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// custom threshold = 2000ms (default보다 낮음)
	writeObservabilityYAML(t, dir, `
observability:
  enabled: true
  hook_metrics:
    slow_hook_threshold_ms: 2000
    output_path: .moai/observability/hook-metrics.jsonl
`)

	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// 3000ms: default 5000 기준 skip, custom 2000 기준 write
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
		t.Fatalf("Handle 에러: %v", err)
	}
	if out == nil {
		t.Fatal("Handle nil 반환")
	}

	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatal("custom threshold(2000ms)이 적용되지 않음 — duration_ms=3000 이지만 메트릭 미기록")
	}

	entries := readMetricsJSONL(t, metricsPath)
	if len(entries) != 1 {
		t.Fatalf("메트릭 엔트리 수 = %d, want 1 (custom threshold 적용 시 기록되어야 함)", len(entries))
	}
	if entries[0].DurationMs != 3000 {
		t.Errorf("duration_ms = %v, want 3000", entries[0].DurationMs)
	}
}
