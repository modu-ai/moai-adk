package conversation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// forkChildCount counts the fork child receipt directories that exist for the
// fixture's family, so a rejected fork can be proven to have created no child
// state.
func forkChildCount(t *testing.T, p Descriptor) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(filepath.Dir(p.ReceiptDir), "forks"))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return len(entries)
}

// TestForkPrefixCrossCheck pins the gateway-layer fork barrier contract
// (AC-MG-026 (c)): ForkSession accepts a --fork-session child only when the
// claimed inherited prefix matches the parent's completed chain
// (Manifest.ChainTo at the boundary). Tampered, unknown-boundary and
// unknown-origin claims are rejected before any child state exists — an
// implementation that accepts the caller-asserted value without contrast is
// red on the tamper variant.
func TestForkPrefixCrossCheck(t *testing.T) {
	ctx := context.Background()
	m, p, chain := boundaryForkFixture(t)
	boundary := chain[1].Prefix
	correct := receipt.ChainDigest(chain[:2])

	// The matching claim is accepted and the child carries exactly the
	// completed chain up to the boundary.
	f, err := m.ForkSession(ctx, p.UUID, boundary, correct)
	if err != nil {
		t.Fatal(err)
	}
	child, err := receipt.OpenStore(ctx, f.ReceiptDir, f.UUID, false)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := child.Snapshot(ctx)
	_ = child.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := snap.Candidates(); len(got) != 2 || got[0] != chain[0] || got[1] != chain[1] {
		t.Fatal("accepted fork carried the wrong boundary chain", got)
	}

	// The tampered claim is rejected before any child state exists.
	before := forkChildCount(t, p)
	if _, err = m.ForkSession(ctx, p.UUID, boundary, receipt.Hash([]byte("tampered-prefix"))); !errors.Is(err, ErrPrefixMismatch) {
		t.Fatalf("tampered prefix accepted: %v", err)
	}
	if after := forkChildCount(t, p); after != before {
		t.Fatalf("rejected fork left child state behind: before %d, after %d", before, after)
	}

	// An unknown boundary and an unknown origin are rejected without child
	// state as well.
	if _, err = m.ForkSession(ctx, p.UUID, receipt.Digest{}, correct); err == nil {
		t.Fatal("zero boundary accepted")
	}
	if _, err = m.ForkSession(ctx, "00000000-0000-4000-8000-00000000dead", boundary, correct); !errors.Is(err, ErrMissing) {
		t.Fatalf("unknown origin accepted: %v", err)
	}
	if after := forkChildCount(t, p); after != before {
		t.Fatalf("rejected fork left child state behind: before %d, after %d", before, after)
	}
}
