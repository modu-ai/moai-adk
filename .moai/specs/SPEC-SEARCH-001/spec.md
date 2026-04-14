---
id: SPEC-SEARCH-001
title: "moai search: JSONL Session Search with SQLite FTS5"
status: planned
created: 2026-03-06
branch: feat/spec-search-001
tags: [search, sqlite, fts5, cli, hook]
---

# SPEC-SEARCH-001: moai search - JSONL 세션 검색 (SQLite FTS5)

## 개요

Claude Code 세션 JSONL 파일을 SQLite FTS5 데이터베이스에 인덱싱하고, `moai search` CLI 명령으로 전문 검색(Full-Text Search)을 제공하는 기능.

## 문제 정의

현재 MoAI-ADK는 758개 이상의 JSONL 세션 파일(`~/.claude/projects/{project-hash}/*.jsonl`)을 보유하고 있으나, 과거 세션 내용을 검색할 수 있는 체계적인 방법이 없다. git 기반 grep은 JSONL 파일 구조를 이해하지 못하며, 파일 수가 증가할수록 성능이 급격히 저하된다.

## 해결 방안

- `moai search "<query>"` CLI 명령으로 BM25 기반 전문 검색 제공
- `SessionEnd` 훅에서 비동기 서브프로세스로 자동 인덱싱
- SQLite FTS5 + trigram 토크나이저로 한국어/CJK 문자 검색 지원
- 글로벌 DB 경로: `~/.moai/search/sessions.db`

## EARS 요구사항

### REQ-SEARCH-001 (이벤트 기반)

**WHEN** `SessionEnd` 훅이 실행될 때, **THEN** 시스템은 완료된 세션의 JSONL 파일을 `~/.moai/search/sessions.db` SQLite FTS5 데이터베이스에 비동기적으로 인덱싱해야 한다.

### REQ-SEARCH-002 (상태 기반)

**IF** 검색 데이터베이스가 존재하면, **THEN** 시스템은 `moai search "<query>"`를 통해 BM25 랭킹 기반 전문 검색 쿼리를 지원해야 한다.

### REQ-SEARCH-003 (기능)

시스템은 `--branch`, `--since`, `--until`, `--role` 플래그로 검색 결과를 필터링하고, `--limit` (기본값 20)으로 결과 수를 제한해야 한다.

### REQ-SEARCH-004 (항상)

시스템은 **항상** SQLite trigram 토크나이저를 통해 한국어, 일본어, 중국어 및 기타 CJK 텍스트 검색을 형태소 분석 없이 지원해야 한다.

### REQ-SEARCH-005 (선택)

**가능하면** `--index-session` 플래그가 제공되면, 시스템은 검색 쿼리를 수행하지 않고 특정 세션 파일만 인덱싱해야 한다.

### REQ-SEARCH-006 (비허용)

시스템은 이미 인덱싱된 세션을 **다시 인덱싱하지 않아야 한다** (`session_id` primary key 기반 멱등성 보장).

## 아키텍처 개요

### 패키지 구조 (5개 패키지)

| 패키지 | 경로 | 책임 |
|--------|------|------|
| search/db | `internal/search/db.go` | DB 연결, WAL 모드, 테이블 생성 |
| search/schema | `internal/search/schema.go` | DDL 정의 (sessions 테이블 + FTS5 가상 테이블) |
| search/parser | `internal/search/parser.go` | JSONL 파싱, 노이즈 필터링 |
| search/indexer | `internal/search/indexer.go` | 세션 인덱싱, 중복 확인 |
| search/searcher | `internal/search/searcher.go` | FTS5 BM25 검색, 필터 적용 |
| cli/search | `internal/cli/search.go` | Cobra 서브커맨드 |

### 데이터베이스 스키마

```sql
CREATE TABLE IF NOT EXISTS sessions (
    session_id   TEXT PRIMARY KEY,
    project_path TEXT NOT NULL DEFAULT '',
    git_branch   TEXT NOT NULL DEFAULT '',
    indexed_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    file_path    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_path);
CREATE INDEX IF NOT EXISTS idx_sessions_branch  ON sessions(git_branch);

CREATE VIRTUAL TABLE IF NOT EXISTS messages USING fts5(
    session_id UNINDEXED,
    role       UNINDEXED,
    timestamp  UNINDEXED,
    text,
    tokenize = 'trigram'
);
```

