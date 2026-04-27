// Package harness — 샘플 세션 replay 통합 테스트.
// REQ-HL-001: 실제 세션과 유사한 이벤트 시퀀스를 재생하여 전체 파이프라인을 검증한다.
package harness

import (
	"encoding/json"
	"testing"
	"time"
)

// TestIntegration_SampleSessionReplay는 실제 /moai 세션에서 발생하는
// 이벤트 시퀀스를 재생하여 observer + retention 파이프라인을 검증한다.
// T-P1-05: integration_test.go 샘플 세션 replay.
func TestIntegration_SampleSessionReplay(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/.moai/harness/usage-log.jsonl"
	archiveDir := dir + "/.moai/harness/learning-history/archive"

	// 고정 시각 사용 (테스트 결정론성)
	baseTime := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	nowIdx := 0
	times := []time.Time{
		baseTime,
		baseTime.Add(1 * time.Second),
		baseTime.Add(2 * time.Second),
		baseTime.Add(3 * time.Second),
		baseTime.Add(4 * time.Second),
	}
	nowFn := func() time.Time {
		t := times[nowIdx%len(times)]
		nowIdx++
		return t
	}

	retention := NewRetention(logPath, archiveDir, func() time.Time { return baseTime })
	obs := &Observer{
		logPath:   logPath,
		retention: retention,
		nowFn:     nowFn,
	}

	// 샘플 세션 이벤트 시퀀스
	events := []struct {
		et      EventType
		subject string
		hash    string
	}{
		{EventTypeMoaiSubcommand, "/moai plan", "ctx-abc"},
		{EventTypeSpecReference, "SPEC-V3R3-HARNESS-LEARNING-001", "ctx-abc"},
		{EventTypeAgentInvocation, "manager-spec", "ctx-abc"},
		{EventTypeAgentInvocation, "plan-auditor", "ctx-abc"},
		{EventTypeFeedback, "/moai feedback", "ctx-abc"},
	}

	for _, e := range events {
		if err := obs.RecordEvent(e.et, e.subject, e.hash); err != nil {
			t.Fatalf("RecordEvent(%s) 실패: %v", e.et, err)
		}
	}

	// 로그 파일 검증
	data, err := readFile(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	if len(lines) != len(events) {
		t.Errorf("기록된 이벤트 수: want=%d got=%d", len(events), len(lines))
	}

	// 각 라인이 유효한 JSON이며 필수 필드가 있는지 확인
	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("줄 %d 파싱 실패: %v", i, err)
			continue
		}
		if evt.SchemaVersion != LogSchemaVersion {
			t.Errorf("줄 %d SchemaVersion: want=%q got=%q", i, LogSchemaVersion, evt.SchemaVersion)
		}
		if evt.EventType == "" {
			t.Errorf("줄 %d EventType이 비어있음", i)
		}
		if evt.Subject == "" {
			t.Errorf("줄 %d Subject이 비어있음", i)
		}
		if evt.Timestamp.IsZero() {
			t.Errorf("줄 %d Timestamp가 zero", i)
		}
	}
}

// TestIntegration_RetentionWithObserver는 Observer가 retention과 통합되어
// 오래된 이벤트를 pruning하는 전체 흐름을 검증한다.
// T-P1-05: integration test.
func TestIntegration_RetentionWithObserver(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"
	archiveDir := dir + "/archive"

	// 오래된 이벤트를 미리 기록
	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldEvent := Event{
		Timestamp:     past,
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "old-subcommand",
		ContextHash:   "old-hash",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	if err := appendEventsJSONL(logPath, []Event{oldEvent}); err != nil {
		t.Fatalf("오래된 이벤트 기록 실패: %v", err)
	}

	// Observer + Retention 통합: now는 2026-04-27
	now := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	retention := NewRetention(logPath, archiveDir, func() time.Time { return now })

	// defaultRetentionDays(30일) 기준: 2026-01-01은 pruning 대상
	obs := NewObserverWithRetention(logPath, retention)

	// 새 이벤트 기록 — 이때 lazy pruning이 실행된다
	if err := obs.RecordEvent(EventTypeAgentInvocation, "expert-backend", "new-hash"); err != nil {
		t.Fatalf("RecordEvent 실패: %v", err)
	}

	// 오래된 이벤트가 제거되었는지 확인
	data, err := readFile(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	// 새 이벤트 1개만 남아야 함
	if len(lines) != 1 {
		t.Errorf("pruning 후 줄 수: want=1 got=%d", len(lines))
	}

	// 아카이브 파일이 생성되었는지 확인
	archiveFiles := listDir(archiveDir)
	if len(archiveFiles) == 0 {
		t.Error("아카이브 파일이 생성되지 않았습니다")
	}
}

// TestIntegration_JSONL_AllEventTypes는 모든 EventType을 기록하고 파싱하는 왕복을 검증한다.
func TestIntegration_JSONL_AllEventTypes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obs := NewObserver(dir + "/usage-log.jsonl")

	allTypes := []EventType{
		EventTypeMoaiSubcommand,
		EventTypeAgentInvocation,
		EventTypeSpecReference,
		EventTypeFeedback,
	}

	for _, et := range allTypes {
		if err := obs.RecordEvent(et, "test-subject", "test-hash"); err != nil {
			t.Errorf("RecordEvent(%s) 실패: %v", et, err)
		}
	}

	data, err := readFile(dir + "/usage-log.jsonl")
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	if len(lines) != len(allTypes) {
		t.Errorf("이벤트 수: want=%d got=%d", len(allTypes), len(lines))
	}

	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("줄 %d 파싱 실패: %v", i, err)
			continue
		}
		if evt.EventType != allTypes[i] {
			t.Errorf("줄 %d EventType: want=%q got=%q", i, allTypes[i], evt.EventType)
		}
	}
}
