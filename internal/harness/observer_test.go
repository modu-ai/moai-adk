// Package harness — observer 및 JSONL 스키마 단위 테스트.
package harness

import (
	"encoding/json"
	"testing"
	"time"
)

// ─────────────────────────────────────────────
// T-P1-01: Event marshal/unmarshal 왕복 테스트
// ─────────────────────────────────────────────

// TestEventMarshalUnmarshal은 Event 구조체가 JSON으로 직렬화/역직렬화되는지 검증한다.
// REQ-HL-001: JSONL 한 줄에 timestamp, event_type, subject, context_hash,
// tier_increment 필드가 포함되어야 한다.
func TestEventMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	original := Event{
		Timestamp:     time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "/moai plan",
		ContextHash:   "abc123",
		TierIncrement: 1,
		SchemaVersion: LogSchemaVersion,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal 실패: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal 실패: %v", err)
	}

	if decoded.EventType != original.EventType {
		t.Errorf("EventType: want=%q got=%q", original.EventType, decoded.EventType)
	}
	if decoded.Subject != original.Subject {
		t.Errorf("Subject: want=%q got=%q", original.Subject, decoded.Subject)
	}
	if decoded.ContextHash != original.ContextHash {
		t.Errorf("ContextHash: want=%q got=%q", original.ContextHash, decoded.ContextHash)
	}
	if decoded.TierIncrement != original.TierIncrement {
		t.Errorf("TierIncrement: want=%d got=%d", original.TierIncrement, decoded.TierIncrement)
	}
	if decoded.SchemaVersion != LogSchemaVersion {
		t.Errorf("SchemaVersion: want=%q got=%q", LogSchemaVersion, decoded.SchemaVersion)
	}
}

// TestEventTypeValues는 EventType 열거형 값이 예상대로 정의되어 있는지 확인한다.
func TestEventTypeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		et   EventType
		want string
	}{
		{EventTypeMoaiSubcommand, "moai_subcommand"},
		{EventTypeAgentInvocation, "agent_invocation"},
		{EventTypeSpecReference, "spec_reference"},
		{EventTypeFeedback, "feedback"},
	}

	for _, tc := range tests {
		if string(tc.et) != tc.want {
			t.Errorf("EventType %q: want string value %q", tc.et, tc.want)
		}
	}
}

// TestLogSchemaVersion은 상수 값이 "v1"인지 확인한다.
func TestLogSchemaVersion(t *testing.T) {
	t.Parallel()

	if LogSchemaVersion != "v1" {
		t.Errorf("LogSchemaVersion: want %q got %q", "v1", LogSchemaVersion)
	}
}

// ─────────────────────────────────────────────
// T-P1-02: RecordEvent 테스트
// ─────────────────────────────────────────────

// TestRecordEventWritesJSONL는 RecordEvent가 파일에 유효한 JSONL을 기록하는지 검증한다.
// REQ-HL-001: observer는 PostToolUse hook handler로 실행되며 이벤트당 <100ms.
func TestRecordEventWritesJSONL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"

	obs := NewObserver(logPath)

	if err := obs.RecordEvent(EventTypeMoaiSubcommand, "/moai plan", "hash001"); err != nil {
		t.Fatalf("RecordEvent 실패: %v", err)
	}
	if err := obs.RecordEvent(EventTypeAgentInvocation, "expert-backend", "hash002"); err != nil {
		t.Fatalf("RecordEvent 두 번째 호출 실패: %v", err)
	}

	// 파일이 생성되었는지 확인
	data, err := readFileBytes(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}

	// JSONL 파싱: 각 줄이 유효한 JSON인지 확인
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 2 {
		t.Errorf("기록된 줄 수: want=2 got=%d", len(lines))
	}

	for i, line := range lines {
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("줄 %d json 파싱 실패: %v", i, err)
		}
		if evt.SchemaVersion != LogSchemaVersion {
			t.Errorf("줄 %d: SchemaVersion want=%q got=%q", i, LogSchemaVersion, evt.SchemaVersion)
		}
	}
}

// TestRecordEvent100Sequential는 100회 연속 RecordEvent가 각각 100ms 이내에 완료되는지 검증한다.
// REQ-HL-001: observer는 부모 tool call을 블록하지 않아야 한다.
func TestRecordEvent100Sequential(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obs := NewObserver(dir + "/usage-log.jsonl")

	const count = 100
	// 첫 호출들은 OS 파일 캐시를 워밍업한다 (Windows/race detector 환경에서
	// 첫 write가 antivirus/file system caching 영향으로 늦을 수 있음).
	// 워밍업 이후의 안정 상태에서만 100ms 한도를 검증한다.
	const warmup = 5
	limit := 100 * time.Millisecond

	for i := range count {
		start := time.Now()
		if err := obs.RecordEvent(EventTypeMoaiSubcommand, "test-subject", "hash"); err != nil {
			t.Fatalf("RecordEvent %d번째 실패: %v", i, err)
		}
		if i < warmup {
			continue
		}
		elapsed := time.Since(start)
		if elapsed > limit {
			t.Errorf("RecordEvent %d번째: %v > %v (100ms 한도 초과)", i, elapsed, limit)
		}
	}
}

// TestRecordEventAppends는 기존 파일에 새 이벤트가 추가(append)되는지 검증한다.
func TestRecordEventAppends(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"
	obs := NewObserver(logPath)

	for range 5 {
		if err := obs.RecordEvent(EventTypeSpecReference, "SPEC-001", "h"); err != nil {
			t.Fatalf("RecordEvent 실패: %v", err)
		}
	}

	data, _ := readFileBytes(logPath)
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 5 {
		t.Errorf("append 후 줄 수: want=5 got=%d", len(lines))
	}
}

