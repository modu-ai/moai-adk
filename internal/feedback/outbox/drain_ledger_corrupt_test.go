package outbox

// The ledger read-corruption test (card-review finding, P2): a CORRUPTED
// ledger (JSON parse failure) used to read as EMPTY history — dedupe and
// the send caps reset, so extra reports enqueued past an exhausted daily
// cap. Only file ABSENCE means empty; a read or parse error is an error,
// which the drain already treats as retryable: the item and the spool tail
// stay for the next drain (preserving the spool) instead of publishing on
// amnesia.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

func TestCorruptedLedgerIsNotSilentlyEmpty(t *testing.T) {
	spoolPath := spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	ledgerPath, err := StorePath(LedgerFileName)
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if err := os.WriteFile(ledgerPath, []byte("{not json at all"), 0o600); err != nil {
		t.Fatalf("seed corrupt ledger: %v", err)
	}

	if _, lerr := loadLedger(); lerr == nil {
		t.Fatal("loadLedger accepted a corrupted ledger as empty history — the caps and the dedupe window reset on amnesia")
	}

	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := queuedBugreportCount(t); got != 0 {
		t.Fatalf("queue holds %d items — the corrupted ledger read as empty and the report enqueued past the dedupe/caps state", got)
	}
	if remaining := spoolKinds(t, spoolPath); len(remaining) != 1 {
		t.Fatalf("the spool did not survive the unreadable-ledger drain (remaining: %v) — the failure must be retryable with the spool preserved", remaining)
	}
	_ = filepath.Join
}
