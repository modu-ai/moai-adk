//go:build windows

package auditreceipt

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLedgerLockReleaseHandlesLongNativePath(t *testing.T) {
	root := filepath.Join(t.TempDir(), strings.Repeat("long-segment-", 8), strings.Repeat("long-segment-", 8))
	path := filepath.Join(root, "ledger.json")
	if len(path+ledgerLockSuffix) < 248 {
		t.Fatal("fixture did not exercise extended Windows path handling")
	}
	release, err := lockLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	release()
	if _, err := os.Stat(path + ledgerLockSuffix); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("long-path owner release left its claim: %v", err)
	}
}

func TestLedgerLockReleaseWaitsForTransientReader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.lock")
	if err := os.WriteFile(path, []byte("owner\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path) // Go readers do not share DELETE access.
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	closed := make(chan error, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		closed <- reader.Close()
	}()
	releaseLedgerLock(path, "owner")
	if err := <-closed; err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("transient reader prevented owner release: %v", err)
	}
}
