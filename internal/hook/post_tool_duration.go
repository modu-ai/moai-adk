package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// slowHookThresholdMs는 메트릭 기록을 트리거하는 기본 임계값(밀리초)이다.
// REQ-CC2122-HOOK-001-004
const slowHookThresholdMs = 5000

// hookMetricsRelPath는 프로젝트 루트에서 메트릭 파일까지의 상대 경로이다.
const hookMetricsRelPath = ".moai/observability/hook-metrics.jsonl"

// hookDurationEntry는 hook-metrics.jsonl에 기록되는 JSON 라인 구조이다.
// REQ-CC2122-HOOK-001-002
type hookDurationEntry struct {
	Ts         string  `json:"ts"`
	Hook       string  `json:"hook"`
	Tool       string  `json:"tool"`
	DurationMs float64 `json:"duration_ms"`
	SessionID  string  `json:"session_id"`
	Outcome    string  `json:"outcome,omitempty"`
}

// extractDurationMs는 HookInput.ToolResponse JSON에서 duration_ms 필드를 추출한다.
// 필드가 없거나 파싱에 실패하면 0을 반환한다.
// REQ-CC2122-HOOK-001-001
func extractDurationMs(input *HookInput) float64 {
	if len(input.ToolResponse) == 0 {
		return 0
	}
	var payload map[string]any
	if err := json.Unmarshal(input.ToolResponse, &payload); err != nil {
		return 0
	}
	switch v := payload["duration_ms"].(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}

// writeHookMetric은 duration_ms가 임계값을 초과하고 observability 디렉토리가
// 존재할 때 hook-metrics.jsonl에 1줄을 원자적으로 추가한다.
// 디렉토리 부재 시 조용히 건너뛴다 (exit 0 보장).
// REQ-CC2122-HOOK-001-002, REQ-CC2122-HOOK-001-003, REQ-CC2122-HOOK-001-004
func writeHookMetric(input *HookInput, hookName, outcome string) {
	durationMs := extractDurationMs(input)
	if durationMs <= slowHookThresholdMs {
		return
	}

	projectRoot := resolveProjectRoot(input)
	if projectRoot == "" {
		return
	}

	obsDir := filepath.Join(projectRoot, ".moai", "observability")
	if _, err := os.Stat(obsDir); os.IsNotExist(err) {
		// REQ-CC2122-HOOK-001-003: 디렉토리 없으면 조용히 건너뜀
		return
	}

	entry := hookDurationEntry{
		Ts:         time.Now().UTC().Format(time.RFC3339),
		Hook:       hookName,
		Tool:       input.ToolName,
		DurationMs: durationMs,
		SessionID:  input.SessionID,
		Outcome:    outcome,
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	metricsPath := filepath.Join(projectRoot, hookMetricsRelPath)
	f, err := os.OpenFile(metricsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintf(f, "%s\n", line)
}
