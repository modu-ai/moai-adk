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

// defaultRetentionDays is the default log retention period in days.
// This value will be replaced when config file integration occurs in Phase 4.
const defaultRetentionDays = 30

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
