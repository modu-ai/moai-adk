package journal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// --- Writer 테스트 ---

func TestNewWriter(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)
	if w == nil {
		t.Fatal("NewWriter가 nil을 반환함")
	}
	expected := filepath.Join(dir, "journal.jsonl")
	if w.path != expected {
		t.Errorf("path = %q, want %q", w.path, expected)
	}
}

func TestNewSessionWriter(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSessionWriter(dir)
	if w == nil {
		t.Fatal("NewSessionWriter가 nil을 반환함")
	}
	expected := filepath.Join(dir, "sessions", "journal.jsonl")
	if w.path != expected {
		t.Errorf("path = %q, want %q", w.path, expected)
	}
}

// TestWriter_Write_RoundTrip: Write 후 파일을 읽어 내용이 올바른지 검증한다.
func TestWriter_Write_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	entry := Entry{
		ID:        "test-id-1",
		SessionID: "sess-001",
		Type:      "session_start",
		SpecID:    "SPEC-001",
		Phase:     "run",
		Status:    "in_progress",
	}

	if err := w.Write(entry); err != nil {
		t.Fatalf("Write 실패: %v", err)
	}

	// 파일 읽기
	data, err := os.ReadFile(filepath.Join(dir, "journal.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	var got Entry
	if err := json.Unmarshal(data[:len(data)-1], &got); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if got.ID != entry.ID {
		t.Errorf("ID = %q, want %q", got.ID, entry.ID)
	}
	if got.SessionID != entry.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, entry.SessionID)
	}
	if got.Type != entry.Type {
		t.Errorf("Type = %q, want %q", got.Type, entry.Type)
	}
	if got.SpecID != entry.SpecID {
		t.Errorf("SpecID = %q, want %q", got.SpecID, entry.SpecID)
	}
}

