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

// hookMetricEntry는 .moai/observability/hook-metrics.jsonl 한 줄의 구조이다.
type hookMetricEntry struct {
	Ts         string  `json:"ts"`
	Hook       string  `json:"hook"`
	Tool       string  `json:"tool"`
	DurationMs float64 `json:"duration_ms"`
	SessionID  string  `json:"session_id"`
	Outcome    string  `json:"outcome,omitempty"`
}

// newDurationInput은 duration_ms 필드를 포함한 HookInput을 생성하는 헬퍼이다.
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

// readMetricsJSONL는 지정 경로의 JSONL 파일을 파싱해 엔트리 슬라이스를 반환한다.
func readMetricsJSONL(t *testing.T, path string) []hookMetricEntry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("메트릭 파일 열기 실패: %v", err)
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
			t.Fatalf("JSONL 파싱 실패: %v (line=%q)", err, line)
		}
		entries = append(entries, e)
	}
	return entries
}

// TestPostToolDuration_SlowHookWritesMetric은 duration_ms가 임계값(5000ms)을
// 초과할 때 .moai/observability/hook-metrics.jsonl에 1줄이 기록되는지 검증한다.
//
// RED: 이 테스트는 PostToolUse 핸들러가 duration_ms 기반 메트릭 기록 기능을
// 아직 구현하지 않았으므로 실패해야 한다.
func TestPostToolDuration_SlowHookWritesMetric(t *testing.T) {
	// t.Setenv 사용으로 병렬 실행 불가
	dir := t.TempDir()

	// observability 디렉토리 생성 (메트릭 쓰기 활성화 조건)
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
		t.Fatalf("Handle() 에러: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() nil 반환")
	}

	// 메트릭 파일이 생성되어야 함
	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatalf("hook-metrics.jsonl 미생성: %s", metricsPath)
	}

	entries := readMetricsJSONL(t, metricsPath)
	if len(entries) != 1 {
		t.Fatalf("메트릭 엔트리 수 = %d, want 1", len(entries))
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
		t.Error("ts 필드 미설정")
	}
	// ts가 ISO 8601 형식인지 확인
	if _, err := time.Parse(time.RFC3339, e.Ts); err != nil {
		t.Errorf("ts 파싱 실패 (ISO 8601 기대): %v", err)
	}
}

// TestPostToolDuration_FastHookSkipsMetric은 duration_ms가 임계값 이하일 때
// 메트릭을 기록하지 않는지 검증한다.
func TestPostToolDuration_FastHookSkipsMetric(t *testing.T) {
	dir := t.TempDir()

	// observability 디렉토리 생성
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
		t.Fatalf("Handle() 에러: %v", err)
	}

	// 메트릭 파일이 생성되지 않아야 함
	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); !os.IsNotExist(err) {
		// 파일이 존재하면 내용 확인 (빈 파일일 수 있음)
		if data, readErr := os.ReadFile(metricsPath); readErr == nil && len(strings.TrimSpace(string(data))) > 0 {
			t.Errorf("fast hook에서 메트릭이 기록됨 (기대: 없음): %s", string(data))
		}
	}
}

// TestPostToolDuration_NoObsDirSkipsMetricSilently는 observability 디렉토리가
// 없을 때 훅이 실패하지 않고 조용히 건너뛰는지 검증한다.
func TestPostToolDuration_NoObsDirSkipsMetricSilently(t *testing.T) {
	dir := t.TempDir()

	// observability 디렉토리를 생성하지 않음
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	input := newDurationInput("sess-nodir-001", "Write", 9000, "PostToolUse")
	input.CWD = dir

	h := NewPostToolHandler()
	ctx := context.Background()
	out, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() 에러 (observability 없음): %v", err)
	}
	if out == nil {
		t.Fatal("Handle() nil 반환")
	}

	// observability 디렉토리가 새로 생성되지 않아야 함
	obsDir := filepath.Join(dir, ".moai", "observability")
	if _, err := os.Stat(obsDir); !os.IsNotExist(err) {
		t.Error("observability 디렉토리가 자동 생성됨 (기대: 미생성)")
	}
}

// TestPostToolFailureDuration_SlowHookWritesFailureOutcome은 PostToolUseFailure
// 이벤트에서 duration_ms가 임계값을 초과할 때 outcome: "failure"가 기록되는지 검증한다.
func TestPostToolFailureDuration_SlowHookWritesFailureOutcome(t *testing.T) {
	dir := t.TempDir()

	obsDir := filepath.Join(dir, ".moai", "observability")
	if err := os.MkdirAll(obsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// PostToolUseFailure용 입력 (failure 핸들러는 duration_ms를 포함한 ToolResponse를 받을 수 있음)
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
		t.Fatalf("Handle() 에러: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() nil 반환")
	}

	// 메트릭 파일 검증
	metricsPath := filepath.Join(obsDir, "hook-metrics.jsonl")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Fatalf("hook-metrics.jsonl 미생성: %s", metricsPath)
	}

	entries := readMetricsJSONL(t, metricsPath)
	if len(entries) != 1 {
		t.Fatalf("메트릭 엔트리 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Outcome != "failure" {
		t.Errorf("outcome = %q, want %q", e.Outcome, "failure")
	}
	if e.Hook != "handle-post-tool-failure" {
		t.Errorf("hook = %q, want %q", e.Hook, "handle-post-tool-failure")
	}
}
