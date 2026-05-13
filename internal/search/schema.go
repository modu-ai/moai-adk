package search

import (
	"database/sql"
	"fmt"
)

// CreateTables creates the sessions table and messages FTS5 virtual table.
// It's safe to call multiple times (idempotent).
func CreateTables(db *sql.DB) error {
	// Create sessions table
	sessionsDDL := `
	CREATE TABLE IF NOT EXISTS sessions (
		session_id   TEXT PRIMARY KEY,
		project_path TEXT NOT NULL DEFAULT '',
		git_branch   TEXT NOT NULL DEFAULT '',
		indexed_at   DATETIME NOT NULL DEFAULT (datetime('now')),
		file_path    TEXT NOT NULL DEFAULT ''
	);
	`
	if _, err := db.Exec(sessionsDDL); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create indexes for sessions table
	projectIndexDDL := `CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_path);`
	if _, err := db.Exec(projectIndexDDL); err != nil {
		return fmt.Errorf("failed to create project_path index: %w", err)
	}

	branchIndexDDL := `CREATE INDEX IF NOT EXISTS idx_sessions_branch ON sessions(git_branch);`
	if _, err := db.Exec(branchIndexDDL); err != nil {
		return fmt.Errorf("failed to create git_branch index: %w", err)
	}

	// Create messages FTS5 virtual table with trigram tokenizer for CJK support
	messagesDDL := `
	CREATE VIRTUAL TABLE IF NOT EXISTS messages USING fts5(
		session_id UNINDEXED,
		role       UNINDEXED,
		timestamp  UNINDEXED,
		text,
		tokenize = 'trigram'
	);
	`
	if _, err := db.Exec(messagesDDL); err != nil {
		return fmt.Errorf("failed to create messages FTS5 table: %w", err)
	}

	return nil
}
