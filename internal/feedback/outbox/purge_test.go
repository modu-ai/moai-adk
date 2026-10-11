package outbox

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestPurgePropagatesSpoolRemovalFailure pins the purge contract: when the
// spool store cannot be removed, PurgeStores reports the failure — never a
// false success with the store still on disk (review-gate finding #7). The
// fixture makes the spool path a non-empty DIRECTORY: os.Remove refuses a
// directory that still has content, on every platform, while the sibling
// stores (files in the same directory) remove cleanly.
func TestPurgePropagatesSpoolRemovalFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)

	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	// The spool survives as a non-empty directory: ClearSpool's remove fails.
	spool := filepath.Join(dir, "spool.jsonl")
	if err := os.Mkdir(spool, 0o700); err != nil {
		t.Fatalf("mkdir spool fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(spool, "entry.jsonl"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("fill spool fixture: %v", err)
	}

	err := PurgeStores()
	if err == nil {
		t.Fatal("PurgeStores reported success while the spool store survived (its removal error was ignored)")
	}
	if _, statErr := os.Stat(spool); statErr != nil {
		t.Fatalf("the spool fixture vanished: %v", statErr)
	}
}
