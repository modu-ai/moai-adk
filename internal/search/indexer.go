//go:build !race

package search

import (
	"database/sql"
	"fmt"
	"log/slog"
)

// IsIndexed는 주어진 세션 ID가 이미 인덱싱되어 있는지 확인한다.
func IsIndexed(db *sql.DB, sessionID string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = ?", sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("세션 인덱스 확인 실패: %w", err)
	}
	return count > 0, nil
}

// IndexSession은 단일 JSONL 세션 파일을 파싱하여 SQLite DB에 인덱싱한다.
// 이미 인덱싱된 세션은 건너뛴다 (멱등성 보장).
// JSONL 파일이 존재하지 않으면 경고만 기록하고 오류 없이 반환한다.
func IndexSession(db *sql.DB, sessionID, filePath, gitBranch, projectPath string) error {
	// 이미 인덱싱된 경우 스킵
	indexed, err := IsIndexed(db, sessionID)
	if err != nil {
		return err
	}
	if indexed {
		slog.Debug("search: 이미 인덱싱된 세션 스킵", "session_id", sessionID)
		return nil
	}

	// JSONL 파싱
	messages, err := ParseJSONL(filePath, gitBranch, projectPath)
	if err != nil {
		// 파일 없음은 경고만 기록하고 정상 반환
		slog.Warn("search: JSONL 파싱 실패, 인덱싱 스킵",
			"session_id", sessionID,
			"path", filePath,
			"error", err,
		)
		return nil
	}

	// 트랜잭션으로 원자적 삽입
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// sessions 테이블에 세션 메타데이터 삽입
	_, err = tx.Exec(
		`INSERT INTO sessions (session_id, project_path, git_branch, file_path) VALUES (?, ?, ?, ?)`,
		sessionID, projectPath, gitBranch, filePath,
	)
	if err != nil {
		return fmt.Errorf("세션 삽입 실패: %w", err)
	}

	// messages FTS5 테이블에 개별 메시지 삽입
	for _, msg := range messages {
		_, err = tx.Exec(
			`INSERT INTO messages (session_id, role, timestamp, text) VALUES (?, ?, ?, ?)`,
			msg.SessionID, msg.Role, msg.Timestamp, msg.Text,
		)
		if err != nil {
			return fmt.Errorf("메시지 삽입 실패 (session=%s): %w", sessionID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	slog.Debug("search: 세션 인덱싱 완료",
		"session_id", sessionID,
		"messages", len(messages),
	)
	return nil
}
