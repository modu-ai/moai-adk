//go:build !race

package search

// createSessionsTableSQL은 세션 메타데이터 테이블 DDL이다.
const createSessionsTableSQL = `
CREATE TABLE IF NOT EXISTS sessions (
    session_id   TEXT PRIMARY KEY,
    project_path TEXT NOT NULL DEFAULT '',
    git_branch   TEXT NOT NULL DEFAULT '',
    indexed_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    file_path    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_path);
CREATE INDEX IF NOT EXISTS idx_sessions_branch  ON sessions(git_branch);
`

// createMessagesTableSQL은 FTS5 전문 검색 가상 테이블 DDL이다.
// trigram 토크나이저를 사용하여 CJK(한/중/일) 텍스트 검색을 지원한다.
const createMessagesTableSQL = `
CREATE VIRTUAL TABLE IF NOT EXISTS messages USING fts5(
    session_id UNINDEXED,
    role       UNINDEXED,
    timestamp  UNINDEXED,
    text,
    tokenize = 'trigram'
);
`
