//go:build !race
// This file and the entire internal/search package are excluded from race-detector
// builds. modernc.org/sqlite (~70K lines of transpiled C) causes OOM when compiled
// with ThreadSanitizer instrumentation on memory-constrained CI runners (Windows,
// macOS). All dependents (indexer, parser, searcher, schema) are excluded too because
// they reference functions (OpenDB, CreateTables) defined here.
// Impact: search feature is not race-tested in CI, but production builds are unaffected.

package search

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// Pure-Go SQLite driver (no CGO required); includes FTS5 with trigram tokenizer.
	_ "modernc.org/sqlite"
)

// OpenDB opens a SQLite database at the given path and returns the connection.
// It creates parent directories if they do not exist, and enables WAL journal mode
// for improved concurrent read/write performance.
func OpenDB(dbPath string) (*sql.DB, error) {
	// Create parent directory if it does not exist.
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("failed to create DB directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite DB: %w", err)
	}

	// Enable WAL mode: improves concurrent read performance and reduces write contention.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	return db, nil
}

// CreateTables creates the sessions metadata table and the FTS5 messages virtual table.
// It is idempotent: if the tables already exist, no action is taken.
func CreateTables(db *sql.DB) error {
	if _, err := db.Exec(createSessionsTableSQL); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}
	if _, err := db.Exec(createMessagesTableSQL); err != nil {
		return fmt.Errorf("failed to create messages FTS5 table: %w", err)
	}
	return nil
}