// TestWriter_Write_TimestampAutoFill: Timestamp가 비어 있으면 자동으로 채워지는지 검증한다.
func TestWriter_Write_TimestampAutoFill(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	before := time.Now().UTC().Add(-time.Second)
	// Timestamp를 명시하지 않음
	entry := Entry{
		ID:        "ts-test",
		SessionID: "sess-ts",
		Type:      "checkpoint",
		Status:    "in_progress",
	}

	if err := w.Write(entry); err != nil {
		t.Fatalf("Write 실패: %v", err)
	}

	after := time.Now().UTC().Add(time.Second)

	data, err := os.ReadFile(filepath.Join(dir, "journal.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	var got Entry
	if err := json.Unmarshal(data[:len(data)-1], &got); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if got.Timestamp.IsZero() {
		t.Error("Timestamp가 자동으로 채워지지 않음")
	}
	if got.Timestamp.Before(before) || got.Timestamp.After(after) {
		t.Errorf("Timestamp %v가 예상 범위 [%v, %v] 밖에 있음", got.Timestamp, before, after)
	}
}

// TestWriter_Write_IDAutoGeneration: ID가 비어 있으면 자동으로 생성되는지 검증한다.
func TestWriter_Write_IDAutoGeneration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	entry := Entry{
		SessionID: "sess-id",
		Type:      "phase_begin",
		Status:    "in_progress",
	}

	if err := w.Write(entry); err != nil {
		t.Fatalf("Write 실패: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "journal.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	var got Entry
	if err := json.Unmarshal(data[:len(data)-1], &got); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if got.ID == "" {
		t.Error("ID가 자동으로 생성되지 않음")
	}
}

// TestWriter_Write_MultipleEntries: 여러 항목을 순서대로 쓰면 JSONL 형식으로 저장되는지 검증한다.
func TestWriter_Write_MultipleEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	entries := []Entry{
		{ID: "e1", SessionID: "s1", Type: "session_start", Status: "in_progress"},
		{ID: "e2", SessionID: "s1", Type: "phase_begin", Status: "in_progress"},
		{ID: "e3", SessionID: "s1", Type: "session_end", Status: "completed"},
	}

	for _, e := range entries {
		if err := w.Write(e); err != nil {
			t.Fatalf("Write 실패 (id=%s): %v", e.ID, err)
		}
	}

	// JSONL 파일의 줄 수 확인
	data, err := os.ReadFile(filepath.Join(dir, "journal.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	// 마지막 빈 줄을 제거하고 줄 수 계산
	lines := splitLines(data)
	if len(lines) != 3 {
		t.Errorf("줄 수 = %d, want 3", len(lines))
	}

	// 각 줄이 올바른 JSON인지 확인
	for i, line := range lines {
		var got Entry
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Errorf("줄 %d JSON 파싱 실패: %v", i+1, err)
		}
		if got.ID != entries[i].ID {
			t.Errorf("줄 %d ID = %q, want %q", i+1, got.ID, entries[i].ID)
		}
	}
}

// TestWriter_Write_CreatesDirectory: 디렉터리가 없으면 자동으로 생성되는지 검증한다.
func TestWriter_Write_CreatesDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// 중첩된 새 디렉터리 경로 사용
	specDir := filepath.Join(dir, "specs", "SPEC-999")
	w := NewWriter(specDir)

	entry := Entry{ID: "dir-test", SessionID: "s1", Type: "session_start", Status: "in_progress"}
	if err := w.Write(entry); err != nil {
		t.Fatalf("Write 실패: %v", err)
	}

	if _, err := os.Stat(filepath.Join(specDir, "journal.jsonl")); err != nil {
		t.Errorf("파일이 생성되지 않음: %v", err)
	}
}

// TestWriter_ConcurrentWrites: 동시에 여러 고루틴이 Write를 호출해도 안전한지 검증한다.
func TestWriter_ConcurrentWrites(t *testing.T) {
	t.Parallel()

	const goroutines = 20
	dir := t.TempDir()
	w := NewWriter(dir)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			entry := Entry{
				SessionID: "concurrent",
				Type:      "checkpoint",
				Status:    "in_progress",
			}
			if err := w.Write(entry); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("동시 쓰기 오류: %v", err)
	}

	// 모든 항목이 기록되었는지 확인
	data, err := os.ReadFile(filepath.Join(dir, "journal.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	lines := splitLines(data)
	if len(lines) != goroutines {
		t.Errorf("기록된 줄 수 = %d, want %d", len(lines), goroutines)
	}
}

// TestWriter_LogSessionStart: LogSessionStart가 올바른 항목을 기록하는지 검증한다.
func TestWriter_LogSessionStart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.LogSessionStart("sess-1", "SPEC-001", "run"); err != nil {
		t.Fatalf("LogSessionStart 실패: %v", err)
	}

	entries := readJSONL(t, filepath.Join(dir, "journal.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Type != "session_start" {
		t.Errorf("Type = %q, want session_start", e.Type)
	}
	if e.SessionID != "sess-1" {
		t.Errorf("SessionID = %q, want sess-1", e.SessionID)
	}
	if e.SpecID != "SPEC-001" {
		t.Errorf("SpecID = %q, want SPEC-001", e.SpecID)
	}
	if e.Phase != "run" {
		t.Errorf("Phase = %q, want run", e.Phase)
	}
	if e.Status != "in_progress" {
		t.Errorf("Status = %q, want in_progress", e.Status)
	}
}

// TestWriter_LogSessionEnd_Completed: reason이 "completed"이면 status도 "completed"인지 검증한다.
func TestWriter_LogSessionEnd_Completed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.LogSessionEnd("sess-1", "SPEC-001", "run", "completed", 5000); err != nil {
		t.Fatalf("LogSessionEnd 실패: %v", err)
	}

	entries := readJSONL(t, filepath.Join(dir, "journal.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Type != "session_end" {
		t.Errorf("Type = %q, want session_end", e.Type)
	}
	if e.Status != "completed" {
		t.Errorf("Status = %q, want completed", e.Status)
	}
	if e.TokensUsed != 5000 {
		t.Errorf("TokensUsed = %d, want 5000", e.TokensUsed)
	}
	if e.Context["reason"] != "completed" {
		t.Errorf("Context[reason] = %q, want completed", e.Context["reason"])
	}
}

// TestWriter_LogSessionEnd_Interrupted: reason이 "token_limit"이면 status가 "interrupted"인지 검증한다.
func TestWriter_LogSessionEnd_Interrupted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.LogSessionEnd("sess-2", "SPEC-002", "plan", "token_limit", 180000); err != nil {
		t.Fatalf("LogSessionEnd 실패: %v", err)
	}

	entries := readJSONL(t, filepath.Join(dir, "journal.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Status != "interrupted" {
		t.Errorf("Status = %q, want interrupted", e.Status)
	}
}

// TestWriter_LogPhaseBegin: description이 Context에 저장되는지 검증한다.
func TestWriter_LogPhaseBegin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	desc := "DDD 실행 시작"
	if err := w.LogPhaseBegin("sess-1", "SPEC-001", "run", desc); err != nil {
		t.Fatalf("LogPhaseBegin 실패: %v", err)
	}

	entries := readJSONL(t, filepath.Join(dir, "journal.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Type != "phase_begin" {
		t.Errorf("Type = %q, want phase_begin", e.Type)
	}
	if e.Context["description"] != desc {
		t.Errorf("Context[description] = %q, want %q", e.Context["description"], desc)
	}
}

// TestWriter_LogPhaseEnd: status가 올바르게 기록되는지 검증한다.
func TestWriter_LogPhaseEnd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.LogPhaseEnd("sess-1", "SPEC-001", "run", "completed"); err != nil {
		t.Fatalf("LogPhaseEnd 실패: %v", err)
	}

	entries := readJSONL(t, filepath.Join(dir, "journal.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Type != "phase_end" {
		t.Errorf("Type = %q, want phase_end", e.Type)
	}
	if e.Status != "completed" {
		t.Errorf("Status = %q, want completed", e.Status)
	}
}

// TestWriter_LogCheckpoint: Context가 올바르게 저장되는지 검증한다.
func TestWriter_LogCheckpoint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	ctx := map[string]string{
		"next_step":      "implement_service",
		"files_modified": "internal/service/user.go",
	}
	if err := w.LogCheckpoint("sess-1", "SPEC-001", "run", ctx); err != nil {
		t.Fatalf("LogCheckpoint 실패: %v", err)
	}

	entries := readJSONL(t, filepath.Join(dir, "journal.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Type != "checkpoint" {
		t.Errorf("Type = %q, want checkpoint", e.Type)
	}
	if e.Context["next_step"] != "implement_service" {
		t.Errorf("Context[next_step] = %q, want implement_service", e.Context["next_step"])
	}
}

// --- ReplayWriter 테스트 ---

func TestNewReplayWriter(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 5)
	if rw == nil {
		t.Fatal("NewReplayWriter가 nil을 반환함")
	}
	if rw.bufCap != 5 {
		t.Errorf("bufCap = %d, want 5", rw.bufCap)
	}
}

// TestNewReplayWriter_DefaultBufSize: bufSize <= 0이면 기본값 10이 사용되는지 검증한다.
func TestNewReplayWriter_DefaultBufSize(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	tests := []struct {
		name    string
		bufSize int
	}{
		{"zero", 0},
		{"negative", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rw := NewReplayWriter(dir, tt.bufSize)
			if rw.bufCap != 10 {
				t.Errorf("bufCap = %d, want 10 (기본값)", rw.bufCap)
			}
		})
	}
}

// TestReplayWriter_Flush_EmptyBuffer: 버퍼가 비어 있을 때 Flush를 호출해도 오류가 없는지 검증한다.
func TestReplayWriter_Flush_EmptyBuffer(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 5)

	if err := rw.Flush(); err != nil {
		t.Errorf("빈 버퍼에서 Flush 오류: %v", err)
	}

	// 파일이 생성되지 않아야 함
	if _, err := os.Stat(filepath.Join(dir, "replay.jsonl")); !os.IsNotExist(err) {
		t.Error("빈 Flush 후 파일이 생성됨 (예상하지 않음)")
	}
}

