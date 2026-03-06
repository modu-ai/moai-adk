//go:build !race

package search

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
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
// If the JSONL file does not exist, the error is silently skipped.
// Other parse errors are returned to the caller.
// Sessions with zero parsed messages are not indexed.
func IndexSession(db *sql.DB, sessionID, filePath, gitBranch, projectPath string) error {
	// Parse JSONL file.
	messages, err := ParseJSONL(filePath, gitBranch, projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing file: silently skip.
			slog.Warn("search: JSONL file not found, skipping indexing",
				"session_id", sessionID,
				"path", filePath,
			)
			return nil
		}
		return fmt.Errorf("failed to parse JSONL (session=%s): %w", sessionID, err)
	}

	// Skip sessions with no indexable messages.
	if len(messages) == 0 {
		slog.Debug("search: no messages parsed, skipping indexing", "session_id", sessionID)
		return nil
	}

	// Atomic insert via transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Idempotency check inside the transaction to prevent TOCTOU races.
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = ?", sessionID).Scan(&count); err != nil {
		return fmt.Errorf("failed to check session index: %w", err)
	}
	if count > 0 {
		slog.Debug("search: skipping already-indexed session", "session_id", sessionID)
		return nil
	}

	// Insert session metadata using INSERT OR IGNORE for additional safety.
	_, err = tx.Exec(
		`INSERT OR IGNORE INTO sessions (session_id, project_path, git_branch, file_path) VALUES (?, ?, ?, ?)`,
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
