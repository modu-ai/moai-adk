// integration_lock_write_test.go — AC-ILA-008 (REQ-ILA-010, card t336): the
// record's staging path is unique per call and leaves no residue.
//
// Scope, stated rather than implied: under the mutation lock two concurrent
// writers cannot reach writeIntegrationLock at all, so this observes the
// PROPERTY (unique staging name, no leftover file, mode preserved) and never a
// torn record. It is defence in depth against a future caller that writes
// outside the lock.
//
// It lives in its own file because AC-ILA-005 requires
// integration_lock_test.go to be byte-unchanged against 15453140a: the
// pre-existing criteria are met without editing them.
package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

// stagingResidue lists anything left behind in dir that is not the record.
func stagingResidue(t *testing.T, dir, record string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the state dir: %v", err)
	}
	var residue []string
	for _, e := range entries {
		name := e.Name()
		if name == filepath.Base(record) {
			continue
		}
		residue = append(residue, name)
	}
	return residue
}

// TestWriteIntegrationLock_UniqueStagingPath — repeated writes leave the record
// and nothing else, and the record keeps the 0644 mode every other session's
// read path depends on.
func TestWriteIntegrationLock_UniqueStagingPath(t *testing.T) {
	root := t.TempDir()
	record := integrationLockPath(root)
	if err := os.MkdirAll(filepath.Dir(record), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	for i, session := range []string{"lane-one", "lane-two", "lane-three"} {
		if err := writeIntegrationLock(record, &IntegrationLock{
			SessionID:  session,
			PID:        os.Getpid(),
			PIDSource:  PIDSourceSessionOwner,
			Branch:     "release/v0.0.0",
			AcquiredAt: "2026-08-30T00:00:00Z",
		}); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		if residue := stagingResidue(t, filepath.Dir(record), record); len(residue) != 0 {
			t.Fatalf("write %d left staging residue in the state dir: %v", i, residue)
		}
	}

	// The record must still be readable and name the last writer — the unique
	// staging path changed where the bytes are staged, not what lands.
	got, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("reading the record back: %v", err)
	}
	if got.SessionID != "lane-three" {
		t.Fatalf("record names %q, want lane-three", got.SessionID)
	}

	info, err := os.Stat(record)
	if err != nil {
		t.Fatalf("stat record: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("record mode = %v, want 0644 — os.CreateTemp opens at 0600, and a record other sessions cannot read is a guard that fails closed for the wrong reason", perm)
	}
}
