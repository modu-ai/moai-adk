package escalation

import (
	"path/filepath"
	"testing"
	"time"
)

// A per-card lock held by one hook makes a second hook of the same card fail
// with a timeout — a REQ-AE-004 fault — and the lock is taken again once
// released.
func TestLockCardTimeoutAndRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store", "t9001.lock")
	release, err := lockCard(path, time.Second)
	if err != nil {
		t.Fatalf("first lock: %v", err)
	}
	if _, err := lockCard(path, 50*time.Millisecond); err == nil {
		t.Fatal("second lock succeeded while the first was held")
	}
	release()
	again, err := lockCard(path, time.Second)
	if err != nil {
		t.Fatalf("lock after release: %v", err)
	}
	again()
}
