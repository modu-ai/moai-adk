// m3_lockfile_classify_test.go — SPEC-USERASSET-DEPLOY-GUARD-001, gate
// round 20 #2: the .lock file takes no flock itself, so the guard file's
// flock-based absence judgment must not classify it — a pid-less leftover
// .lock is a genuine manual-recovery state (acquisition refuses with
// ErrLocked) on every platform.
package userassets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLockFilePidlessLeftoverIsOwnerlessNotAbsent(t *testing.T) {
	home := t.TempDir()
	path := LockPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("leftover without a pid record\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	state, pid := ClassifyLockFile(path)
	if state != GuardMarkerOwnerless || pid != 0 {
		t.Fatalf("a pid-less leftover .lock classified %s pid=%d — the flock-based absence judgment leaked onto the lock file (acquisition refuses here; the doctor must report the manual-recovery state)", state, pid)
	}

	// The mismatch the gate reproduced: classification vs acquisition.
	if _, err := AcquireUserLock(home, 200*time.Millisecond); err == nil {
		t.Fatal("acquisition succeeded over a pid-less leftover .lock — the pre-condition of this repro is gone")
	}
}
