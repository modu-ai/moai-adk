// Package harness — usage-log.jsonl event recorder.
// REQ-HL-001: Runs as PostToolUse hook handler, <100ms per event.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Observer records events to .moai/harness/usage-log.jsonl file.
// This struct cannot use zero value and must be created with NewObserver.
//
// @MX:ANCHOR: [AUTO] RecordEvent is the entry point called from all hook paths.
// @MX:REASON: [AUTO] fan_in >= 3: observer_test.go, integration_test.go, hook CLI path
type Observer struct {
	// logPath is the absolute/relative path to usage-log.jsonl file.
	logPath string

	// retention is the lazy pruning component (nil means pruning disabled).
	retention *Retention

	// nowFn is a function that returns current time (overridable in tests).
	nowFn func() time.Time
}

// NewObserver creates an Observer using the specified logPath.
// Parent directory of logPath is auto-created at RecordEvent time if it does not exist.
func NewObserver(logPath string) *Observer {
	return &Observer{
		logPath: logPath,
		nowFn:   time.Now,
	}
}

// NewObserverWithRetention creates an Observer with retention pruning enabled.
func NewObserverWithRetention(logPath string, retention *Retention) *Observer {
	obs := NewObserver(logPath)
	obs.retention = retention
	return obs
}

// RecordEvent records an event to usage-log.jsonl as a single JSONL line.
// REQ-HL-001: Each call must complete within <100ms.
//
// Creates file if it does not exist, appends if it exists.
// Auto-creates parent directory if it does not exist.
//
// @MX:TODO: [AUTO] Phase 4: Plan to add gate with learning.enabled setting.
// @MX:SPEC: SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-001
func (o *Observer) RecordEvent(eventType EventType, subject, contextHash string) error {
	evt := Event{
		Timestamp:     o.nowFn().UTC(),
		EventType:     eventType,
		Subject:       subject,
		ContextHash:   contextHash,
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}

	// Auto-create parent directory
	if dir := filepath.Dir(o.logPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("observer: 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// JSONL serialization
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("observer: 이벤트 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	// atomic append: O_APPEND|O_CREATE|O_WRONLY
	f, err := os.OpenFile(o.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("observer: 파일 열기 실패 %s: %w", o.logPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("observer: 파일 쓰기 실패 %s: %w", o.logPath, err)
	}

	// lazy pruning: attempt pruning if retention is set (includes 1-hour skip logic)
	if o.retention != nil {
		// pruning failure does not cause recording failure (non-blocking)
		_ = o.retention.PruneStaleEntries(defaultRetentionDays)
	}

	return nil
}

// RecordExtendedEvent records a fully-populated Event to usage-log.jsonl.
// Unlike the existing RecordEvent, accepts the full Event struct including optional fields (omitempty).
// Used by the T-A3/A4/A5 handlers without modifying the existing RecordEvent.
//
// @MX:ANCHOR: [AUTO] RecordExtendedEvent is the common entry point for the Stop/SubagentStop/UserPromptSubmit handlers.
// @MX:REASON: [AUTO] fan_in >= 3: runHarnessObserveStop, runHarnessObserveSubagentStop, runHarnessObserveUserPromptSubmit
func (o *Observer) RecordExtendedEvent(evt Event) error {
	// Fill in default metadata: nowFn and SchemaVersion when unset.
	if evt.Timestamp.IsZero() {
		evt.Timestamp = o.nowFn().UTC()
	}
	if evt.SchemaVersion == "" {
		evt.SchemaVersion = LogSchemaVersion
	}

	// Auto-create parent directory
	if dir := filepath.Dir(o.logPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("observer: 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// JSONL serialization
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("observer: 이벤트 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	// atomic append: O_APPEND|O_CREATE|O_WRONLY
	f, err := os.OpenFile(o.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("observer: 파일 열기 실패 %s: %w", o.logPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("observer: 파일 쓰기 실패 %s: %w", o.logPath, err)
	}

	// lazy pruning: attempt pruning if retention is set (non-blocking)
	if o.retention != nil {
		_ = o.retention.PruneStaleEntries(defaultRetentionDays)
	}

	return nil
}

// defaultRetentionDays is the default log retention period in days.
// This value will be replaced when config file integration occurs in Phase 4.
const defaultRetentionDays = 30

// ─────────────────────────────────────────────
// Context-Governance Axis weight estimation (SPEC-V3R6-CONTEXT-GOV-AXIS-001)
// ─────────────────────────────────────────────

// WeightUnitTokens / WeightUnitBytes are the two accepted WeightUnit values.
// The estimation records bytes (sum of file sizes); a bytes/4 heuristic to
// tokens is left to the reader — exactness is out of scope (spec.md §X.5).
const (
	WeightUnitTokens = "tokens"
	WeightUnitBytes  = "bytes"
)

// eagerWeightSources are the eagerly-loaded context files whose combined byte
// size estimates the eager context weight (REQ-CGA-001). Paths are relative to
// the project root (cwd). Missing files are skipped silently (EC-1 fail-open).
var eagerWeightSources = []string{
	"CLAUDE.md",
	".claude/output-styles/moai/moai.md",
	".claude/agent-memory/manager-develop/MEMORY.md",
}

// eagerWeightRulesGlob is the glob pattern for auto-loaded rule files included
// in the eager weight. Walked via filepath.Glob (fail-open on glob error).
const eagerWeightRulesGlob = ".claude/rules/moai/**/*.md"

// EstimateContextWeight estimates the eager-vs-on-demand context-weight split
// for the current turn, rooted at projectRoot (the hook's cwd). It populates
// the three weight fields on the Event in place (REQ-CGA-001). Fail-open
// guarantee (REQ-CGA-003): on ANY error (unreadable file, stat error, glob
// error), the function returns nil and leaves the Event's weight fields at
// their Go zero values — the caller still stamps schema_version="v2.1" so a
// reader can distinguish estimation-skipped (v2.1 + sentinel) from pre-SPEC
// legacy (v1/v2). The function NEVER returns a non-nil error to the hook
// caller; it is fail-open by construction.
//
// On-demand weight is recorded as 0 in this SPEC — per-skill attribution is
// out of scope (spec.md §X.6); the field exists so future SPECs can populate
// it without another schema bump.
//
// @MX:NOTE: [AUTO] REQ-CGA-001/003: eager-vs-on-demand weight estimator with fail-open.
// @MX:SPEC: SPEC-V3R6-CONTEXT-GOV-AXIS-001
func EstimateContextWeight(evt *Event, projectRoot string) {
	if evt == nil {
		return
	}

	eager, ok := estimateEagerWeight(projectRoot)
	if !ok {
		// fail-open: leave weight fields at zero value; schema_version is still
		// stamped "v2.1" by the recorder, distinguishing estimation-skipped from
		// pre-SPEC legacy lines.
		return
	}

	evt.EagerContextWeight = eager
	evt.OnDemandContextWeight = 0 // §X.6: per-skill attribution deferred
	evt.WeightUnit = WeightUnitBytes
}

// estimateEagerWeight sums the byte sizes of the eager-weight source files
// under projectRoot. Returns (sum, true) on success; (0, false) on any error
// (fail-open). Missing individual files are skipped (EC-1), NOT treated as
// errors — only a total failure (e.g. projectRoot empty) returns false.
func estimateEagerWeight(projectRoot string) (int, bool) {
	if projectRoot == "" {
		return 0, false
	}

	total := 0

	// Fixed-path sources (CLAUDE.md, moai.md, MEMORY.md index).
	for _, rel := range eagerWeightSources {
		abs := filepath.Join(projectRoot, rel)
		info, err := os.Stat(abs)
		if err != nil {
			// EC-1: missing file is skipped, not an error.
			continue
		}
		if info.IsDir() {
			continue
		}
		total += int(info.Size())
	}

	// Globbed rule files (.claude/rules/moai/**/*.md).
	matches, err := filepath.Glob(filepath.Join(projectRoot, eagerWeightRulesGlob))
	if err == nil {
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}
			total += int(info.Size())
		}
	}
	// Glob error is fail-open: we still return the fixed-path subtotal.

	return total, true
}

// ─────────────────────────────────────────────
// Internal package helpers (also used in tests)
// ─────────────────────────────────────────────

// readFile reads file contents as bytes.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// appendEventsJSONL appends event slices to file in JSONL format.
// Used in test helpers and retention.
func appendEventsJSONL(path string, events []Event) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("appendEventsJSONL: 디렉토리 생성 실패: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("appendEventsJSONL: 파일 열기 실패: %w", err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			return fmt.Errorf("appendEventsJSONL: 인코딩 실패: %w", err)
		}
	}
	return nil
}

// listDir returns a list of filenames in a directory. Returns empty slice if directory does not exist.
func listDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}
