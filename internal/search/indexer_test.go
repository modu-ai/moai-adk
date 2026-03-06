//go:build !race

package search_test

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/search"
)

// TestIsIndexed_NotIndexed는 인덱싱되지 않은 세션에 대해 false를 반환하는지 확인한다.
func TestIsIndexed_NotIndexed(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	indexed, err := search.IsIndexed(db, "non-existent-session")
	if err != nil {
		t.Fatalf("IsIndexed 실패: %v", err)
	}
	if indexed {
		t.Error("존재하지 않는 세션에 대해 true 반환")
	}
}

// TestIsIndexed_AlreadyIndexed는 이미 인덱싱된 세션에 대해 true를 반환하는지 확인한다.
func TestIsIndexed_AlreadyIndexed(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	// 세션 직접 삽입
	_, err := db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
		VALUES ('existing-session', '/project', 'main', '/path/file.jsonl')`)
	if err != nil {
		t.Fatalf("세션 삽입 실패: %v", err)
	}

	indexed, err := search.IsIndexed(db, "existing-session")
	if err != nil {
		t.Fatalf("IsIndexed 실패: %v", err)
	}
	if !indexed {
		t.Error("인덱싱된 세션에 대해 false 반환")
	}
}

// TestIndexSession_NewSession은 새 세션을 인덱싱하는지 확인한다.
func TestIndexSession_NewSession(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	// 테스트용 JSONL 파일 생성
	jsonlContent := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"new-session","message":{"role":"user","content":[{"type":"text","text":"JWT 인증 구현 방법을 알고 싶습니다 상세하게 알려주세요"}]}}
`
	filePath := writeJSONL(t, tmpDir, "new-session.jsonl", jsonlContent)

	err := search.IndexSession(db, "new-session", filePath, "main", "/project")
	if err != nil {
		t.Fatalf("IndexSession 실패: %v", err)
	}

	// sessions 테이블 확인
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id='new-session'").Scan(&count); err != nil {
		t.Fatalf("세션 조회 실패: %v", err)
	}
	if count != 1 {
		t.Errorf("sessions 행 수 불일치: 원함=1, 실제=%d", count)
	}
}

// TestIndexSession_Idempotent는 같은 세션을 두 번 인덱싱해도 중복이 없는지 확인한다.
func TestIndexSession_Idempotent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	jsonlContent := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"dup-session","message":{"role":"user","content":[{"type":"text","text":"중복 인덱싱 테스트입니다 충분히 긴 텍스트"}]}}
`
	filePath := writeJSONL(t, tmpDir, "dup-session.jsonl", jsonlContent)

	// 두 번 인덱싱
	if err := search.IndexSession(db, "dup-session", filePath, "main", "/project"); err != nil {
		t.Fatalf("첫 번째 IndexSession 실패: %v", err)
	}
	if err := search.IndexSession(db, "dup-session", filePath, "main", "/project"); err != nil {
		t.Fatalf("두 번째 IndexSession 실패: %v", err)
	}

	// sessions 테이블에 1개만 존재해야 함
	var sessCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id='dup-session'").Scan(&sessCount); err != nil {
		t.Fatalf("세션 조회 실패: %v", err)
	}
	if sessCount != 1 {
		t.Errorf("중복 세션이 삽입됨: %d개", sessCount)
	}
}

// TestIndexSession_WithMessages는 메시지가 FTS5 테이블에 삽입되는지 확인한다.
func TestIndexSession_WithMessages(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	jsonlContent := `{"type":"user","timestamp":"2026-03-06T10:00:00Z","sessionId":"msg-session","message":{"role":"user","content":[{"type":"text","text":"첫 번째 사용자 메시지입니다 검색 테스트용"}]}}
{"type":"assistant","timestamp":"2026-03-06T10:01:00Z","sessionId":"msg-session","message":{"role":"assistant","content":[{"type":"text","text":"첫 번째 어시스턴트 응답입니다 검색 테스트용"}]}}
`
	filePath := writeJSONL(t, tmpDir, "msg-session.jsonl", jsonlContent)

	if err := search.IndexSession(db, "msg-session", filePath, "main", "/project"); err != nil {
		t.Fatalf("IndexSession 실패: %v", err)
	}

	// messages FTS5에 행이 삽입되었는지 확인
	var msgCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id='msg-session'").Scan(&msgCount); err != nil {
		t.Fatalf("메시지 조회 실패: %v", err)
	}
	if msgCount != 2 {
		t.Errorf("메시지 수 불일치: 원함=2, 실제=%d", msgCount)
	}

	// JSONL 파일 없어도 오류 없이 스킵되는지 확인
	if err := search.IndexSession(db, "no-file-session", "/nonexistent/path.jsonl", "main", "/project"); err != nil {
		t.Fatalf("존재하지 않는 파일에 대한 IndexSession이 오류 반환: %v", err)
	}

	// 파일이 없는 세션은 인덱싱되지 않아야 함
	indexed, _ := search.IsIndexed(db, "no-file-session")
	if indexed {
		t.Error("파일 없는 세션이 인덱싱됨")
	}

	_ = os.Remove(filePath)
}