// TestReplayWriter_BufferFlushAtCapacity: 버퍼가 꽉 차면 자동으로 flush되는지 검증한다.
func TestReplayWriter_BufferFlushAtCapacity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 3)

	for i := 0; i < 3; i++ {
		entry := ActionEntry{
			SessionID: "sess-1",
			AgentName: "test-agent",
			Action:    "tool_use",
		}
		if err := rw.LogAction(entry); err != nil {
			t.Fatalf("LogAction 실패 (i=%d): %v", i, err)
		}
	}

	// 3번째 항목에서 자동 flush 발생 -> 파일에 3개 항목이 있어야 함
	entries := readReplayJSONL(t, filepath.Join(dir, "replay.jsonl"))
	if len(entries) != 3 {
		t.Errorf("항목 수 = %d, want 3", len(entries))
	}

	// 버퍼가 비워졌는지 확인
	rw.mu.Lock()
	bufLen := len(rw.buf)
	rw.mu.Unlock()

	if bufLen != 0 {
		t.Errorf("flush 후 버퍼 길이 = %d, want 0", bufLen)
	}
}

// TestReplayWriter_Flush_WritesBuffered: Flush가 버퍼에 쌓인 항목을 모두 기록하는지 검증한다.
func TestReplayWriter_Flush_WritesBuffered(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 10) // 10개 버퍼 - 자동 flush 없음

	for i := 0; i < 5; i++ {
		entry := ActionEntry{
			SessionID: "sess-1",
			AgentName: "agent",
			Action:    "start",
		}
		if err := rw.LogAction(entry); err != nil {
			t.Fatalf("LogAction 실패: %v", err)
		}
	}

	// 아직 파일이 없어야 함
	if _, err := os.Stat(filepath.Join(dir, "replay.jsonl")); !os.IsNotExist(err) {
		t.Error("Flush 전에 파일이 존재함")
	}

	if err := rw.Flush(); err != nil {
		t.Fatalf("Flush 실패: %v", err)
	}

	entries := readReplayJSONL(t, filepath.Join(dir, "replay.jsonl"))
	if len(entries) != 5 {
		t.Errorf("항목 수 = %d, want 5", len(entries))
	}
}

