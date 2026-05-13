package search

import (
	"testing"
)

func TestCreateTables_SessionsTable(t *testing.T) {
	// Setup: Create temp database
	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Act: Create tables
	if err := CreateTables(db); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	// Assert: Check sessions table exists
	var tableName string
	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions';")
	if err := row.Scan(&tableName); err != nil {
		t.Errorf("Sessions table does not exist: %v", err)
	}
	if tableName != "sessions" {
		t.Errorf("Expected table name 'sessions', got '%s'", tableName)
	}
}

func TestCreateTables_FTS5Table(t *testing.T) {
	// Setup: Create temp database
	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Act: Create tables
	if err := CreateTables(db); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	// Assert: Check messages FTS5 table exists
	var tableName string
	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='messages';")
	if err := row.Scan(&tableName); err != nil {
		t.Errorf("Messages FTS5 table does not exist: %v", err)
	}
	if tableName != "messages" {
		t.Errorf("Expected table name 'messages', got '%s'", tableName)
	}
}

func TestCreateTables_Idempotent(t *testing.T) {
	// Setup: Create temp database
	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Act: Create tables twice (should not error)
	if err := CreateTables(db); err != nil {
		t.Fatalf("First CreateTables failed: %v", err)
	}
	if err := CreateTables(db); err != nil {
		t.Errorf("Second CreateTables failed (not idempotent): %v", err)
	}
}

func TestCreateTables_SessionsIndexes(t *testing.T) {
	// Setup: Create temp database
	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Act: Create tables
	if err := CreateTables(db); err != nil {
		t.Fatalf("CreateTables failed: %v", err)
	}

	// Assert: Check indexes exist
	var indexCount int
	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND tbl_name='sessions';")
	if err := row.Scan(&indexCount); err != nil {
		t.Errorf("Failed to query indexes: %v", err)
	}

	// Should have 2 indexes: idx_sessions_project and idx_sessions_branch
	if indexCount < 2 {
		t.Errorf("Expected at least 2 indexes on sessions table, got %d", indexCount)
	}
}
