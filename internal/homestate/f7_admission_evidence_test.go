package homestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ReadMigrationMarker distinguishes its failure modes: unreadable marker,
// unparseable marker, and a marker that names a different migration or
// project.
func TestReadMigrationMarkerFailureModes(t *testing.T) {
	root := factorySandbox(t)
	path, err := MigrationBarrierPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadMigrationMarker(root); err == nil {
		t.Fatal("read missing marker: err = nil, want read failure")
	}

	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadMigrationMarker(root); err == nil {
		t.Fatal("read garbage marker: err = nil, want unmarshal failure")
	}

	// A well-formed marker for a different project is an identity mismatch,
	// not a successful read.
	foreign := `{"migration_id":"m","project_key":"somewhere-else","project_root":"/elsewhere"}`
	if err := os.WriteFile(path, []byte(foreign), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadMigrationMarker(root); err == nil || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("read foreign marker: err = %v, want identity mismatch", err)
	}

	if err := ClearMigrationMarker(root); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("marker still present after clear: %v", err)
	}
	if err := ClearMigrationMarker(root); err != nil {
		t.Fatalf("clear already-absent marker: %v", err)
	}
}

// WriteMigrationMarker fails when the barrier path's parent cannot hold it.
func TestWriteMigrationMarkerParentIsFile(t *testing.T) {
	root := factorySandbox(t)
	path, err := MigrationBarrierPath(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteMigrationMarker(root, []byte(`{}`)); err == nil || !strings.Contains(err.Error(), "mkdir") {
		t.Fatalf("write marker under file parent: err = %v, want MkdirAll (mkdir) failure", err)
	}
}

// readBoundedFile enforces the evidence-size bound and the regular-file
// requirement: an over-size file and a directory are both refused.
func TestReadBoundedFileRefusals(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big")
	if err := os.WriteFile(big, make([]byte, maxEvidenceFileSize+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBoundedFile(big); err == nil || !strings.Contains(err.Error(), "at most") {
		t.Fatalf("oversize file: err = %v, want bounded refusal", err)
	}
	if _, err := readBoundedFile(dir); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("directory: err = %v, want regular-file refusal", err)
	}
	small := filepath.Join(dir, "small")
	if err := os.WriteFile(small, []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	if b, err := readBoundedFile(small); err != nil || string(b) != "ok" {
		t.Fatalf("small file = %q err=%v, want passthrough", b, err)
	}
}
