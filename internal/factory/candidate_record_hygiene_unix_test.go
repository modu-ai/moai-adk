//go:build unix

// The candidate store's path-hygiene fixtures are unix-only (symlink and
// FIFO semantics; the B1 split rule, card t1478 M2 repairs).

package factory

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestCandidateRecordWriteConfinesSymlinkedStore pins the write-side
// confinement: a `.moai/state/candidate/<card>` swapped for a symlink to an
// external directory must REFUSE the write, never follow the link and
// overwrite an external <pinnedSHA>.json (card t1478 M2 repair, the
// archiveThenRemove ensureNoSymlinkPath precedent).
func TestCandidateRecordWriteConfinesSymlinkedStore(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	cardDir := filepath.Join(root, ".moai", "state", "candidate", "t9001")
	if err := os.MkdirAll(filepath.Dir(cardDir), 0o755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	if err := os.Symlink(external, cardDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	victim := filepath.Join(external, "abc123def456.json")
	if err := os.WriteFile(victim, []byte("do not touch"), 0o600); err != nil {
		t.Fatalf("seed victim: %v", err)
	}

	err := WriteCandidateRecord(root, CandidateRecord{
		CardID: "t9001", PinnedSHA: "abc123def456", CandidateSHA: "c", Verdict: CandidateVerdictPending, PushedAt: "t",
	})
	if err == nil {
		t.Fatal("want the write to refuse a symlinked card directory")
	}
	data, readErr := os.ReadFile(victim)
	if readErr != nil {
		t.Fatalf("victim unreadable: %v", readErr)
	}
	if string(data) != "do not touch" {
		t.Errorf("external victim overwritten with %q — the write followed the symlink", data)
	}
}

// TestCandidateRecordReadRefusesFifoWithoutBlocking pins the read-side
// non-regular refusal: a record path swapped for a FIFO must be refused
// without opening it — the former plain os.ReadFile parked inside the
// candidate mutation lock and wedged every later candidate call of the
// card (card t1478 M2 repair, the sectionRereadFn precedent).
func TestCandidateRecordReadRefusesFifoWithoutBlocking(t *testing.T) {
	root := t.TempDir()
	recordDir := filepath.Join(root, ".moai", "state", "candidate", "t9001")
	if err := os.MkdirAll(recordDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	fifo := filepath.Join(recordDir, "abc123def456.json")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("mkfifo unavailable on this platform: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := ReadCandidateRecord(root, "t9001", "abc123def456")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("want a refusal for a FIFO at the record path")
		}
		if errors.Is(err, ErrCandidateRecordAbsent) {
			t.Logf("refusal reads as absent: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ReadCandidateRecord blocked 3s on a FIFO at the record path — the read does not refuse non-regular files")
	}
}
