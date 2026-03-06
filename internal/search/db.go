//go:build !race

package search

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// modernc.org/sqlite: CGO 없는 순수 Go SQLite 드라이버
	_ "modernc.org/sqlite"
)

// OpenDB는 주어진 경로에 SQLite 데이터베이스를 열고 반환한다.
// 경로의 상위 디렉터리가 없으면 자동으로 생성한다.
// WAL 저널 모드를 활성화하여 동시 읽기/쓰기 성능을 높인다.
func OpenDB(dbPath string) (*sql.DB, error) {
	// 상위 디렉터리 생성 (이미 존재하면 무시)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("DB 디렉터리 생성 실패: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("SQLite DB 오픈 실패: %w", err)
	}

	// WAL 모드 활성화: 동시 읽기 성능 향상 및 쓰기 충돌 감소
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("WAL 모드 설정 실패: %w", err)
	}

	return db, nil
}

// CreateTables는 sessions 테이블과 messages FTS5 가상 테이블을 생성한다.
// 이미 테이블이 존재하면 아무 작업도 수행하지 않는다 (멱등성 보장).
func CreateTables(db *sql.DB) error {
	if _, err := db.Exec(createSessionsTableSQL); err != nil {
		return fmt.Errorf("sessions 테이블 생성 실패: %w", err)
	}
	if _, err := db.Exec(createMessagesTableSQL); err != nil {
		return fmt.Errorf("messages FTS5 테이블 생성 실패: %w", err)
	}
	return nil
}
