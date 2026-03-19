package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- Reader 테스트 ---

func TestNewReader(t *testing.T) {
	t.Parallel()

	r := NewReader()
	if r == nil {
		t.Fatal("NewReader가 nil을 반환함")
	}
}

// TestReader_ReadAll_EmptyFile: 비어 있는 파일에서 읽으면 빈 슬라이스를 반환하는지 검증한다.
func TestReader_ReadAll_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// 빈 파일 생성
	path := filepath.Join(dir, "journal.jsonl")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("빈 파일 생성 실패: %v", err)
	}

	r := NewReader()
	entries, err := r.ReadAll(dir)
	if err != nil {
		t.Fatalf("ReadAll 오류: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("항목 수 = %d, want 0", len(entries))
	}
}

// TestReader_ReadAll_NonexistentFile: 파일이 없으면 오류 없이 nil을 반환하는지 검증한다.
func TestReader_ReadAll_NonexistentFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // journal.jsonl 파일 없음
	r := NewReader()

	entries, err := r.ReadAll(dir)
	if err != nil {
		t.Fatalf("존재하지 않는 파일에서 ReadAll 오류: %v", err)
	}
	if entries != nil {
		t.Errorf("entries = %v, want nil", entries)
	}
}

// TestReader_ReadAll_MalformedJSONSkipped: 잘못된 JSON 줄은 건너뛰고 나머지는 읽는지 검증한다.
func TestReader_ReadAll_MalformedJSONSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "journal.jsonl")

	content := `{"id":"e1","session_id":"s1","ts":"2025-01-01T00:00:00Z","type":"session_start","status":"in_progress"}
THIS IS NOT JSON
{"id":"e2","session_id":"s1","ts":"2025-01-01T00:01:00Z","type":"session_end","status":"completed"}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}

	r := NewReader()
	entries, err := r.ReadAll(dir)
	if err != nil {
		t.Fatalf("ReadAll 오류: %v", err)
	}

	// 유효한 2개만 반환되어야 함
	if len(entries) != 2 {
		t.Errorf("항목 수 = %d, want 2 (malformed 줄 건너뜀)", len(entries))
	}
	if entries[0].ID != "e1" {
		t.Errorf("entries[0].ID = %q, want e1", entries[0].ID)
	}
	if entries[1].ID != "e2" {
		t.Errorf("entries[1].ID = %q, want e2", entries[1].ID)
	}
}

// TestReader_ReadLast_Normal: 마지막 N개 항목을 올바르게 반환하는지 검증한다.
func TestReader_ReadLast_Normal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	// 5개 항목 기록
	for i := 1; i <= 5; i++ {
		_ = w.Write(Entry{
			ID:        fmt.Sprintf("e%d", i),
			SessionID: "sess-1",
			Type:      "checkpoint",
			Status:    "in_progress",
		})
	}

	r := NewReader()
	entries, err := r.ReadLast(dir, 3)
	if err != nil {
		t.Fatalf("ReadLast 오류: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("항목 수 = %d, want 3", len(entries))
	}
	// 마지막 3개: e3, e4, e5
	if entries[0].ID != "e3" {
		t.Errorf("entries[0].ID = %q, want e3", entries[0].ID)
	}
	if entries[2].ID != "e5" {
		t.Errorf("entries[2].ID = %q, want e5", entries[2].ID)
	}
}

// TestReader_ReadLast_NGreaterThanTotal: N이 전체 항목 수보다 크면 모두 반환하는지 검증한다.
func TestReader_ReadLast_NGreaterThanTotal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	for i := 1; i <= 3; i++ {
		_ = w.Write(Entry{
			ID:        fmt.Sprintf("e%d", i),
			SessionID: "sess-1",
			Type:      "checkpoint",
			Status:    "in_progress",
		})
	}

	r := NewReader()
	entries, err := r.ReadLast(dir, 100) // N > 실제 항목 수
	if err != nil {
		t.Fatalf("ReadLast 오류: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("항목 수 = %d, want 3 (N > total이면 전체 반환)", len(entries))
	}
}

// TestReader_ReadLast_ExactN: N이 정확히 전체 항목 수와 같으면 모두 반환하는지 검증한다.
func TestReader_ReadLast_ExactN(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	for i := 1; i <= 4; i++ {
		_ = w.Write(Entry{
			ID:        fmt.Sprintf("e%d", i),
			SessionID: "sess-1",
			Type:      "checkpoint",
			Status:    "in_progress",
		})
	}

	r := NewReader()
	entries, err := r.ReadLast(dir, 4)
	if err != nil {
		t.Fatalf("ReadLast 오류: %v", err)
	}

	if len(entries) != 4 {
		t.Errorf("항목 수 = %d, want 4", len(entries))
	}
}

// --- BuildResumeContext 테스트 ---

// TestBuildResumeContext_NormalSession: 정상적으로 완료된 세션의 컨텍스트를 검증한다.
func TestBuildResumeContext_NormalSession(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	_ = w.LogSessionStart("sess-1", "SPEC-001", "run")
	_ = w.LogPhaseBegin("sess-1", "SPEC-001", "run", "구현 시작")
	_ = w.LogSessionEnd("sess-1", "SPEC-001", "run", "completed", 3000)

	r := NewReader()
	ctx, err := r.BuildResumeContext(dir)
	if err != nil {
		t.Fatalf("BuildResumeContext 오류: %v", err)
	}

	if ctx.SpecID != "SPEC-001" {
		t.Errorf("SpecID = %q, want SPEC-001", ctx.SpecID)
	}
	if ctx.LastSessionID != "sess-1" {
		t.Errorf("LastSessionID = %q, want sess-1", ctx.LastSessionID)
	}
	if ctx.LastPhase != "run" {
		t.Errorf("LastPhase = %q, want run", ctx.LastPhase)
	}
	if ctx.LastStatus != "completed" {
		t.Errorf("LastStatus = %q, want completed", ctx.LastStatus)
	}
	if ctx.Resumable {
		t.Error("완료된 세션은 Resumable이 false여야 함")
	}
	if ctx.SessionCount != 1 {
		t.Errorf("SessionCount = %d, want 1", ctx.SessionCount)
	}
	if ctx.TokensUsed != 3000 {
		t.Errorf("TokensUsed = %d, want 3000", ctx.TokensUsed)
	}
}

// TestBuildResumeContext_InterruptedSession: 중단된 세션은 Resumable이 true인지 검증한다.
func TestBuildResumeContext_InterruptedSession(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	_ = w.LogSessionStart("sess-1", "SPEC-002", "run")
	_ = w.LogSessionEnd("sess-1", "SPEC-002", "run", "token_limit", 180000)

	r := NewReader()
	ctx, err := r.BuildResumeContext(dir)
	if err != nil {
		t.Fatalf("BuildResumeContext 오류: %v", err)
	}

	if !ctx.Resumable {
		t.Error("중단된 세션은 Resumable이 true여야 함")
	}
	if ctx.EndReason != "token_limit" {
		t.Errorf("EndReason = %q, want token_limit", ctx.EndReason)
	}
}

// TestBuildResumeContext_CrashedSession: session_end 없이 세션이 끝나면 crash로 판정되는지 검증한다.
func TestBuildResumeContext_CrashedSession(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	_ = w.LogSessionStart("sess-crash", "SPEC-003", "plan")
	// session_end 없이 종료 (크래시 시뮬레이션)

	r := NewReader()
	ctx, err := r.BuildResumeContext(dir)
	if err != nil {
		t.Fatalf("BuildResumeContext 오류: %v", err)
	}

	if !ctx.Resumable {
		t.Error("크래시 세션은 Resumable이 true여야 함")
	}
	if ctx.EndReason != "crash" {
		t.Errorf("EndReason = %q, want crash", ctx.EndReason)
	}
	if ctx.LastSessionID != "sess-crash" {
		t.Errorf("LastSessionID = %q, want sess-crash", ctx.LastSessionID)
	}
	if ctx.LastPhase != "plan" {
		t.Errorf("LastPhase = %q, want plan", ctx.LastPhase)
	}
}

// TestBuildResumeContext_MultipleSessions: 여러 세션이 있을 때 가장 마지막 세션 상태를 반환하는지 검증한다.
func TestBuildResumeContext_MultipleSessions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	// 세션 1: 완료
	_ = w.LogSessionStart("sess-1", "SPEC-001", "plan")
	_ = w.LogSessionEnd("sess-1", "SPEC-001", "plan", "completed", 1000)

	// 세션 2: 중단
	_ = w.LogSessionStart("sess-2", "SPEC-001", "run")
	_ = w.LogSessionEnd("sess-2", "SPEC-001", "run", "context_limit", 190000)

	r := NewReader()
	ctx, err := r.BuildResumeContext(dir)
	if err != nil {
		t.Fatalf("BuildResumeContext 오류: %v", err)
	}

	if ctx.SessionCount != 2 {
		t.Errorf("SessionCount = %d, want 2", ctx.SessionCount)
	}
	// 가장 마지막 세션(sess-2)의 상태를 반영해야 함
	if ctx.LastSessionID != "sess-2" {
		t.Errorf("LastSessionID = %q, want sess-2", ctx.LastSessionID)
	}
	if !ctx.Resumable {
		t.Error("마지막 세션이 중단되었으므로 Resumable이 true여야 함")
	}
}

// TestBuildResumeContext_EmptyEntries: 항목이 없으면 오류를 반환하는지 검증한다.
func TestBuildResumeContext_EmptyEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // journal.jsonl 없음
	r := NewReader()

	_, err := r.BuildResumeContext(dir)
	if err == nil {
		t.Error("빈 journal에서 BuildResumeContext는 오류를 반환해야 함")
	}
}

// TestBuildResumeContext_CheckpointNextAction: checkpoint의 next_step이 NextAction에 저장되는지 검증한다.
func TestBuildResumeContext_CheckpointNextAction(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewWriter(dir)

	_ = w.LogSessionStart("sess-1", "SPEC-001", "run")
	_ = w.LogCheckpoint("sess-1", "SPEC-001", "run", map[string]string{
		"next_step":      "write_tests",
		"files_modified": "internal/service/user.go",
	})
	// session_end 없이 종료 (크래시)

	r := NewReader()
	ctx, err := r.BuildResumeContext(dir)
	if err != nil {
		t.Fatalf("BuildResumeContext 오류: %v", err)
	}

	if ctx.NextAction != "write_tests" {
		t.Errorf("NextAction = %q, want write_tests", ctx.NextAction)
	}
	if len(ctx.FilesModified) == 0 || ctx.FilesModified[0] != "internal/service/user.go" {
		t.Errorf("FilesModified = %v, want [internal/service/user.go]", ctx.FilesModified)
	}
}

// --- FindResumableSPECs 테스트 ---

// TestFindResumableSPECs_FindsResumable: 재개 가능한 SPEC들을 찾는지 검증한다.
func TestFindResumableSPECs_FindsResumable(t *testing.T) {
	t.Parallel()

	specsDir := t.TempDir()

	// SPEC-001: 중단됨 (재개 가능)
	spec1Dir := filepath.Join(specsDir, "SPEC-001")
	_ = os.MkdirAll(spec1Dir, 0o755)
	w1 := NewWriter(spec1Dir)
	_ = w1.LogSessionStart("sess-1", "SPEC-001", "run")
	_ = w1.LogSessionEnd("sess-1", "SPEC-001", "run", "token_limit", 180000)

	// SPEC-002: 완료됨 (재개 불필요)
	spec2Dir := filepath.Join(specsDir, "SPEC-002")
	_ = os.MkdirAll(spec2Dir, 0o755)
	w2 := NewWriter(spec2Dir)
	_ = w2.LogSessionStart("sess-2", "SPEC-002", "run")
	_ = w2.LogSessionEnd("sess-2", "SPEC-002", "run", "completed", 5000)

	// SPEC-003: 크래시 (재개 가능)
	spec3Dir := filepath.Join(specsDir, "SPEC-003")
	_ = os.MkdirAll(spec3Dir, 0o755)
	w3 := NewWriter(spec3Dir)
	_ = w3.LogSessionStart("sess-3", "SPEC-003", "plan")
	// session_end 없음

	r := NewReader()
	results, err := r.FindResumableSPECs(specsDir)
	if err != nil {
		t.Fatalf("FindResumableSPECs 오류: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("재개 가능 SPEC 수 = %d, want 2", len(results))
	}

	// 재개 가능한 결과만 포함되어야 함
	for _, ctx := range results {
		if !ctx.Resumable {
			t.Errorf("Resumable이 false인 항목이 포함됨: %+v", ctx)
		}
	}
}

// TestFindResumableSPECs_IgnoresCompleted: 완료된 SPEC은 결과에서 제외되는지 검증한다.
func TestFindResumableSPECs_IgnoresCompleted(t *testing.T) {
	t.Parallel()

	specsDir := t.TempDir()

	// 모두 완료된 SPEC
	for i := 1; i <= 3; i++ {
		specDir := filepath.Join(specsDir, fmt.Sprintf("SPEC-%03d", i))
		_ = os.MkdirAll(specDir, 0o755)
		w := NewWriter(specDir)
		_ = w.LogSessionStart(fmt.Sprintf("sess-%d", i), fmt.Sprintf("SPEC-%03d", i), "run")
		_ = w.LogSessionEnd(fmt.Sprintf("sess-%d", i), fmt.Sprintf("SPEC-%03d", i), "run", "completed", 1000)
	}

	r := NewReader()
	results, err := r.FindResumableSPECs(specsDir)
	if err != nil {
		t.Fatalf("FindResumableSPECs 오류: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("재개 가능 SPEC 수 = %d, want 0 (모두 완료)", len(results))
	}
}

// TestFindResumableSPECs_NonexistentDir: 존재하지 않는 디렉터리에서 오류를 반환하는지 검증한다.
func TestFindResumableSPECs_NonexistentDir(t *testing.T) {
	t.Parallel()

	r := NewReader()
	_, err := r.FindResumableSPECs("/nonexistent/path/to/specs")
	if err == nil {
		t.Error("존재하지 않는 디렉터리에서 오류를 반환해야 함")
	}
}

// TestFindResumableSPECs_EmptyDir: 빈 디렉터리에서 빈 결과를 반환하는지 검증한다.
func TestFindResumableSPECs_EmptyDir(t *testing.T) {
	t.Parallel()

	specsDir := t.TempDir()
	r := NewReader()

	results, err := r.FindResumableSPECs(specsDir)
	if err != nil {
		t.Fatalf("빈 디렉터리에서 FindResumableSPECs 오류: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("항목 수 = %d, want 0", len(results))
	}
}

// TestFindResumableSPECs_IgnoresFiles: 파일은 무시하고 디렉터리만 처리하는지 검증한다.
func TestFindResumableSPECs_IgnoresFiles(t *testing.T) {
	t.Parallel()

	specsDir := t.TempDir()

	// 파일 하나 생성 (디렉터리가 아님)
	_ = os.WriteFile(filepath.Join(specsDir, "not-a-spec.txt"), []byte("data"), 0o644)

	// 실제 SPEC 디렉터리 (재개 가능)
	specDir := filepath.Join(specsDir, "SPEC-001")
	_ = os.MkdirAll(specDir, 0o755)
	w := NewWriter(specDir)
	_ = w.LogSessionStart("sess-1", "SPEC-001", "run")
	// 크래시 시뮬레이션

	r := NewReader()
	results, err := r.FindResumableSPECs(specsDir)
	if err != nil {
		t.Fatalf("FindResumableSPECs 오류: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("재개 가능 SPEC 수 = %d, want 1 (파일은 무시)", len(results))
	}
}