### Go 인터페이스

```go
// internal/search/parser.go
type Message struct {
    SessionID   string
    Role        string
    Text        string
    Timestamp   string
    GitBranch   string
    ProjectPath string
}
func ParseJSONL(path, gitBranch, projectPath string) ([]Message, error)

// internal/search/db.go
func OpenDB(dbPath string) (*sql.DB, error)
func CreateTables(db *sql.DB) error

// internal/search/indexer.go
func IsIndexed(db *sql.DB, sessionID string) (bool, error)
func IndexSession(db *sql.DB, sessionID, filePath, gitBranch, projectPath string) error

// internal/search/searcher.go
type SearchOptions struct {
    Query, Branch, Since, Until, Role string
    Limit int
}
type SearchResult struct {
    SessionID, Role, Excerpt, Timestamp, GitBranch, ProjectPath string
    Score float64
}
func Search(db *sql.DB, opts SearchOptions) ([]SearchResult, error)
```

### 노이즈 필터링 (ParseJSONL)

다음 JSONL 레코드를 건너뜀:
- `type`이 `user` 또는 `assistant`가 아닌 경우
- `message.content`가 비어있거나 XML 태그만 포함하는 경우 (`<local-command-caveat>`, `<command-name>`, `<system-reminder>`, `<function_calls>`)
- 클리닝 후 텍스트 길이가 20자 미만인 경우

### 비동기 통합 패턴 (SessionEnd 훅)

```go
// internal/hook/session_end.go
func triggerSessionIndex(sessionID, projectDir, gitBranch string) {
    cmd := exec.Command("moai", "search", "--index-session", sessionID,
        "--project-path", projectDir, "--git-branch", gitBranch)
    if err := cmd.Start(); err != nil {
        slog.Warn("search: failed to start indexer subprocess", "error", err)
    }
    // cmd.Wait() 없음 - fire-and-forget
}
```

### JSONL 파일 위치

```
~/.claude/projects/{project-hash}/*.jsonl
```

`--index-session <id>` 제공 시, `~/.claude/projects/` 하위 디렉터리에서 `<id>.jsonl`을 탐색.

### 의존성

- `modernc.org/sqlite v1.35.0` - 순수 Go SQLite (CGO 불필요, FTS5 기본 활성화)

## 범위 외 (v1)

1. 시맨틱/벡터 검색 (키워드/부분 문자열 매칭만 지원)
2. 과거 세션 소급 벌크 임포트
3. DB 자동 vacuum (`moai search --vacuum` 미구현)
4. 크로스 프로젝트 검색 (DB는 글로벌 `~/.moai/search/`)
5. 검색 결과 하이라이팅
6. 100MB 이상 JSONL 파일 최적화

## 파일 변경 목록

### 신규 파일

- `internal/search/db.go` - OpenDB, WAL 모드
- `internal/search/schema.go` - CreateTables DDL
- `internal/search/parser.go` - ParseJSONL, 노이즈 필터링
- `internal/search/indexer.go` - IsIndexed, IndexSession
- `internal/search/searcher.go` - Search, BM25 + 필터
- `internal/search/db_test.go`
- `internal/search/parser_test.go`
- `internal/search/indexer_test.go`
- `internal/search/searcher_test.go`
- `internal/cli/search.go` - Cobra 서브커맨드
- `internal/cli/search_test.go`
- `internal/template/templates/.claude/skills/moai/workflows/search.md` - 스킬 정의

### 수정 파일

- `internal/defs/dirs.go` - `DirMoaiSearch` 추가
- `internal/defs/files.go` - `FileSearchDB` 추가
- `internal/hook/session_end.go` - `triggerSessionIndex()` 호출 추가
- `go.mod` - `modernc.org/sqlite v1.35.0` 추가
- `internal/cli/root.go` - `newSearchCmd()` 등록

## 추적성

- SPEC: SPEC-SEARCH-001
- Plan: `.moai/specs/SPEC-SEARCH-001/plan.md`
- Acceptance: `.moai/specs/SPEC-SEARCH-001/acceptance.md`
- Research: `.moai/specs/SPEC-SEARCH-001/research.md`
