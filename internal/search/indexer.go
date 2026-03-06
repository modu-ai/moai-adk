//go:build !race

package search

import (
	"database/sql"
	"fmt"
	"log/slog"
)

// IsIndexed reports whether the given session ID has already been indexed.
func IsIndexed(db *sql.DB, sessionID string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = ?", sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check session index: %w", err)
	}
	return count > 0, nil
}

// IndexSession parses a single JSONL session file and indexes it into the SQLite DB.
// Already-indexed sessions are skipped (idempotent).
// If the JSONL file does not exist, a warning is logged and nil is returned.
func IndexSession(db *sql.DB, sessionID, filePath, gitBranch, projectPath string) error {
	// Skip if already indexed.
	indexed, err := IsIndexed(db, sessionID)
	if err != nil {
		return err
	}
	if indexed {
		slog.Debug("search: skipping already-indexed session", "session_id", sessionID)
		return nil
	}

	// Parse JSONL file.
	messages, err := ParseJSONL(filePath, gitBranch, projectPath)
	if err != nil {
		// Log a warning for missing files and return without error.
		slog.Warn("search: failed to parse JSONL, skipping indexing",
			"session_id", sessionID,
			"path", filePath,
			"error", err,
		)
		return nil
	}

	// Atomic insert via transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Insert session metadata into sessions table.
	_, err = tx.Exec(
		`INSERT INTO sessions (session_id, project_path, git_branch, file_path) VALUES (?, ?, ?, ?)`,
		sessionID, projectPath, gitBranch, filePath,
	)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	// Insert individual messages into the FTS5 messages table.
	for _, msg := range messages {
		_, err = tx.Exec(
			`INSERT INTO messages (session_id, role, timestamp, text) VALUES (?, ?, ?, ?)`,
			msg.SessionID, msg.Role, msg.Timestamp, msg.Text,
		)
		if err != nil {
			return fmt.Errorf("failed to insert message (session=%s): %w", sessionID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Debug("search: session indexed successfully",
		"session_id", sessionID,
		"messages", len(messages),
	)
	return nil
}
