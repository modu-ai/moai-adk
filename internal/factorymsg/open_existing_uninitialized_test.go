package factorymsg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Card t2068 — the window between a first opener creating broker.db and its
// schema commit. OpenExisting stats the file, sees it, and reads peers: on the
// empty file that is "no such table: peers". The file is created by the first
// connection (before the schema statement runs), so the window is real; the
// reproduction pins it by creating the empty file by hand.
func TestOpenExistingToleratesABrokerFileWithoutSchemaYet(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	path, err := BrokerPath(root, "run-window")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := OpenExistingWithDeadline(root, "run-window", 5*time.Second)
	if err != nil {
		t.Fatalf("OpenExisting on a broker file whose schema is not committed yet: %v", err)
	}
	defer func() { _ = s.Close() }()
}
