package receipt

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const restorationScope = "profile-generation-restoration"

// TestRebaseLedgerRestoration pins the production applied-epoch restoration
// contract (AC-MG-026 (b)): a rebase records its epoch in the same durable
// manifest transaction, a reopened session reads the recorded value back into
// a fresh ledger, missing state starts at zero, and corrupt state is an
// explicit error. The exact-value principle leaves no room for a generous
// (stale-rejecting) or low (double-rebasing) approximation.
func TestRebaseLedgerRestoration(t *testing.T) {
	ctx := context.Background()
	const turn = "post-restart summary"
	dir := filepath.Join(privateDir(t), "receipt")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	s, err := OpenStore(ctx, dir, testUUID, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Publish(ctx, candidate("p1", "root", "A")); err != nil {
		t.Fatal(err)
	}
	// Rebase the durable store at epoch 3: the ledger contrast happened
	// upstream; the store applies the reset and must record the epoch in the
	// same manifest write.
	if err = s.Rebase(ctx, 3); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}

	// A reopened session restores the ledger from the recorded value — the
	// caller never guesses it.
	reopened, err := OpenStore(ctx, dir, testUUID, false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reopened.Close() }()
	restored, err := reopened.RestoreRebaseLedger(ctx, restorationScope)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Applied() != 3 {
		t.Fatalf("restored ledger applied epoch = %d, want the recorded 3", restored.Applied())
	}

	// The first rebase after restoration is judged against the restored
	// value: a replay of the pre-restart epoch is a duplicate, the next epoch
	// proceeds.
	if err = restored.Rebase(CompactBase{Scope: restorationScope, Epoch: 3, Summary: Hash(summary(turn))}, summary(turn)); !errors.Is(err, ErrCompactDuplicate) {
		t.Fatal("restored ledger accepted a pre-restart epoch replay", err)
	}
	if err = restored.Rebase(CompactBase{Scope: restorationScope, Epoch: 4, Summary: Hash(summary(turn))}, summary(turn)); err != nil {
		t.Fatal("post-restoration epoch rejected", err)
	}

	// Missing state (a fresh scope that never rebased) restores zero.
	freshRoot := filepath.Join(privateDir(t), "fresh")
	if err = os.MkdirAll(freshRoot, 0700); err != nil {
		t.Fatal(err)
	}
	fresh, err := OpenStore(ctx, freshRoot, testUUID, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fresh.Close() }()
	zero, err := fresh.RestoreRebaseLedger(ctx, restorationScope)
	if err != nil {
		t.Fatal(err)
	}
	if zero.Applied() != 0 {
		t.Fatalf("fresh scope applied epoch = %d, want 0", zero.Applied())
	}

	// Corrupt state is an explicit error, never a fallback value.
	corruptDir := filepath.Join(privateDir(t), "corrupt")
	if err = os.MkdirAll(corruptDir, 0700); err != nil {
		t.Fatal(err)
	}
	corrupt, err := OpenStore(ctx, corruptDir, testUUID, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = corrupt.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(corruptDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw[0] ^= 0x7f
	if err = os.WriteFile(filepath.Join(corruptDir, "manifest.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	broken, openErr := OpenStore(ctx, corruptDir, testUUID, false)
	if openErr != nil {
		// Failing explicitly at open already satisfies the contract: corrupt
		// state never yields a ledger.
		t.Logf("corrupt state rejected at open: %v", openErr)
		return
	}
	defer func() { _ = broken.Close() }()
	if _, err = broken.RestoreRebaseLedger(ctx, restorationScope); err == nil {
		t.Fatal("corrupt receipt state restored a ledger instead of failing explicitly")
	}
}
