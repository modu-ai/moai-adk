package conversation

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// boundaryForkFixture completes a family and publishes a three-turn receipt
// chain into its store, returning the manager and the chain.
func boundaryForkFixture(t *testing.T) (*Manager, Descriptor, []receipt.Candidate) {
	t.Helper()
	m := newTestManager(t)
	p := mustCompleted(t, m, "p", 1)
	s, err := receipt.OpenStore(context.Background(), p.ReceiptDir, p.UUID, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	previous := receipt.Hash([]byte("root"))
	chain := make([]receipt.Candidate, 0, 3)
	for _, label := range []string{"c1", "c2", "c3"} {
		c := receipt.Candidate{Prefix: receipt.Hash([]byte(label)), Previous: previous, Provider: "openai", Opaque: receipt.Hash([]byte(label + "-opaque")), Items: 1, Required: true, Complete: true}
		if err = s.Publish(context.Background(), c); err != nil {
			t.Fatal(err)
		}
		chain = append(chain, c)
		previous = c.Prefix
	}
	return m, p, chain
}

func TestManagerForkAtBoundaryCarriesOnlyPreBoundaryFacts(t *testing.T) {
	m, p, chain := boundaryForkFixture(t)
	f, err := m.ForkAt(context.Background(), p.UUID, chain[1].Prefix)
	if err != nil {
		t.Fatal(err)
	}
	if f.UUID == p.UUID || f.FamilyID != p.FamilyID || len(f.Args) != 5 {
		t.Fatalf("fork=%+v", f)
	}
	child, err := receipt.OpenStore(context.Background(), f.ReceiptDir, f.UUID, false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Close() }()
	snap, err := child.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := snap.Candidates()
	if len(got) != 2 || got[0] != chain[0] || got[1] != chain[1] {
		t.Fatal("child carried the wrong boundary chain", got)
	}
	// The parent keeps completing turns after the fork; the child never sees them.
	post := receipt.Observation{Prefix: chain[2].Prefix, Previous: chain[1].Prefix, Provider: chain[2].Provider, Opaque: chain[2].Opaque, Items: chain[2].Items}
	if err = snap.Check(f.UUID, []receipt.Observation{post}); !errors.Is(err, receipt.ErrInvalid) {
		t.Fatal("post-boundary fact validated in the child", err)
	}
}

func TestManagerForkAtRejectionGroups(t *testing.T) {
	m, p, chain := boundaryForkFixture(t)
	if _, err := m.ForkAt(context.Background(), "00000000-0000-4000-8000-00000000dead", chain[1].Prefix); !errors.Is(err, ErrMissing) {
		t.Fatal("unknown origin accepted", err)
	}
	if _, err := m.ForkAt(context.Background(), p.UUID, receipt.Hash([]byte("future-turn"))); err == nil {
		t.Fatal("unknown boundary accepted")
	}
	if _, err := m.ForkAt(context.Background(), p.UUID, receipt.Digest{}); err == nil {
		t.Fatal("zero boundary accepted")
	}
}
