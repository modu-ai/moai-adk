package hook

import (
"encoding/json"
"fmt"
"os"
"path/filepath"
"time"

"gopkg.in/yaml.v3"
)

// defaultSlowHookThresholdMs is the default threshold value (milliseconds) used when
// yaml configuration is missing or invalid.
// REQ-CC2122-HOOK-001-004 (default), REQ-CC2122-HOOK-002-002 (fallback)
const defaultSlowHookThresholdMs int64 = 5000

// hookMetricsRelPath is the path to the metrics file relative to the project root.
const hookMetricsRelPath = ".moai/observability/hook-metrics.jsonl"

// observabilityConfigRelPath is the path to the yaml configuration file relative to the project root.
const observabilityConfigRelPath = ".moai/config/sections/observability.yaml"

// observabilityConfig is the structure of the hook_metrics section in observability.yaml.
// Other fields (trace_dir, retention_days, etc.) are ignored as they are not this helper's concern.
type observabilityConfig struct {
Observability struct {
HookMetrics struct {
SlowHookThresholdMs int64 `yaml:"slow_hook_threshold_ms"`
} `yaml:"hook_metrics"`
} `yaml:"observability"`
}

// loadSlowHookThreshold reads observability.hook_metrics.slow_hook_threshold_ms from
// .moai/config/sections/observability.yaml and returns it.
//
// Falls back to defaultSlowHookThresholdMs (5000ms) in all of the following cases:
// - file does not exist
// - yaml parsing fails
// - key is missing
// - 0 or invalid value
//
// REQ-CC2122-HOOK-002-001 (yaml priority), REQ-CC2122-HOOK-002-002 (fallback),
// REQ-CC2122-HOOK-002-003 (invalid value guard)
//
// Call frequency is low (only on slow hook occurrence) and file size is small (< 1KB),
// so re-reading yaml on every call has no performance impact without caching.
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

// hookDurationEntry is the structure of a JSON line recorded in hook-metrics.jsonl.
// REQ-CC2122-HOOK-001-002
type hookDurationEntry struct {
Ts string `json:"ts"`
Hook string `json:"hook"`
Tool string `json:"tool"`
DurationMs float64 `json:"duration_ms"`
SessionID string `json:"session_id"`
Outcome string `json:"outcome,omitempty"`
}

// extractDurationMs extracts the duration_ms field from HookInput.ToolResponse JSON.
// Returns 0 if the field does not exist or parsing fails.
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

// writeHookMetric atomically adds one line to hook-metrics.jsonl when duration_ms exceeds
// the threshold and the observability directory exists.
// Silently skips when directory does not exist (exit 0 behavior).
// REQ-CC2122-HOOK-001-002, REQ-CC2122-HOOK-001-003, REQ-CC2122-HOOK-001-004
// REQ-CC2122-HOOK-002-001: threshold from observability.yaml, default fallback
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
// REQ-CC2122-HOOK-001-003: silently skip when directory does not exist
return
}

entry := hookDurationEntry{
Ts: time.Now().UTC().Format(time.RFC3339),
Hook: hookName,
Tool: input.ToolName,
DurationMs: durationMs,
SessionID: input.SessionID,
Outcome: outcome,
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
