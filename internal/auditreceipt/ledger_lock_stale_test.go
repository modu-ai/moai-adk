package auditreceipt

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// Observe stale A, then retire it and publish B at the same name before the
// removal. B stays open to exercise Windows delete-pending publication too.
func TestLedgerLockStaleBreakKeepsSuccessor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.lock")
	if err := os.WriteFile(path, []byte("owner-A\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * ledgerLockStale)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	// File.Stat captures identity from the handle. Windows path Stat defers
	// loading file IDs until SameFile, when the old name may already be reused.
	originalHandle, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	original, statErr := originalHandle.Stat()
	if err := originalHandle.Close(); err != nil {
		t.Fatal(err)
	}
	if statErr != nil {
		t.Fatal(statErr)
	}
	var successor *os.File
	t.Cleanup(func() {
		if successor != nil {
			_ = successor.Close()
		}
	})
	observed := false
	if !breakStaleLedgerLockObserved(path, func() {
		observed = true
		if err := os.Rename(path, path+".retired"); err != nil {
			t.Fatalf("retire stale claim: %v", err)
		}
		var err error
		successor, err = atomicfile.CreateExclusive(path, 0o600)
		if err != nil {
			t.Fatalf("publish successor: %v", err)
		}
		if _, err := successor.WriteString("owner-B\n"); err != nil {
			t.Fatal(err)
		}
		current, err := successor.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if os.SameFile(original, current) {
			t.Fatal("successor reused the stale file identity")
		}
		if time.Since(current.ModTime()) > ledgerLockStale {
			t.Fatal("successor fixture is already stale")
		}
		t.Logf("stale-to-successor handoff: same-file=false old-mtime=%s successor-mtime=%s", original.ModTime(), current.ModTime())
	}) {
		t.Fatal("stale claim was not processed")
	}
	if !observed {
		t.Fatal("stale-claim handoff did not run")
	}
	if err := successor.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := atomicfile.ReadFile(path)
	if err != nil || string(data) != "owner-B\n" {
		t.Fatalf("stale break removed successor: data=%q err=%v", data, err)
	}
	third, err := atomicfile.CreateExclusive(path, 0o600)
	if third != nil {
		if closeErr := third.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		t.Fatal("third holder acquired successor's live claim")
	}
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("third holder error=%v, want existing claim", err)
	}
	releaseLedgerLock(path, "owner-B")
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("successor could not release own claim: %v", err)
	}
}

func TestLedgerLockStaleBreakPreservesFreshClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.lock")
	if err := os.WriteFile(path, []byte("fresh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if breakStaleLedgerLockObserved(path, func() { t.Fatal("fresh claim observed as stale") }) {
		t.Fatal("fresh claim reclaimed")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "fresh\n" {
		t.Fatalf("fresh claim changed: data=%q err=%v", data, err)
	}
}

func TestLedgerLockStaleBreakRemovesAbandonedClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.lock")
	if err := os.WriteFile(path, []byte("abandoned\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * ledgerLockStale)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	if !breakStaleLedgerLockObserved(path, nil) {
		t.Fatal("abandoned claim not reclaimed")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("abandoned claim remains: %v", err)
	}
	if breakStaleLedgerLockObserved(path, nil) {
		t.Fatal("missing claim reported reclaimed")
	}
}
