package guardliveness

import (
	"context"
	"strings"
	"testing"
	"time"
)

// renderBoundUnderTest mirrors the join bound the host surface declares
// (internal/hook: guardLivenessJoinBound). It is a literal here because the
// binding constant lives at the join site, which is a different package; a
// second declaration would be a second bound.
const renderBoundUnderTest = 250 * time.Millisecond

// AC-GDL-012 — the refresh is initiated, never awaited, and its result is
// persisted for a SUBSEQUENT activation to read.
//
// The fixture is the one where the two obligations conflict: subject queries
// slower than the render join bound. Clause (c) is what separates a correct
// implementation from the mutant that initiates, abandons at the bound, and
// DISCARDS — under which the render stays fast, the entailment looks intact,
// and the persisted verdict never advances.
func TestRefreshResultIsPersistedForALaterActivation(t *testing.T) {
	store := NewStore(t.TempDir())
	root := t.TempDir()

	// A verdict from an earlier activation, which is what this activation's
	// render must read.
	earlier := time.Now().Add(-2 * time.Hour)
	if err := store.Save(root, resultB(), earlier); err != nil {
		t.Fatalf("seed the earlier result: %v", err)
	}

	stalled := &blockingProducer{release: make(chan struct{})}
	ev := New(stalled, WithSink(store))

	// (a) the refresh is initiated.
	started := time.Now()
	refresh := ev.OnActivation(context.Background(), Activation{Root: root})
	initiation := time.Since(started)
	if refresh == nil {
		t.Fatal("activation yielded no refresh")
	}
	if initiation >= renderBoundUnderTest {
		t.Fatalf("initiating the refresh took %v, at or beyond the %v render join bound — the refresh is awaited", initiation, renderBoundUnderTest)
	}

	// (b) the render completes within the bound, from the PREVIOUSLY persisted
	// result, while the subject queries are still blocked.
	renderStart := time.Now()
	snap, err := store.Load(root)
	if err != nil {
		t.Fatalf("Load during a stalled refresh: %v", err)
	}
	text, _ := Advisory(snap, RenderRecord{}, time.Now())
	if elapsed := time.Since(renderStart); elapsed >= renderBoundUnderTest {
		t.Fatalf("render took %v, at or beyond the %v bound", elapsed, renderBoundUnderTest)
	}
	if !strings.Contains(text, "subject-8") {
		t.Fatalf("render did not come from the previously persisted result:\n%s", text)
	}

	// (c) when the stalled refresh completes, its result is persisted and is
	// what a subsequent activation reads.
	close(stalled.release)
	if _, err := refresh.Wait(); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	next, err := store.Load(root)
	if err != nil {
		t.Fatalf("Load after the refresh completed: %v", err)
	}
	if !next.TakenAt.After(earlier) {
		t.Fatalf("persisted TakenAt = %v, not after the seeded %v — the abandoned refresh never landed", next.TakenAt, earlier)
	}
	after, _ := Advisory(next, RenderRecord{}, time.Now())
	if !strings.Contains(after, "subject-2") || !strings.Contains(after, "subject-3") {
		t.Fatalf("the subsequent activation did not read the completed refresh's result:\n%s", after)
	}
}

// A refresh the producer could not answer must not overwrite the last verdict
// that was actually measured: replacing it with an empty result would report an
// all-clear about a set nothing evaluated.
func TestAFailedRefreshDoesNotOverwriteThePersistedResult(t *testing.T) {
	store := NewStore(t.TempDir())
	root := t.TempDir()
	takenAt := time.Now().Add(-time.Hour)
	if err := store.Save(root, resultA(), takenAt); err != nil {
		t.Fatalf("seed: %v", err)
	}

	ev := New(Unwired(), WithSink(store))
	if _, err := ev.OnActivation(context.Background(), Activation{Root: root}).Wait(); err == nil {
		t.Fatal("the unwired producer reported success")
	}

	snap, err := store.Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !snap.TakenAt.Equal(takenAt) {
		t.Fatalf("TakenAt = %v, want the seeded %v — a failed refresh replaced a measured verdict", snap.TakenAt, takenAt)
	}
}
