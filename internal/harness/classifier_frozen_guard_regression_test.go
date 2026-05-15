// Package harness — classifier_frozen_guard_regression_test.go
// Wave D T-D4: Frozen 계약 회귀 방지 테스트.
// EC-A5: Stage2Enabled 제로값 = false — 절대 변경 금지.
// EC-A6: AggregatePatterns 시그니처 동결 — func(logPath string) (map[string]*Pattern, error).
// REQ-HRN-CLS-001: Stage2Enabled=false 시 Stage-1 byte-identical 경로 유지.
package harness

import (
	"path/filepath"
	"testing"
)

// TestFrozenGuard_Stage2EnabledFalseByDefault는 ClassifierConfig 제로값에서
// Stage2Enabled가 false임을 검증한다 (EC-A5, REQ-HRN-CLS-001).
// 이 테스트가 깨지면 backward compatibility 계약 위반 — 절대 실패 허용 금지.
func TestFrozenGuard_Stage2EnabledFalseByDefault(t *testing.T) {
	t.Parallel()

	var cfg ClassifierConfig
	if cfg.Stage2Enabled {
		t.Fatal("[FROZEN CONTRACT VIOLATION] ClassifierConfig 제로값의 Stage2Enabled = true (EC-A5 위반)")
	}
}

// TestFrozenGuard_AggregatePatternsFunctionSignature는 AggregatePatterns가
// func(logPath string) (map[string]*Pattern, error) 시그니처를 유지하는지 컴파일 타임 검증한다.
// EC-A6: AggregatePatterns 시그니처 동결.
func TestFrozenGuard_AggregatePatternsFunctionSignature(t *testing.T) {
	t.Parallel()

	// 컴파일 타임 타입 검사: 시그니처가 맞지 않으면 빌드 실패.
	// checkAggregateSignature는 AggregatePatterns 함수를 정확한 타입으로 받는 헬퍼다.
	checkAggregateSignature := func(fn func(string) (map[string]*Pattern, error)) {
		_ = fn
	}
	checkAggregateSignature(AggregatePatterns)
	t.Log("AggregatePatterns 시그니처 검증 PASS")
}

// TestFrozenGuard_Stage1BackwardCompatOnEmptyLog는 Stage2Enabled=false 시
// 빈 로그 파일로 AggregatePatterns가 빈 맵을 반환하는지 검증한다 (Wave A 회귀).
func TestFrozenGuard_Stage1BackwardCompatOnEmptyLog(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")

	// 파일 없음 → 빈 맵 반환 (REQ-HL-002)
	patterns, err := AggregatePatterns(logPath)
	if err != nil {
		t.Fatalf("AggregatePatterns 오류: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("패턴 수 = %d, want 0 (파일 없음)", len(patterns))
	}
}

// TestFrozenGuard_Stage2OffPassthroughPreservesAllPatterns는 Stage2Enabled=false 시
// clusterSingletons가 입력 패턴 맵을 변경 없이 반환하는지 검증한다 (REQ-HRN-CLS-001).
func TestFrozenGuard_Stage2OffPassthroughPreservesAllPatterns(t *testing.T) {
	t.Parallel()

	patterns := map[string]*Pattern{
		"user_prompt:a:h1": {Key: "user_prompt:a:h1", EventType: EventTypeUserPrompt, Subject: "a", Count: 1, Confidence: defaultConfidence},
		"user_prompt:b:h2": {Key: "user_prompt:b:h2", EventType: EventTypeUserPrompt, Subject: "b", Count: 1, Confidence: defaultConfidence},
	}
	events := []Event{
		{EventType: EventTypeUserPrompt, Subject: "a", ContextHash: "h1"},
		{EventType: EventTypeUserPrompt, Subject: "b", ContextHash: "h2"},
	}

	cfg := ClassifierConfig{} // Stage2Enabled=false
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// Stage-2 비활성: 입력 패턴 맵이 그대로 반환되어야 함
	if len(result) != len(patterns) {
		t.Errorf("패턴 수 변화: got %d, want %d (Stage-2 off passthrough 위반)", len(result), len(patterns))
	}
	for k := range patterns {
		if _, ok := result[k]; !ok {
			t.Errorf("패턴 키 %q 유실됨", k)
		}
	}
}
