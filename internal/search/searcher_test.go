package search_test

import (
	"database/sql"
	"testing"

	"github.com/modu-ai/moai-adk/internal/search"
)

// setupSearchDB는 테스트용 DB를 생성하고 데이터를 삽입한다.
// FTS5 trigram 토크나이저는 최소 3 Unicode 문자 쿼리가 필요하므로
// 검색어는 반드시 3자 이상의 한국어 단어를 사용한다.
func setupSearchDB(t *testing.T) *sql.DB {
	t.Helper()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	// 테스트 세션 삽입
	sessions := []struct {
		id       string
		project  string
		branch   string
		filePath string
	}{
		{"sess-main-1", "/project/alpha", "main", "/path/sess-main-1.jsonl"},
		{"sess-main-2", "/project/alpha", "main", "/path/sess-main-2.jsonl"},
		{"sess-feat-1", "/project/alpha", "feat/auth", "/path/sess-feat-1.jsonl"},
	}

	for _, s := range sessions {
		_, err := db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
			VALUES (?, ?, ?, ?)`, s.id, s.project, s.branch, s.filePath)
		if err != nil {
			t.Fatalf("세션 삽입 실패: %v", err)
		}
	}

	// 테스트 메시지 삽입
	// 3자 이상 한국어 단어: 인덱스(3자), 플로우(3자), 최적화(3자), 방법론(3자)
	messages := []struct {
		sessionID string
		role      string
		timestamp string
		text      string
	}{
		{"sess-main-1", "user", "2026-03-01T10:00:00Z",
			"JWT 인증 토큰 구현 방법론을 알려주세요"},
		{"sess-main-1", "assistant", "2026-03-01T10:01:00Z",
			"JWT 토큰은 Header Payload Signature로 구성됩니다"},
		{"sess-main-2", "user", "2026-03-02T10:00:00Z",
			"데이터베이스 인덱스 최적화 전략"},
		{"sess-main-2", "assistant", "2026-03-02T10:01:00Z",
			"BTree 인덱스와 Hash 인덱스의 차이점을 설명하겠습니다"},
		{"sess-feat-1", "user", "2026-03-03T10:00:00Z",
			"OAuth2 인증 플로우 구현 방법론"},
		{"sess-feat-1", "assistant", "2026-03-03T10:01:00Z",
			"Authorization Code 플로우를 권장합니다"},
	}

	for _, m := range messages {
		_, err := db.Exec(`INSERT INTO messages (session_id, role, timestamp, text)
			VALUES (?, ?, ?, ?)`, m.sessionID, m.role, m.timestamp, m.text)
		if err != nil {
			t.Fatalf("메시지 삽입 실패: %v", err)
		}
	}

	return db
}

// TestSearch_BasicQuery는 기본 쿼리(3자 한국어)가 결과를 반환하는지 확인한다.
func TestSearch_BasicQuery(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	// "인덱스" = 인(1)+덱(1)+스(1) = 3자 한국어 (trigram 최소 요건 충족)
	results, err := search.Search(db, search.SearchOptions{
		Query: "인덱스",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	if len(results) == 0 {
		t.Error("결과 없음: '인덱스'에 매칭되는 메시지가 존재해야 함")
	}

	for _, r := range results {
		if r.SessionID == "" {
			t.Error("결과에 SessionID 없음")
		}
		if r.Role == "" {
			t.Error("결과에 Role 없음")
		}
	}
}

// TestSearch_BM25Ranking은 BM25 점수 기준으로 정렬되는지 확인한다.
func TestSearch_BM25Ranking(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	results, err := search.Search(db, search.SearchOptions{
		Query: "인덱스",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	if len(results) < 2 {
		t.Skip("랭킹 확인을 위해 최소 2개의 결과 필요")
	}

	// SQLite FTS5 rank는 음수 BM25: ORDER BY rank이므로 첫 번째가 가장 관련성 높음
	// Score 값이 오름차순이거나 동등해야 함
	for i := 1; i < len(results); i++ {
		if results[i-1].Score > results[i].Score {
			t.Errorf("정렬 순서 오류: 결과[%d].Score(%f) > 결과[%d].Score(%f)",
				i-1, results[i-1].Score, i, results[i].Score)
		}
	}
}

// TestSearch_BranchFilter는 브랜치 필터가 동작하는지 확인한다.
// "플로우" = 플(1)+로(1)+우(1) = 3자, feat/auth 브랜치에만 존재
func TestSearch_BranchFilter(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	results, err := search.Search(db, search.SearchOptions{
		Query:  "플로우",
		Branch: "feat/auth",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	if len(results) == 0 {
		t.Error("결과 없음: feat/auth 브랜치에 '플로우' 포함 메시지가 존재해야 함")
	}

	for _, r := range results {
		if r.GitBranch != "feat/auth" {
			t.Errorf("브랜치 필터 실패: 예상=feat/auth, 실제=%q", r.GitBranch)
		}
	}
}

// TestSearch_RoleFilter는 역할 필터가 동작하는지 확인한다.
func TestSearch_RoleFilter(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	results, err := search.Search(db, search.SearchOptions{
		Query: "인덱스",
		Role:  "user",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	for _, r := range results {
		if r.Role != "user" {
			t.Errorf("역할 필터 실패: 예상=user, 실제=%q", r.Role)
		}
	}
}

// TestSearch_DateFilter는 날짜 범위 필터가 동작하는지 확인한다.
// "플로우"는 2026-03-03 이후 메시지에만 있음
func TestSearch_DateFilter(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	results, err := search.Search(db, search.SearchOptions{
		Query: "플로우",
		Since: "2026-03-03T00:00:00Z",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	for _, r := range results {
		if r.Timestamp < "2026-03-03T00:00:00Z" {
			t.Errorf("날짜 필터 실패: timestamp=%q가 since보다 이전", r.Timestamp)
		}
	}
}

// TestSearch_LimitResults는 Limit 옵션이 결과 수를 제한하는지 확인한다.
func TestSearch_LimitResults(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	// "인덱스"는 2개의 메시지에 매칭됨; limit=1로 제한 검증
	results, err := search.Search(db, search.SearchOptions{
		Query: "인덱스",
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	if len(results) > 1 {
		t.Errorf("Limit 초과: 원함<=1, 실제=%d", len(results))
	}
}

// TestSearch_NoResults는 매칭이 없을 때 빈 슬라이스(nil 아님)를 반환하는지 확인한다.
func TestSearch_NoResults(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	results, err := search.Search(db, search.SearchOptions{
		Query: "xyzabcnotexist123",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search 실패: %v", err)
	}

	if results == nil {
		t.Error("결과 없을 때 nil 반환 (빈 슬라이스여야 함)")
	}
	if len(results) != 0 {
		t.Errorf("결과가 없어야 하는데 %d개 반환", len(results))
	}
}

// TestSearch_CJKText는 3자 한국어 텍스트를 trigram으로 검색하는지 확인한다.
// FTS5 trigram 최소 요건: 쿼리가 3자 이상이어야 함
// "최적화" = 최(1)+적(1)+화(1) = 3자 한국어 ✓
func TestSearch_CJKText(t *testing.T) {
	t.Parallel()

	db := setupSearchDB(t)

	// "최적화" = 3자 한국어 (FTS5 trigram 최소 요건 충족)
	results, err := search.Search(db, search.SearchOptions{
		Query: "최적화",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("CJK 검색 실패: %v", err)
	}

	if len(results) == 0 {
		t.Error("한국어 trigram 검색 결과 없음: '최적화' 매칭 실패")
	}
}

// TestSearch_Korean은 한국어 2자 쿼리가 LIKE 폴백으로 검색되는지 확인한다.
// "인증" = 인(1)+증(1) = 2자 → FTS5 trigram 미지원 → LIKE 폴백 사용
func TestSearch_Korean(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	// 한국어 2자 단어 "인증"이 포함된 세션 삽입
	_, err := db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
		VALUES ('ko-sess', '/project', 'main', '/path.jsonl')`)
	if err != nil {
		t.Fatalf("세션 삽입 실패: %v", err)
	}
	_, err = db.Exec(`INSERT INTO messages (session_id, role, timestamp, text)
		VALUES ('ko-sess', 'user', '2026-03-06T10:00:00Z', '인증 구현 방법을 알려주세요')`)
	if err != nil {
		t.Fatalf("메시지 삽입 실패: %v", err)
	}

	// "인증" (2자) → LIKE 폴백으로 검색
	results, err := search.Search(db, search.SearchOptions{
		Query: "인증",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("한국어 검색 실패: %v", err)
	}
	if len(results) == 0 {
		t.Error("한국어 2자 쿼리 '인증' 검색 결과 없음 (LIKE 폴백 미작동)")
	}
}

// TestSearch_English는 영어 4자 쿼리가 FTS5 trigram으로 검색되는지 확인한다.
// "auth" = 4자 ASCII → FTS5 trigram 직접 사용 가능
func TestSearch_English(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	_, err := db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
		VALUES ('en-sess', '/project', 'main', '/path.jsonl')`)
	if err != nil {
		t.Fatalf("세션 삽입 실패: %v", err)
	}
	_, err = db.Exec(`INSERT INTO messages (session_id, role, timestamp, text)
		VALUES ('en-sess', 'user', '2026-03-06T10:00:00Z',
		        'How to implement authentication with JWT tokens')`)
	if err != nil {
		t.Fatalf("메시지 삽입 실패: %v", err)
	}

	// "auth" (4자) → FTS5 trigram으로 검색
	results, err := search.Search(db, search.SearchOptions{
		Query: "auth",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("영어 검색 실패: %v", err)
	}
	if len(results) == 0 {
		t.Error("영어 쿼리 'auth' 검색 결과 없음")
	}
}

// TestSearch_Japanese는 일본어 2자 쿼리가 LIKE 폴백으로 검색되는지 확인한다.
// "認証" = 認(1)+証(1) = 2자 → FTS5 trigram 미지원 → LIKE 폴백 사용
func TestSearch_Japanese(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	_, err := db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
		VALUES ('ja-sess', '/project', 'main', '/path.jsonl')`)
	if err != nil {
		t.Fatalf("세션 삽입 실패: %v", err)
	}
	_, err = db.Exec(`INSERT INTO messages (session_id, role, timestamp, text)
		VALUES ('ja-sess', 'user', '2026-03-06T10:00:00Z',
		        '認証の実装方法を教えてください')`)
	if err != nil {
		t.Fatalf("메시지 삽입 실패: %v", err)
	}

	// "認証" (2자) → LIKE 폴백으로 검색
	results, err := search.Search(db, search.SearchOptions{
		Query: "認証",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("일본어 검색 실패: %v", err)
	}
	if len(results) == 0 {
		t.Error("일본어 2자 쿼리 '認証' 검색 결과 없음 (LIKE 폴백 미작동)")
	}
}

// TestSearch_Chinese는 중국어 2자 쿼리가 LIKE 폴백으로 검색되는지 확인한다.
// "认证" = 认(1)+证(1) = 2자 → FTS5 trigram 미지원 → LIKE 폴백 사용
func TestSearch_Chinese(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	if err := search.CreateTables(db); err != nil {
		t.Fatalf("CreateTables 실패: %v", err)
	}

	_, err := db.Exec(`INSERT INTO sessions (session_id, project_path, git_branch, file_path)
		VALUES ('zh-sess', '/project', 'main', '/path.jsonl')`)
	if err != nil {
		t.Fatalf("세션 삽입 실패: %v", err)
	}
	_, err = db.Exec(`INSERT INTO messages (session_id, role, timestamp, text)
		VALUES ('zh-sess', 'user', '2026-03-06T10:00:00Z',
		        '认证实现的最佳实践是什么')`)
	if err != nil {
		t.Fatalf("메시지 삽입 실패: %v", err)
	}

	// "认证" (2자) → LIKE 폴백으로 검색
	results, err := search.Search(db, search.SearchOptions{
		Query: "认证",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("중국어 검색 실패: %v", err)
	}
	if len(results) == 0 {
		t.Error("중국어 2자 쿼리 '认证' 검색 결과 없음 (LIKE 폴백 미작동)")
	}
}
