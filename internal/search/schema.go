//go:build !race

package search

// createSessionsTableSQL is the DDL for the sessions metadata table.
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

// createMessagesTableSQL is the DDL for the FTS5 full-text search virtual table.
// Uses the trigram tokenizer to support CJK (Korean/Chinese/Japanese) text search.
const createMessagesTableSQL = `
CREATE VIRTUAL TABLE IF NOT EXISTS messages USING fts5(
    session_id UNINDEXED,
    role       UNINDEXED,
    timestamp  UNINDEXED,
    text,
    tokenize = 'trigram'
);
`
