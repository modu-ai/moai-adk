package spec

import (
	"sync/atomic"
	"testing"
)

// countingTransitionRunner installs a fake git-history lookup that records how
// many times it was called, and restores the production hook at test end.
//
// The counter — not wall-clock time — is what these tests observe: the property
// under test is "how many history walks does one Lint() run perform?", and a
// timing assertion would measure the machine rather than the code.
func countingTransitionRunner(t *testing.T, rec *ownershipTransitionRecord) *int32 {
	t.Helper()
	var calls int32
	prev := getOwnershipTransitionRunner
	getOwnershipTransitionRunner = func(_, _ string) (*ownershipTransitionRecord, error) {
		atomic.AddInt32(&calls, 1)
		return rec, nil
	}
	t.Cleanup(func() { getOwnershipTransitionRunner = prev })
	return &calls
}

func fixtureTransitionRecord() *ownershipTransitionRecord {
	return &ownershipTransitionRecord{
		PreviousStatus:  "draft",
		CurrentStatus:   "completed",
		CommitSubject:   "chore(" + transitionFixtureID + "): move status",
		CommitSHA:       "0123456789abcdef",
		AuthoredByAgent: "manager-develop",
	}
}

// TestOwnershipTransitionLookupMemoizedPerRun pins the property F1 is about:
// the two rules that read the same document's status transition
// (OwnershipTransitionRule and StatusTransitionValidityRule) share ONE history
// walk per Lint() run, not one each.
//
// Before memoization this test observes 2 calls for a single-SPEC run.
func TestOwnershipTransitionLookupMemoizedPerRun(t *testing.T) {
	b := buildTransitionFixture(t, transitionFixture{
		from:              "draft",
		to:                "completed",
		trailer:           "manager-develop",
		outOfScopeHeading: true,
	})
	calls := countingTransitionRunner(t, fixtureTransitionRecord())

	lintFixture(t, b, false)

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("history lookups per Lint() run = %d, want 1 (both rules must share one walk)", got)
	}
}

// TestOwnershipTransitionLookupInvalidatedPerRun is the other half of the
// contract: the memo is scoped to one Lint(), so a second Lint() in the same
// process re-reads git rather than serving a stale record.
func TestOwnershipTransitionLookupInvalidatedPerRun(t *testing.T) {
	b := buildTransitionFixture(t, transitionFixture{
		from:              "draft",
		to:                "completed",
		trailer:           "manager-develop",
		outOfScopeHeading: true,
	})
	calls := countingTransitionRunner(t, fixtureTransitionRecord())

	lintFixture(t, b, false)
	lintFixture(t, b, false)

	if got := atomic.LoadInt32(calls); got != 2 {
		t.Errorf("history lookups across two Lint() runs = %d, want 2 (per-run invalidation)", got)
	}
}

// TestOwnershipTransitionLookupUncachedOutsideLint pins the direct-caller
// behavior: with no per-run cache active, every call reaches git, exactly as
// before memoization. This is what keeps callers outside Lint() unchanged.
func TestOwnershipTransitionLookupUncachedOutsideLint(t *testing.T) {
	calls := countingTransitionRunner(t, fixtureTransitionRecord())

	for i := 0; i < 3; i++ {
		if _, err := cachedOwnershipTransition("a/spec.md", "SPEC-X-001"); err != nil {
			t.Fatalf("cachedOwnershipTransition: %v", err)
		}
	}

	if got := atomic.LoadInt32(calls); got != 3 {
		t.Errorf("lookups with no cache active = %d, want 3 (uncached passthrough)", got)
	}
}

// TestOwnershipTransitionMemoKeyedPerDocument guards the memo key: two distinct
// SPEC documents must not share one cached record. A key collision here would
// make every SPEC in a corpus inherit the first one's transition — a silent,
// corpus-wide wrong answer rather than a visible failure.
func TestOwnershipTransitionMemoKeyedPerDocument(t *testing.T) {
	var calls int32
	prev := getOwnershipTransitionRunner
	getOwnershipTransitionRunner = func(specPath, specID string) (*ownershipTransitionRecord, error) {
		atomic.AddInt32(&calls, 1)
		return &ownershipTransitionRecord{
			PreviousStatus: "draft",
			CurrentStatus:  specID, // echo the key back so a collision is visible
			CommitSHA:      specPath,
		}, nil
	}
	t.Cleanup(func() { getOwnershipTransitionRunner = prev })

	startGitQueryCache()
	defer stopGitQueryCache()

	first, err := cachedOwnershipTransition("a/spec.md", "SPEC-A-001")
	if err != nil {
		t.Fatalf("cachedOwnershipTransition: %v", err)
	}
	second, err := cachedOwnershipTransition("b/spec.md", "SPEC-B-001")
	if err != nil {
		t.Fatalf("cachedOwnershipTransition: %v", err)
	}
	again, err := cachedOwnershipTransition("a/spec.md", "SPEC-A-001")
	if err != nil {
		t.Fatalf("cachedOwnershipTransition: %v", err)
	}

	if first.CurrentStatus != "SPEC-A-001" || second.CurrentStatus != "SPEC-B-001" {
		t.Errorf("memo keys collided: first=%q second=%q", first.CurrentStatus, second.CurrentStatus)
	}
	if again.CurrentStatus != "SPEC-A-001" {
		t.Errorf("repeat lookup returned %q, want SPEC-A-001", again.CurrentStatus)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("lookups for 2 distinct documents (3 calls) = %d, want 2", got)
	}
}