// ─────────────────────────────────────────────
// T-P1-04: PruneStaleEntries 테스트
// ─────────────────────────────────────────────

// TestPruneStaleEntriesRemovesOldEvents는 retentionDays보다 오래된 이벤트가
// 제거되고 아카이브에 추가되는지 검증한다. REQ-HL-011.
func TestPruneStaleEntriesRemovesOldEvents(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"

	// 테스트용 이벤트: 2개는 오래된 것, 1개는 최신
	now := time.Now().UTC()
	old1 := Event{
		Timestamp:     now.AddDate(0, 0, -10), // 10일 전
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "old1",
		ContextHash:   "h1",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	old2 := Event{
		Timestamp:     now.AddDate(0, 0, -8), // 8일 전
		EventType:     EventTypeAgentInvocation,
		Subject:       "old2",
		ContextHash:   "h2",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	fresh := Event{
		Timestamp:     now.AddDate(0, 0, -1), // 1일 전 (신선)
		EventType:     EventTypeFeedback,
		Subject:       "fresh",
		ContextHash:   "h3",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}

	// 이벤트를 파일에 직접 기록
	if err := writeEventsToFile(logPath, []Event{old1, old2, fresh}); err != nil {
		t.Fatalf("테스트 데이터 기록 실패: %v", err)
	}

	archiveDir := dir + "/archive"
	retention := NewRetention(logPath, archiveDir, func() time.Time { return now })

	if err := retention.PruneStaleEntries(7); err != nil { // 7일 보존
		t.Fatalf("PruneStaleEntries 실패: %v", err)
	}

	// 로그 파일에는 fresh 이벤트만 남아야 한다
	data, err := readFileBytes(logPath)
	if err != nil {
		t.Fatalf("로그 파일 읽기 실패: %v", err)
	}
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 1 {
		t.Errorf("prune 후 줄 수: want=1 got=%d", len(lines))
	}
	if len(lines) == 1 {
		var evt Event
		if err := json.Unmarshal([]byte(lines[0]), &evt); err != nil {
			t.Errorf("남은 이벤트 파싱 실패: %v", err)
		}
		if evt.Subject != "fresh" {
			t.Errorf("남은 이벤트 subject: want=fresh got=%q", evt.Subject)
		}
	}

	// 아카이브 파일이 생성되었는지 확인 (gzip)
	archiveFiles := listFilesInDir(archiveDir)
	if len(archiveFiles) == 0 {
		t.Error("아카이브 파일이 생성되지 않았습니다")
	}
}

// TestPruneSkipsIfRecentlyPruned는 마지막 prune으로부터 1시간 이내면 skip하는지 검증한다.
func TestPruneSkipsIfRecentlyPruned(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := dir + "/usage-log.jsonl"

	now := time.Now().UTC()
	// 오래된 이벤트 1개 기록
	old := Event{
		Timestamp:     now.AddDate(0, 0, -10),
		EventType:     EventTypeMoaiSubcommand,
		Subject:       "old",
		ContextHash:   "h",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}
	if err := writeEventsToFile(logPath, []Event{old}); err != nil {
		t.Fatalf("테스트 데이터 기록 실패: %v", err)
	}

	archiveDir := dir + "/archive"

	// 첫 번째 prune: 성공해야 함
	retention := NewRetention(logPath, archiveDir, func() time.Time { return now })
	if err := retention.PruneStaleEntries(7); err != nil {
		t.Fatalf("첫 번째 PruneStaleEntries 실패: %v", err)
	}

	// 이벤트 다시 추가 (오래된 것)
	if err := writeEventsToFile(logPath, []Event{old}); err != nil {
		t.Fatalf("재기록 실패: %v", err)
	}

	// 두 번째 prune: 1시간 이내이므로 skip해야 한다
	// (동일한 retention 인스턴스 재사용 — lastPruneAt이 설정되어 있음)
	if err := retention.PruneStaleEntries(7); err != nil {
		t.Fatalf("두 번째 PruneStaleEntries 실패: %v", err)
	}

	// 이벤트가 아직 파일에 있어야 한다 (pruning이 skip되었으므로)
	data, _ := readFileBytes(logPath)
	lines := splitNonEmptyLines(string(data))
	if len(lines) != 1 {
		t.Errorf("skip 후 줄 수: want=1 got=%d (pruning이 skip되지 않음)", len(lines))
	}
}

// ─────────────────────────────────────────────
// 헬퍼 함수
// ─────────────────────────────────────────────

// readFileBytes는 파일 내용을 바이트로 읽는다.
func readFileBytes(path string) ([]byte, error) {
	// os는 harness 패키지에서 이미 임포트됨 (observer.go에서)
	return readFile(path)
}

// splitNonEmptyLines는 문자열을 개행 문자로 분리하고 빈 줄을 제거한다.
func splitNonEmptyLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		if line := s[start:]; line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// writeEventsToFile는 이벤트 슬라이스를 JSONL 형식으로 파일에 기록한다.
func writeEventsToFile(path string, events []Event) error {
	return appendEventsJSONL(path, events)
}

// listFilesInDir는 디렉토리 내 파일 목록을 반환한다.
func listFilesInDir(dir string) []string {
	return listDir(dir)
}
