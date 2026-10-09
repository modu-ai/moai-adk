package atomicfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// previousBootFixture writes a lock file whose recorded owner ran on a
// PREVIOUS boot (a boot id different from this machine's wire form) — a
// verified-dead owner by the boot-comparison path, whatever lives at the
// pid today. The current pid keeps the record self-consistent otherwise.
func previousBootFixture(t *testing.T, path string) {
	t.Helper()
	identity := BootIDIdentity()
	if identity == "" {
		t.Skip("no boot identity on this platform; the boot-comparison path is not exercised")
	}
	owner := LockOwner{PID: os.Getpid(), BootID: "previous-boot-" + identity}
	raw, err := json.Marshal(owner)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

// liveOwnerFixture seeds path with THIS process's owner record — a
// verifiably LIVE owner (same pid, same boot). The guard-walk tests use it
// for rivals whose disposal must be refused, never broken.
func liveOwnerFixture(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatalf("seed %s: %v", path, err)
	}
	if werr := writeOwnerLabel(path, 0o600); werr != nil {
		t.Fatalf("label %s: %v", path, werr)
	}
}

// TestBreakAbortsWhenTheFileChangedUnderneath is the DETERMINISTIC form of
// review-gate finding #6: a reclaimer that verified a stale snapshot must
// not remove whatever sits at the lock path NOW. The repro drives
// BreakStaleLock against a snapshot of the stale bytes while the file at
// the path has already been replaced by a fresh labelled lock — exactly the
// A-claimed-B-removed interleaving the concurrent repro widens for.
func TestBreakAbortsWhenTheFileChangedUnderneath(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")

	// (1) The stale record the reclaimer verifies.
	previousBootFixture(t, lockPath)
	staleBytes, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read stale fixture: %v", err)
	}
	var staleOwner LockOwner
	if err := json.Unmarshal(staleBytes, &staleOwner); err != nil {
		t.Fatalf("decode stale: %v", err)
	}
	if !OwnerIsDead(staleOwner) {
		t.Fatal("fixture: the previous-boot owner must verify dead")
	}

	// (2) Meanwhile, another reclaimer already reclaimed and acquired: the
	// path now holds a FRESH, labelled, live lock.
	fresh := LockOwner{PID: os.Getpid(), BootID: BootIDIdentity()}
	freshBytes, err := json.Marshal(fresh)
	if err != nil {
		t.Fatalf("marshal fresh: %v", err)
	}
	if err := os.WriteFile(lockPath, freshBytes, 0o600); err != nil {
		t.Fatalf("write fresh lock: %v", err)
	}

	// (3) The stale-armed reclaimer acts. It must abort: the bytes at the
	// path are not the ones it verified.
	if BreakStaleLock(lockPath) {
		t.Fatal("BreakStaleLock removed a lock whose bytes had changed under it")
	}
	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("the fresh lock was destroyed: %v", err)
	}
	if string(after) != string(freshBytes) {
		t.Fatalf("the fresh lock's bytes were altered: %s", after)
	}
}

// TestBreakStillFiresOnUnchangedStaleLock pins the other half: the re-read
// abort must not disarm the break against the lock it verified.
func TestBreakStillFiresOnUnchangedStaleLock(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")
	previousBootFixture(t, lockPath)
	if !BreakStaleLock(lockPath) {
		t.Fatal("BreakStaleLock did not fire on an unchanged verified-dead lock")
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("the stale lock survived the break: %v", err)
	}
}

// TestBreakGateAbortsWhenContentSwapsBeforeRemove drives the re-read gate
// with the removal seam: between the verdict's re-read and the remove, the
// file at the path is swapped for a fresh labelled lock — the exact
// interleaving of review-gate finding #6 (a second reclaimer already
// acquired; the stale-armed reclaimer must not delete their lock). The gate
// must abort and leave the fresh lock byte-identical.
func TestBreakGateAbortsWhenContentSwapsBeforeRemove(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "queue.lock")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	previousBootFixture(t, lockPath)

	freshBytes := []byte(`{"pid":` + strconv.Itoa(os.Getpid()) + `,"boot_id":"` + BootIDIdentity() + `"}`)
	// The injection point is the VERDICT read: it returns the stale bytes
	// the break will verify dead, and — modelling the second reclaimer's
	// re-acquire landing between that verdict and the gate's re-read —
	// replaces the file with a fresh labelled lock on the way out. The gate
	// re-read must see the change and abort.
	prevReread := sectionRereadFn
	t.Cleanup(func() { sectionRereadFn = prevReread })
	firstRead := true
	sectionRereadFn = func(path string) ([]byte, error) {
		raw, err := prevReread(path)
		if firstRead {
			firstRead = false
			if werr := os.WriteFile(path, freshBytes, 0o600); werr != nil {
				return nil, werr
			}
		}
		return raw, err
	}

	if BreakStaleLock(lockPath) {
		t.Fatal("the break fired although the file changed between the verdict read and the gate re-read")
	}
	after, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read after aborted break: %v", err)
	}
	if string(after) != string(freshBytes) {
		t.Fatalf("the fresh lock's bytes were altered: %s", after)
	}
}
