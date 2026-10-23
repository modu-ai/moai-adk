//go:build !windows

// audit_cache_state_fifo_unix_test.go — card t1632 card-review P2-B: the
// verdict file is the cache entry's evidence, so only a REGULAR file may be
// read as that evidence. A FIFO (or any non-regular target) read with
// os.ReadFile blocks forever; the store must refuse it promptly.
package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// storeWithin runs StoreVerdict in a goroutine so that a regression fails
// the test at the deadline instead of hanging the suite.
func storeWithin(t *testing.T, c *StateAuditCache, report string, d time.Duration) (error, bool) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- c.StoreVerdict("SPEC-X-001", "hash-1", passResult(report)) }()
	select {
	case err := <-done:
		return err, true
	case <-time.After(d):
		return nil, false
	}
}

func TestStateCache_StoreRefusesFIFOReportPromptly(t *testing.T) {
	root := t.TempDir()
	fifo := filepath.Join(root, "report.fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	c := NewStateAuditCache(root)
	err, returned := storeWithin(t, c, fifo, 3*time.Second)
	if !returned {
		t.Fatal("StoreVerdict blocked on a FIFO report_path — a non-regular verdict file must be refused, not read")
	}
	if err == nil || !strings.Contains(err.Error(), "verdict file evidence") {
		t.Fatalf("StoreVerdict(FIFO) err = %v, want a `verdict file evidence: ...` refusal", err)
	}
	if _, ok := c.Lookup("SPEC-X-001", "hash-1", time.Now()); ok {
		t.Fatal("a FIFO report produced a cache entry")
	}
}

func TestStateCache_StoreRefusesSymlinkedReport(t *testing.T) {
	root := t.TempDir()
	target := writeVerdictFile(t, root, "real.md", "# pass")
	link := filepath.Join(root, "link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	c := NewStateAuditCache(root)
	err, returned := storeWithin(t, c, link, 3*time.Second)
	if !returned {
		t.Fatal("StoreVerdict blocked on a symlinked report_path")
	}
	if err == nil || !strings.Contains(err.Error(), "verdict file evidence") {
		t.Fatalf("StoreVerdict(symlink) err = %v, want a `verdict file evidence: ...` refusal (a link is never opened, as in internal/auditreceipt)", err)
	}
}

func TestStateCache_StoreRefusesOversizedReport(t *testing.T) {
	root := t.TempDir()
	big := filepath.Join(root, "big.md")
	if err := os.WriteFile(big, make([]byte, maxVerdictFileBytes+1), 0o644); err != nil {
		t.Fatal(err)
	}
	c := NewStateAuditCache(root)
	err, returned := storeWithin(t, c, big, 3*time.Second)
	if !returned {
		t.Fatal("StoreVerdict did not return for an oversized report")
	}
	if err == nil || !strings.Contains(err.Error(), "verdict file evidence") {
		t.Fatalf("StoreVerdict(oversized) err = %v, want a `verdict file evidence: ...` refusal", err)
	}
}

func TestStateCache_StoreStillAcceptsRegularReport(t *testing.T) {
	root := t.TempDir()
	report := writeVerdictFile(t, root, "review.md", "# plan review: pass")
	c := NewStateAuditCache(root)
	err, returned := storeWithin(t, c, report, 3*time.Second)
	if !returned || err != nil {
		t.Fatalf("StoreVerdict(regular file) returned=%v err=%v, want a normal store", returned, err)
	}
	if _, ok := c.Lookup("SPEC-X-001", "hash-1", time.Now()); !ok {
		t.Fatal("a regular report did not produce a lookup hit")
	}
}
