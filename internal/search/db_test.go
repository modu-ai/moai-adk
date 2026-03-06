//go:build !race

package search_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/search"
)

// TestOpenDB_CreatesDirectory는 DB 경로의 상위 디렉터리를 자동 생성하는지 확인한다.
func TestOpenDB_CreatesDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "nested", "sessions.db")

	db, err := search.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB 실패: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("OpenDB가 상위 디렉터리를 생성하지 않았다")
	}
}

// TestOpenDB_WALMode는 WAL 저널 모드가 활성화되는지 확인한다.
func TestOpenDB_WALMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "sessions.db")

	db, err := search.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB 실패: %v", err)
	}
	defer func() { _ = db.Close() }()

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("journal_mode PRAGMA 조회 실패: %v", err)
	}
	if mode != "wal" {
		t.Errorf("WAL 모드가 아님: journal_mode = %q", mode)
	}
}

// TestOpenDB_ReturnsExisting는 기존 DB 파일을 열 수 있는지 확인한다.
func TestOpenDB_ReturnsExisting(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "sessions.db")

	// 첫 번째 오픈
	db1, err := search.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("첫 번째 OpenDB 실패: %v", err)
	}
	if err := search.CreateTables(db1); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}
	_ = db1.Close()

	// 두 번째 오픈 (기존 DB 재사용)
	db2, err := search.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("두 번째 OpenDB 실패: %v", err)
	}
	defer func() { _ = db2.Close() }()

	// 기존 테이블이 존재하는지 확인
	var tableName string
	err = db2.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&tableName)
	if err != nil {
		t.Fatalf("기존 테이블 조회 실패: %v", err)
	}
	if tableName != "sessions" {
		t.Errorf("기존 테이블을 찾지 못함: %q", tableName)
	}
}

// TestCreateTables_SessionsTable은 sessions 테이블이 올바르게 생성되는지 확인한다.
func TestCreateTables_SessionsTable(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	// sessions 테이블 존재 확인
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&count)
	if err != nil {
		t.Fatalf("테이블 조회 실패: %v", err)
	}
	if count != 1 {
		t.Error("sessions 테이블이 생성되지 않았다")
	}

	// PRIMARY KEY 삽입 테스트
	_, err = db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
		VALUES ('test-id', '/project', 'main', '/path/to/file.jsonl')`)
	if err != nil {
		t.Fatalf("sessions 행 삽입 실패: %v", err)
	}
}

// TestCreateTables_FTS5Table은 messages FTS5 가상 테이블이 생성되는지 확인한다.
func TestCreateTables_FTS5Table(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	// messages FTS5 테이블 존재 확인
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&count)
	if err != nil {
		t.Fatalf("FTS5 테이블 조회 실패: %v", err)
	}
	if count != 1 {
		t.Error("messages FTS5 테이블이 생성되지 않았다")
	}
}

// TestCreateTables_Idempotent는 CreateTables를 여러 번 호출해도 오류가 없는지 확인한다.
func TestCreateTables_Idempotent(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	for i := range 3 {
		if err := search.CreateTables(db); err != nil {
			t.Fatalf("CreateTables 호출 #%d 실패: %v", i+1, err)
		}
	}
}

// openTestDB는 테스트용 인메모리 SQLite DB를 열어 반환한다.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := search.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("테스트 DB 오픈 실패: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
