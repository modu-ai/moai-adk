package receipt

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const compactScope = "profile-generation-1"

// summary is the exact public summary text a compaction turn returned.
func summary(text string) []byte { return []byte(text) }

func authenticated(epoch uint64, text string) CompactBase {
	return CompactBase{Scope: compactScope, Epoch: epoch, Summary: Hash(summary(text))}
}

func TestRebaseContrastsExactSummaryDigestOncePerEpoch(t *testing.T) {
	l := NewRebaseLedger(compactScope, 0)
	const turn = "compacted summary of the whole prior conversation"
	if err := l.Rebase(authenticated(1, turn), summary(turn)); err != nil {
		t.Fatal("authenticated PostCompact rejected", err)
	}
	if l.Applied() != 1 {
		t.Fatal(l.Applied())
	}
	if err := l.Rebase(authenticated(1, turn), summary(turn)); !errors.Is(err, ErrCompactDuplicate) {
		t.Fatal("duplicate notice rebased twice", err)
	}
	// A summary that merely contains the authenticated text never passes.
	if err := l.Rebase(authenticated(2, turn+" plus trailing synthesis"), summary(turn)); !errors.Is(err, ErrCompactDigest) {
		t.Fatal("substring-only summary accepted", err)
	}
	if err := l.Rebase(authenticated(2, turn), summary(turn)); err != nil {
		t.Fatal("next epoch rejected", err)
	}
	if err := l.Rebase(CompactBase{Scope: "profile-generation-2", Epoch: 3, Summary: Hash(summary(turn))}, summary(turn)); !errors.Is(err, ErrCompactForeign) {
		t.Fatal("foreign scope accepted", err)
	}
	if err := l.Rebase(authenticated(4, turn), summary(turn)); !errors.Is(err, ErrCompactAmbiguous) {
		t.Fatal("lost-notice gap silently applied", err)
	}
	if l.Applied() != 2 {
		t.Fatal("rejected notices mutated the ledger", l.Applied())
	}
}

func TestRebaseLedgerRestartRejectsReplayedEpochs(t *testing.T) {
	l := NewRebaseLedger(compactScope, 3)
	const turn = "post-restart summary"
	// A replay of the pre-restart epoch must not rebase again.
	if err := l.Rebase(authenticated(3, turn), summary(turn)); !errors.Is(err, ErrCompactDuplicate) {
		t.Fatal("replayed epoch accepted after restart", err)
	}
	if err := l.Rebase(authenticated(2, turn), summary(turn)); !errors.Is(err, ErrCompactStale) {
		t.Fatal("stale epoch accepted after restart", err)
	}
	if err := l.Rebase(authenticated(4, turn), summary(turn)); err != nil {
		t.Fatal("post-restart epoch rejected", err)
	}
	if err := l.Rebase(CompactBase{Scope: compactScope, Epoch: 0, Summary: Digest{}}, nil); !errors.Is(err, ErrInvalid) {
		t.Fatal("zero epoch accepted", err)
	}
}

func TestManifestRebaseResetsPublicHistory(t *testing.T) {
	m, err := New(testUUID)
	if err != nil {
		t.Fatal(err)
	}
	c := candidate("p1", "root", "A")
	if err = m.Publish(c); err != nil {
		t.Fatal(err)
	}
	observation := Observation{Prefix: c.Prefix, Previous: c.Previous, Provider: c.Provider, Opaque: c.Opaque, Items: c.Items}
	if err = m.Check(testUUID, []Observation{observation}); err != nil {
		t.Fatal(err)
	}
	if err = m.Rebase(1); err != nil {
		t.Fatal(err)
	}
	if err = m.Check(testUUID, []Observation{observation}); !errors.Is(err, ErrInvalid) {
		t.Fatal("pre-compaction history still validated after the reset", err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	back, err := Parse(raw, testUUID)
	if err != nil || len(back.Candidates()) != 0 {
		t.Fatal("reset did not survive persistence", err, len(back.Candidates()))
	}
}

func TestStoreRebaseResetsDurableHistory(t *testing.T) {
	dir := filepath.Join(privateDir(t), "receipt")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	s, err := OpenStore(ctx, dir, testUUID, true)
	if err != nil {
		t.Fatal(err)
	}
	c := candidate("p1", "root", "A")
	if err = s.Publish(ctx, c); err != nil {
		t.Fatal(err)
	}
	if err = s.Rebase(ctx, 1); err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot(ctx)
	if err != nil || len(snap.Candidates()) != 0 {
		t.Fatal("durable history survived the reset", err, snap)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenStore(ctx, dir, testUUID, false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reopened.Close() }()
	snap, err = reopened.Snapshot(ctx)
	if err != nil || len(snap.Candidates()) != 0 {
		t.Fatal("reset lost across reopen", err, snap)
	}
}