// TestReplayWriter_TimestampAutoFill: Timestamp가 비어 있으면 자동으로 채워지는지 검증한다.
func TestReplayWriter_TimestampAutoFill(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 1) // 버퍼 1개 -> 즉시 flush

	before := time.Now().UTC().Add(-time.Second)
	entry := ActionEntry{
		SessionID: "sess-ts",
		AgentName: "agent",
		Action:    "start",
	}
	if err := rw.LogAction(entry); err != nil {
		t.Fatalf("LogAction 실패: %v", err)
	}
	after := time.Now().UTC().Add(time.Second)

	entries := readReplayJSONL(t, filepath.Join(dir, "replay.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("항목 수 = %d, want 1", len(entries))
	}

	if entries[0].Timestamp.IsZero() {
		t.Error("Timestamp가 자동으로 채워지지 않음")
	}
	if entries[0].Timestamp.Before(before) || entries[0].Timestamp.After(after) {
		t.Errorf("Timestamp %v가 예상 범위 밖에 있음", entries[0].Timestamp)
	}
}

// TestReplayWriter_PartialWriteDedup: 에러 발생 시 이미 쓰인 항목이 버퍼에서 제거되는지 검증한다.
// flushLocked의 buf 슬라이싱 로직(written 커서)을 검증한다.
func TestReplayWriter_PartialWriteDedup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 10)

	// 정상 항목 3개를 버퍼에 추가
	for i := 0; i < 3; i++ {
		_ = rw.LogAction(ActionEntry{
			SessionID: "sess-1",
			AgentName: "agent",
			Action:    "tool_use",
		})
	}

	// 첫 번째 Flush 성공
	if err := rw.Flush(); err != nil {
		t.Fatalf("첫 번째 Flush 실패: %v", err)
	}

	// 두 번째 Flush (이미 비워진 버퍼) - 중복 없어야 함
	if err := rw.Flush(); err != nil {
		t.Fatalf("두 번째 Flush 실패: %v", err)
	}

	entries := readReplayJSONL(t, filepath.Join(dir, "replay.jsonl"))
	if len(entries) != 3 {
		t.Errorf("중복 flush 후 항목 수 = %d, want 3 (중복 없어야 함)", len(entries))
	}
}

// TestReplayWriter_MultipleFlush_Appends: 여러 번 Flush를 호출하면 파일에 누적되는지 검증한다.
func TestReplayWriter_MultipleFlush_Appends(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rw := NewReplayWriter(dir, 10)

	// 첫 배치
	for i := 0; i < 2; i++ {
		_ = rw.LogAction(ActionEntry{SessionID: "s1", Action: "start"})
	}
	if err := rw.Flush(); err != nil {
		t.Fatalf("첫 번째 Flush 실패: %v", err)
	}

	// 두 번째 배치
	for i := 0; i < 3; i++ {
		_ = rw.LogAction(ActionEntry{SessionID: "s1", Action: "tool_use"})
	}
	if err := rw.Flush(); err != nil {
		t.Fatalf("두 번째 Flush 실패: %v", err)
	}

	entries := readReplayJSONL(t, filepath.Join(dir, "replay.jsonl"))
	if len(entries) != 5 {
		t.Errorf("항목 수 = %d, want 5", len(entries))
	}
}

// --- 헬퍼 함수 ---

// splitLines는 JSONL 데이터를 줄 단위로 분리한다 (빈 줄 제외).
func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, string(data[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}

// readJSONL은 지정된 파일에서 Entry 목록을 읽는다.
func readJSONL(t *testing.T, path string) []Entry {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("파일 읽기 실패 (%s): %v", path, err)
	}
	lines := splitLines(data)
	entries := make([]Entry, 0, len(lines))
	for _, line := range lines {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("JSON 파싱 실패: %v", err)
		}
		entries = append(entries, e)
	}
	return entries
}

// readReplayJSONL은 지정된 파일에서 ActionEntry 목록을 읽는다.
func readReplayJSONL(t *testing.T, path string) []ActionEntry {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("파일 읽기 실패 (%s): %v", path, err)
	}
	lines := splitLines(data)
	entries := make([]ActionEntry, 0, len(lines))
	for _, line := range lines {
		var e ActionEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("JSON 파싱 실패: %v", err)
		}
		entries = append(entries, e)
	}
	return entries
}
