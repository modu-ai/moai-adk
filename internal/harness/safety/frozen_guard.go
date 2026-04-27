// Package safety — 5-Layer Safety Architecture.
// Layer 1: Frozen Guard (REQ-HL-006).
// FROZEN 경로에 대한 학습 자동 업데이트를 차단하고 위반을 기록한다.
package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// frozenPrefixes는 학습 자동 업데이트가 절대 불가능한 경로 접두사 목록이다.
// REQ-HL-006: 이 목록은 hardcoded이며 어떠한 config/env로도 변경할 수 없다.
//
// [HARD] 이 상수는 configuration 파일이나 환경변수로 Override 불가능하다.
var frozenPrefixes = []string{
	".claude/agents/moai/",
	".claude/skills/moai-",
	".claude/rules/moai/",
	".moai/project/brand/",
}

// IsFrozen은 path가 FROZEN 영역에 해당하면 true를 반환한다.
// REQ-HL-006: filepath.Clean으로 정규화 후 hardcoded frozenPrefixes와 대조한다.
//
// 특성:
//   - 빈 경로는 false (차단할 대상 없음)
//   - 역슬래시는 슬래시로 변환하여 OS 독립적으로 처리
//   - filepath.Clean으로 ".." 등을 제거한 후 평가
//   - 절대경로는 frozenPrefixes(상대경로 기반)와 매칭되지 않음
//
// [HARD] config/env로 이 함수의 동작을 변경할 수 없다.
func IsFrozen(path string) bool {
	if path == "" {
		return false
	}

	// 역슬래시 → 슬래시 변환 (Windows 호환)
	norm := filepath.ToSlash(path)

	// filepath.Clean으로 "..", "." 등 정규화
	norm = filepath.ToSlash(filepath.Clean(norm))

	// hardcoded prefix 목록과 대조
	for _, prefix := range frozenPrefixes {
		if hasPrefix(norm, prefix) {
			return true
		}
	}

	return false
}

// hasPrefix는 s가 prefix로 시작하는지 확인한다.
// filepath.Clean 후에는 이미 슬래시가 통일되어 있으므로 단순 비교한다.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// violationEntry는 frozen-guard-violations.jsonl의 단일 라인 스키마이다.
type violationEntry struct {
	// Timestamp는 위반 발생 시각 (UTC RFC3339).
	Timestamp time.Time `json:"timestamp"`

	// Path는 위반이 탐지된 대상 경로이다.
	Path string `json:"path"`

	// Caller는 위반을 일으킨 호출자 식별자이다.
	Caller string `json:"caller"`

	// Message는 위반 상세 메시지이다.
	Message string `json:"message"`
}

// LogViolation은 FROZEN 경로 접근 위반을 logPath에 JSONL로 기록하고
// stderr에 경고 메시지를 출력한다.
// REQ-HL-006: 위반은 non-blocking — 기록 실패해도 프로세스를 중단하지 않는다.
//
// @MX:ANCHOR: [AUTO] LogViolation은 FROZEN 위반 기록의 단일 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: frozen_guard_test.go, pipeline.go, Phase 4 coordinator
func LogViolation(logPath, path, caller string) error {
	entry := violationEntry{
		Timestamp: time.Now().UTC(),
		Path:      path,
		Caller:    caller,
		Message:   fmt.Sprintf("FROZEN_VIOLATION: %s 경로에 대한 자동 업데이트 시도 차단", path),
	}

	// 부모 디렉토리 생성
	dir := filepath.Dir(logPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("safety/frozen_guard: 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// JSONL 직렬화
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("safety/frozen_guard: 위반 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	// append 모드로 기록
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("safety/frozen_guard: 위반 로그 파일 열기 실패 %s: %w", logPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("safety/frozen_guard: 위반 로그 쓰기 실패: %w", err)
	}

	// stderr 경고 출력 (non-blocking, 학습 시스템 관찰자용)
	fmt.Fprintf(os.Stderr, "[WARN] safety/frozen_guard: FROZEN 경로 접근 차단: %s (caller: %s)\n", path, caller)

	return nil
}
