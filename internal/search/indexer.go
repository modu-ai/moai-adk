package search

import (
	"database/sql"
	"fmt"
)

// IsIndexed checks if a session has already been indexed in the database.
func IsIndexed(db *sql.DB, sessionID string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = ?", sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if session is indexed: %w", err)
	}
	return count > 0, nil
}

// IndexSession indexes a session's messages into the FTS5 database.
// It is idempotent - re-indexing the same session will not create duplicates.
func IndexSession(db *sql.DB, sessionID, filePath, gitBranch, projectPath string) error {
	// Check if already indexed
	indexed, err := IsIndexed(db, sessionID)
	if err != nil {
		return err
	}
	if indexed {
		return nil // Already indexed, skip
	}

	// Parse messages from JSONL file
	messages, err := ParseJSONL(filePath, gitBranch, projectPath)
	if err != nil {
		return fmt.Errorf("failed to parse JSONL: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert session metadata
	_, err = tx.Exec(
		"INSERT INTO sessions (session_id, project_path, git_branch, file_path) VALUES (?, ?, ?, ?)",
		sessionID, projectPath, gitBranch, filePath,
	)
	if err != nil {
		return fmt.Errorf("failed to insert session metadata: %w", err)
	}

	// Insert messages into FTS5 table
	stmt, err := tx.Prepare("INSERT INTO messages (session_id, role, timestamp, text) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, msg := range messages {
		msg.SessionID = sessionID // Set session ID from context
		_, err = stmt.Exec(msg.SessionID, msg.Role, msg.Timestamp, msg.Text)
		if err != nil {
			return fmt.Errorf("failed to insert message: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
