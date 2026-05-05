package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// defaultSlowHookThresholdMs는 yaml 설정이 없거나 invalid 일 때 사용되는
// 기본 임계값(밀리초)이다.
// REQ-CC2122-HOOK-001-004 (default), REQ-CC2122-HOOK-002-002 (fallback)
const defaultSlowHookThresholdMs int64 = 5000

// hookMetricsRelPath는 프로젝트 루트에서 메트릭 파일까지의 상대 경로이다.
const hookMetricsRelPath = ".moai/observability/hook-metrics.jsonl"

// observabilityConfigRelPath는 yaml 설정 파일의 프로젝트 루트 기준 상대 경로이다.
const observabilityConfigRelPath = ".moai/config/sections/observability.yaml"

// observabilityConfig는 observability.yaml 의 hook_metrics 섹션 구조이다.
// 다른 필드(trace_dir, retention_days 등)는 본 helper의 관심사가 아니므로 무시된다.
type observabilityConfig struct {
	Observability struct {
		HookMetrics struct {
			SlowHookThresholdMs int64 `yaml:"slow_hook_threshold_ms"`
		} `yaml:"hook_metrics"`
	} `yaml:"observability"`
}

// loadSlowHookThreshold는 .moai/config/sections/observability.yaml 에서
// observability.hook_metrics.slow_hook_threshold_ms 를 읽어 반환한다.
//
// 다음 경우 모두 defaultSlowHookThresholdMs (5000ms) 로 fallback 한다:
//   - 파일 부재
//   - yaml 파싱 실패
//   - 키 누락
//   - 0 이하의 invalid 값
//
// REQ-CC2122-HOOK-002-001 (yaml 우선), REQ-CC2122-HOOK-002-002 (fallback),
// REQ-CC2122-HOOK-002-003 (invalid value 보호)
//
// 호출 빈도가 낮고 (slow hook 발생 시에만) 파일 크기가 작아서 (< 1KB)
// 캐싱 없이 매 호출 시 yaml 을 다시 읽어도 성능 영향이 없다.
func loadSlowHookThreshold(projectRoot string) int64 {
	if projectRoot == "" {
		return defaultSlowHookThresholdMs
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, observabilityConfigRelPath))
	if err != nil {
		return defaultSlowHookThresholdMs
	}
	var cfg observabilityConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultSlowHookThresholdMs
	}
	if cfg.Observability.HookMetrics.SlowHookThresholdMs <= 0 {
		return defaultSlowHookThresholdMs
	}
	return cfg.Observability.HookMetrics.SlowHookThresholdMs
}

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
// REQ-CC2122-HOOK-002-001: 임계값은 observability.yaml 우선, default fallback
func writeHookMetric(input *HookInput, hookName, outcome string) {
	projectRoot := resolveProjectRoot(input)

	threshold := loadSlowHookThreshold(projectRoot)
	durationMs := extractDurationMs(input)
	if durationMs <= float64(threshold) {
		return
	}

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
