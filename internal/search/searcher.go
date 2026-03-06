package search

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchOptions는 검색 쿼리 옵션이다.
type SearchOptions struct {
	Query  string // FTS5 전문 검색 쿼리
	Branch string // git 브랜치 필터 (빈 문자열이면 전체)
	Since  string // 시작 날짜 필터 (RFC3339)
	Until  string // 종료 날짜 필터 (RFC3339)
	Role   string // 역할 필터: user|assistant (빈 문자열이면 전체)
	Limit  int    // 최대 결과 수 (0이면 기본값 20 사용)
}

// SearchResult는 단일 검색 결과를 나타낸다.
type SearchResult struct {
	SessionID   string
	Role        string
	Excerpt     string  // 최대 200자의 발췌 텍스트
	Timestamp   string
	GitBranch   string
	ProjectPath string
	Score       float64 // BM25 점수 (작을수록 관련성 높음)
}

// defaultLimit는 Limit가 지정되지 않았을 때 사용하는 기본 최대 결과 수이다.
const defaultLimit = 20

// Search는 SQLite FTS5를 사용하여 인덱싱된 세션에서 메시지를 검색한다.
// 결과는 BM25 점수 기준으로 정렬된다 (가장 관련성 높은 결과 먼저).
// 결과가 없으면 빈 슬라이스(nil 아님)를 반환한다.
func Search(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}

	// 동적 WHERE 절 구성
	var conditions []string
	var args []any

	// FTS5 MATCH 쿼리 (필수)
	conditions = append(conditions, "messages MATCH ?")
	args = append(args, opts.Query)

	if opts.Branch != "" {
		conditions = append(conditions, "s.git_branch = ?")
		args = append(args, opts.Branch)
	}
	if opts.Role != "" {
		conditions = append(conditions, "m.role = ?")
		args = append(args, opts.Role)
	}
	if opts.Since != "" {
		conditions = append(conditions, "m.timestamp >= ?")
		args = append(args, opts.Since)
	}
	if opts.Until != "" {
		conditions = append(conditions, "m.timestamp <= ?")
		args = append(args, opts.Until)
	}

	args = append(args, opts.Limit)

	query := fmt.Sprintf(`
		SELECT m.session_id, m.role,
		       snippet(messages, 3, '', '', '...', 64) AS excerpt,
		       m.timestamp, s.git_branch, s.project_path,
		       rank AS score
		FROM messages m
		JOIN sessions s ON m.session_id = s.session_id
		WHERE %s
		ORDER BY rank
		LIMIT ?`,
		strings.Join(conditions, " AND "),
	)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("검색 쿼리 실패: %w", err)
	}
	defer func() { _ = rows.Close() }()

	results := []SearchResult{}
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(
			&r.SessionID, &r.Role, &r.Excerpt,
			&r.Timestamp, &r.GitBranch, &r.ProjectPath,
			&r.Score,
		); err != nil {
			return nil, fmt.Errorf("결과 스캔 실패: %w", err)
		}
		// Excerpt 최대 200자 제한
		if len([]rune(r.Excerpt)) > 200 {
			r.Excerpt = string([]rune(r.Excerpt)[:200])
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 반복 실패: %w", err)
	}

	return results, nil
}
