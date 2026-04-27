// Package harness — usage-log.jsonl 이벤트 기록기.
// REQ-HL-001: PostToolUse hook handler로 실행되며 이벤트당 <100ms.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Observer는 .moai/harness/usage-log.jsonl 파일에 이벤트를 기록한다.
// 이 구조체는 Zero Value 사용이 불가하며 NewObserver로 생성해야 한다.
//
// @MX:ANCHOR: [AUTO] RecordEvent는 모든 hook 경로에서 호출되는 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: observer_test.go, integration_test.go, hook CLI 경로
type Observer struct {
	// logPath는 usage-log.jsonl 파일의 절대/상대 경로이다.
	logPath string

	// retention은 lazy pruning 컴포넌트이다 (nil이면 pruning 비활성화).
	retention *Retention

	// nowFn은 현재 시각을 반환하는 함수 (테스트에서 override 가능).
	nowFn func() time.Time
}

// NewObserver는 지정된 logPath를 사용하는 Observer를 생성한다.
// logPath의 부모 디렉토리가 없으면 RecordEvent 시점에 자동으로 생성된다.
func NewObserver(logPath string) *Observer {
	return &Observer{
		logPath: logPath,
		nowFn:   time.Now,
	}
}

// NewObserverWithRetention은 retention pruning이 활성화된 Observer를 생성한다.
func NewObserverWithRetention(logPath string, retention *Retention) *Observer {
	obs := NewObserver(logPath)
	obs.retention = retention
	return obs
}

// RecordEvent는 이벤트를 usage-log.jsonl에 JSONL 단일 라인으로 기록한다.
// REQ-HL-001: 각 호출은 <100ms 이내에 완료되어야 한다.
//
// 파일이 없으면 생성하고, 있으면 append 방식으로 기록한다.
// 부모 디렉토리가 없으면 자동 생성한다.
//
// @MX:TODO: [AUTO] Phase 4: learning.enabled 설정으로 gate 추가 예정.
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

	// 부모 디렉토리 자동 생성
	if dir := filepath.Dir(o.logPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("observer: 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// JSONL 직렬화
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

	// lazy pruning: retention이 설정된 경우 pruning 시도 (1시간 skip 로직 포함)
	if o.retention != nil {
		// pruning 실패는 기록 실패로 이어지지 않는다 (non-blocking)
		_ = o.retention.PruneStaleEntries(defaultRetentionDays)
	}

	return nil
}

// defaultRetentionDays는 기본 로그 보존 일수이다.
// Phase 4에서 설정 파일 연동 시 이 값을 대체한다.
const defaultRetentionDays = 30

// ─────────────────────────────────────────────
// 패키지 내부 헬퍼 (테스트에서도 사용)
// ─────────────────────────────────────────────

// readFile은 파일 내용을 바이트로 읽는다.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// appendEventsJSONL는 이벤트 슬라이스를 파일에 JSONL 형식으로 append한다.
// 테스트 헬퍼 및 retention에서 사용한다.
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

// listDir은 디렉토리의 파일명 목록을 반환한다. 디렉토리가 없으면 빈 슬라이스를 반환한다.
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
