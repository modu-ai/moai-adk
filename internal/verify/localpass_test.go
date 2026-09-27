package verify

import (
	"testing"
	"time"
)

// HasLocalPass (SPEC-CI-VERDICT-PRODUCER-001 REQ-CV-006): a head has a
// recorded local pass iff a snapshot whose key's head portion equals the
// head contains at least one check entry with exit code 0.
func TestHasLocalPass(t *testing.T) {
	head := "0123456789abcdef0123456789abcdef01234567"
	other := "ffffffffffffffffffffffffffffffffffffffff"

	// saveIn writes one exit-code-only snapshot into an explicit root.
	saveIn := func(t *testing.T, root, key string, exit int) {
		t.Helper()
		s := &Snapshot{Key: key, Checks: []CheckEntry{{
			CheckID: "test", Command: "go test ./...", ExitCode: exit, RecordedAt: time.Now(),
		}}}
		if err := Save(root, s); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	t.Run("exit-zero-snapshot", func(t *testing.T) {
		root := t.TempDir()
		saveIn(t, root, head+":d1", 0)
		if !HasLocalPass(root, head) {
			t.Error("exit-0 snapshot not recognized as local pass")
		}
	})
	t.Run("no-snapshot", func(t *testing.T) {
		if HasLocalPass(t.TempDir(), head) {
			t.Error("absent snapshot judged as local pass")
		}
	})
	t.Run("head-mismatch", func(t *testing.T) {
		root := t.TempDir()
		saveIn(t, root, other+":d1", 0)
		if HasLocalPass(root, head) {
			t.Error("foreign-head snapshot judged as local pass")
		}
	})
	t.Run("no-zero-exit-entry", func(t *testing.T) {
		root := t.TempDir()
		saveIn(t, root, head+":d1", 1)
		if HasLocalPass(root, head) {
			t.Error("all-failing snapshot judged as local pass")
		}
	})
	t.Run("second-entry-zero-counts", func(t *testing.T) {
		root := t.TempDir()
		saveIn(t, root, head+":d1", 1)
		saveIn(t, root, head+":d2", 0)
		if !HasLocalPass(root, head) {
			t.Error("any exit-0 snapshot at the head must count as local pass")
		}
	})
}
