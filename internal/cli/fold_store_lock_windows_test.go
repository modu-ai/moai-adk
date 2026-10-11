//go:build windows

package cli

import "testing"

// TestFoldStoreLockWindowsSemantics is AC-MRR-008's executed-semantics
// test: on the REAL LockFileEx path — acquire → contending acquire refused
// → release → acquire succeeds, the same semantics the POSIX file executes
// on flock. The judge is the Windows leg of release-pr-multi-os.yml
// (acceptance.md AC-MRR-008): the only Windows surface that executes
// internal/cli root packages. GOOS=windows go build ./internal/cli/...
// remains the local compile gate.
func TestFoldStoreLockWindowsSemantics(t *testing.T) {
	dir := t.TempDir()

	held := newFoldStoreLock()
	if err := held.acquire(dir); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	// A contending non-blocking acquire is refused immediately, and the
	// bounded acquire refuses within the deadline — never an unbounded
	// block (REQ-MRR-003).
	contender := newFoldStoreLock()
	if err := contender.tryAcquire(dir); err == nil {
		_ = contender.release()
		t.Fatal("the contending non-blocking acquire succeeded while the store lock was held")
	}
	if err := contender.acquire(dir); err == nil {
		_ = contender.release()
		t.Fatal("the bounded acquire succeeded while the store lock was held")
	}

	// After release, acquisition succeeds again.
	if err := held.release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	again := newFoldStoreLock()
	if err := again.acquire(dir); err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	if err := again.release(); err != nil {
		t.Fatalf("release after re-acquire: %v", err)
	}
}
