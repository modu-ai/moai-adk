package search

import (
	"database/sql"
	"fmt"
	"strings"
)

// SearchOptions는 검색 쿼리 옵션이다.
type SearchOptions struct {
	Query  string // FTS5 전문 검색 쿼리 (3자 이상) 또는 LIKE 쿼리 (1~2자)
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
	Score       float64 // BM25 점수 (작을수록 관련성 높음). LIKE 폴백 시 0.0
}

// defaultLimit는 Limit가 지정되지 않았을 때 사용하는 기본 최대 결과 수이다.
const defaultLimit = 20

// fts5MinQueryRunes는 FTS5 trigram 토크나이저가 요구하는 최소 쿼리 길이(Unicode 문자 수)이다.
// 이보다 짧은 쿼리(한/중/일 2자 단어 포함)는 LIKE 폴백을 사용한다.
const fts5MinQueryRunes = 3

// Search는 인덱싱된 세션에서 메시지를 검색한다.
//
// 쿼리 길이에 따라 검색 방식이 다르다:
//   - 3자 이상: FTS5 MATCH (BM25 점수 기반 정렬, 고성능)
//   - 1~2자:    LIKE '%query%' 폴백 (한국어/일본어/중국어 2자 단어 지원)
//
// 결과가 없으면 빈 슬라이스(nil 아님)를 반환한다.
func Search(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = defaultLimit
	}

	queryRunes := []rune(opts.Query)
	if len(queryRunes) >= fts5MinQueryRunes {
		return searchFTS5(db, opts)
	}
	return searchLike(db, opts)
}

// searchFTS5는 FTS5 MATCH를 사용하여 BM25 점수 기반으로 검색한다.
// trigram 토크나이저 특성상 쿼리는 3 Unicode 문자 이상이어야 한다.
func searchFTS5(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	var conditions []string
	var args []any

	conditions = append(conditions, "messages MATCH ?")
	args = append(args, opts.Query)

	appendCommonFilters(&conditions, &args, opts)
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

	return runSearchQuery(db, query, args)
}

// searchLike는 LIKE 패턴 매칭을 사용하여 검색한다.
// FTS5 trigram이 지원하지 않는 1~2자 CJK 쿼리(한국어 "인증", 일본어 "認証", 중국어 "认证")를 처리한다.
// BM25 점수는 제공되지 않으며 Score는 0.0으로 고정된다.
func searchLike(db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	var conditions []string
	var args []any

	// LIKE 패턴: SQLite에서 '%' || ? || '%' 형식으로 파라미터 바인딩
	conditions = append(conditions, "m.text LIKE '%' || ? || '%'")
	args = append(args, opts.Query)

	appendCommonFilters(&conditions, &args, opts)
	args = append(args, opts.Limit)

	query := fmt.Sprintf(`
		SELECT m.session_id, m.role,
		       m.text AS excerpt,
		       m.timestamp, s.git_branch, s.project_path,
		       0.0 AS score
		FROM messages m
		JOIN sessions s ON m.session_id = s.session_id
		WHERE %s
		ORDER BY m.timestamp DESC
		LIMIT ?`,
		strings.Join(conditions, " AND "),
	)

	return runSearchQuery(db, query, args)
}

// appendCommonFilters는 브랜치, 역할, 날짜 범위 필터를 조건 슬라이스에 추가한다.
func appendCommonFilters(conditions *[]string, args *[]any, opts SearchOptions) {
	if opts.Branch != "" {
		*conditions = append(*conditions, "s.git_branch = ?")
		*args = append(*args, opts.Branch)
	}
	if opts.Role != "" {
		*conditions = append(*conditions, "m.role = ?")
		*args = append(*args, opts.Role)
	}
	if opts.Since != "" {
		*conditions = append(*conditions, "m.timestamp >= ?")
		*args = append(*args, opts.Since)
	}
	if opts.Until != "" {
		*conditions = append(*conditions, "m.timestamp <= ?")
		*args = append(*args, opts.Until)
	}
}

// runSearchQuery는 SQL 쿼리를 실행하고 SearchResult 슬라이스를 반환한다.
func runSearchQuery(db *sql.DB, query string, args []any) ([]SearchResult, error) {
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
		// Excerpt 최대 200자 제한 ([]rune 기반으로 멀티바이트 문자 안전 처리)
		if runes := []rune(r.Excerpt); len(runes) > 200 {
			r.Excerpt = string(runes[:200])
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 반복 실패: %w", err)
	}

	return results, nil
}
