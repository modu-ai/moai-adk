package factorymsg

import (
	"testing"
	"time"
)

// TestOpenPathsSetSynchronousNormal pins the synchronous(NORMAL) connection
// pragma on both broker open paths. Card t1592 measured the hook's 2s bind
// budget dying inside the fresh-DB schema DDL on windows-latest — per-commit
// FULL fsyncs on a cold filesystem — and NORMAL is the fix's load-bearing
// pragma (WAL + NORMAL is crash-safe and drops the per-commit flush). A future
// edit that loses the pragma from either DSN fails here through the real open
// path, not by reading the DSN string.
func TestOpenPathsSetSynchronousNormal(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()

	store, err := OpenWithDeadline(root, "syncpin", 5*time.Second)
	if err != nil {
		t.Fatalf("init-path open: %v", err)
	}
	var mode int
	if err := store.db.QueryRow("PRAGMA synchronous").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if mode != 1 { // 1 = NORMAL (0 OFF, 2 FULL)
		t.Fatalf("init-path synchronous = %d, want 1 (NORMAL)", mode)
	}

	existing, err := OpenExistingWithDeadline(root, "syncpin", time.Second)
	if err != nil {
		t.Fatalf("existing-path open: %v", err)
	}
	defer func() { _ = existing.Close() }()
	if err := existing.db.QueryRow("PRAGMA synchronous").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != 1 {
		t.Fatalf("existing-path synchronous = %d, want 1 (NORMAL)", mode)
	}
}
