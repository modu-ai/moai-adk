package search

import (
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpenDB_CreatesDirectory(t *testing.T) {
	// Setup: Create a temp directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "subdir", "test.db")

	// Act: Open database (should create directory)
	db, err := OpenDB(dbPath)

	// Assert: No error and directory exists
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	if db == nil {
		t.Fatal("OpenDB returned nil database")
	}
	defer db.Close()

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Database directory was not created")
	}
}

func TestOpenDB_WALMode(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Act
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	// Assert: Check WAL mode is enabled
	var journalMode string
	row := db.QueryRow("PRAGMA journal_mode;")
	if err := row.Scan(&journalMode); err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected journal mode 'wal', got '%s'", journalMode)
	}
}

func TestOpenDB_ReturnsExisting(t *testing.T) {
	// Setup: Create database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Act 1: Open database first time
	db1, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("First OpenDB failed: %v", err)
	}
	defer db1.Close()

	// Create a test table to verify it's the same database
	_, err = db1.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY);")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Close first connection
	if err := db1.Close(); err != nil {
		t.Fatalf("Failed to close first database: %v", err)
	}

	// Act 2: Open database second time
	db2, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Second OpenDB failed: %v", err)
	}
	defer db2.Close()

	// Assert: Verify we can query the test table (proves it's the same database)
	var count int
	row := db2.QueryRow("SELECT COUNT(*) FROM test_table;")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("Failed to query test table: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 rows in test table, got %d", count)
	}
}

func TestOpenDB_EmptyPath(t *testing.T) {
	// Act: Try to open with empty path
	db, err := OpenDB("")

	// Assert: Should fail
	if err == nil {
		t.Error("Expected error for empty path, got nil")
		if db != nil {
			db.Close()
		}
	}
}
