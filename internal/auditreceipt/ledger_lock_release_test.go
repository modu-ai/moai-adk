package auditreceipt

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// Schedule the stale holder's token read before a successor claims the same
// name. Keeping B open reproduces Windows delete-pending during publication.
func TestLedgerLockReleaseBindsRemovalToReadClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.lock")
	if err := os.WriteFile(path, []byte("owner-A\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var successor *os.File
	t.Cleanup(func() {
		if successor != nil {
			_ = successor.Close()
		}
	})
	releaseLedgerLockReading(path, "owner-A", func(r io.Reader) ([]byte, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "owner-A\n" {
			t.Fatalf("read claim = %q, want old owner", data)
		}
		// Rename retires A while preserving its open identity. Windows cannot
		// reuse a delete-pending name until the last old handle is closed.
		if err := os.Rename(path, path+".retired"); err != nil {
			t.Fatalf("retire old claim: %v", err)
		}
		successor, err = atomicfile.CreateExclusive(path, 0o600)
		if err != nil {
			t.Fatalf("create successor: %v", err)
		}
		if _, err := successor.WriteString("owner-B\n"); err != nil {
			t.Fatal(err)
		}
		return data, nil
	})
	if err := successor.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := atomicfile.ReadFile(path)
	if err != nil || string(data) != "owner-B\n" {
		t.Fatalf("stale release removed successor after publication: data=%q err=%v", data, err)
	}
	third, err := atomicfile.CreateExclusive(path, 0o600)
	if third != nil {
		if err := third.Close(); err != nil {
			t.Fatalf("close unexpected third claim: %v", err)
		}
		t.Fatal("third holder acquired a successor's live ledger claim")
	}
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("third claim error = %v, want existing owner", err)
	}
	releaseLedgerLock(path, "owner-B")
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("successor could not release its own claim: %v", err)
	}
}

func TestLedgerLockReleaseReadFailureKeepsClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.lock")
	if err := os.WriteFile(path, []byte("owner\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	releaseLedgerLockReading(path, "owner", func(io.Reader) ([]byte, error) {
		return []byte("owner\n"), io.ErrUnexpectedEOF
	})
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "owner\n" {
		t.Fatalf("failed read removed claim: data=%q err=%v", data, err)
	}
}
