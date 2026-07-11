package spec

import (
	"os"
	"testing"
)

// Tests for the four public drift entry points (SPEC-SESSIONSTART-PERF-001 M1).
//
// DetectDrift / DriftCount are the cached entry points (session-start critical path);
// DetectDriftFresh / DriftCountFresh are the authoritative --no-cache counterparts
// (REQ-SSP-006a). All four run against a real git fixture here, so they exercise the
// production deps (gitHeadSHA / cachedMainBranch / gitLogAllFullMessage) rather than
// the injected seam.

// setupEntryPointFixture builds a one-drifting-SPEC corpus in a real git repo.
func setupEntryPointFixture(t *testing.T) string {
	t.Helper()

	specs := []fixtureSpec{
		{id: "SPEC-ENTRY-001", status: "in-progress", era: "V3R6"},
	}
	commits := []fixtureCommit{
		{title: "feat(SPEC-ENTRY-001): M1 implementation"},
	}

	return setupDriftCorpusFixture(t, specs, commits)
}

// TestDriftCount_CachedEntryPoint covers the entry point the SessionStart hook calls.
func TestDriftCount_CachedEntryPoint(t *testing.T) {
	root := setupEntryPointFixture(t)

	count, err := DriftCount(root)
	if err != nil {
		t.Fatalf("DriftCount: %v", err)
	}
	if count != 1 {
		t.Errorf("DriftCount = %d, want 1", count)
	}

	// A second call at the same HEAD is served from the cache and must agree.
	cached, err := DriftCount(root)
	if err != nil {
		t.Fatalf("DriftCount (cached): %v", err)
	}
	if cached != count {
		t.Errorf("cached DriftCount = %d, want %d", cached, count)
	}
}

// TestDetectDriftFresh_WritesAndBypassesCache covers the --no-cache report path.
func TestDetectDriftFresh_WritesAndBypassesCache(t *testing.T) {
	root := setupEntryPointFixture(t)

	// Prime the cache through the cached entry point.
	if _, err := DetectDrift(root); err != nil {
		t.Fatalf("DetectDrift (priming): %v", err)
	}
	if _, err := os.Stat(driftCachePath(root)); err != nil {
		t.Fatalf("DetectDrift did not persist a cache file: %v", err)
	}

	// Corrupt the cached COUNT while keeping the HEAD key valid. A cached read would
	// return this bogus value; a fresh computation must ignore it.
	cached, ok := loadDriftCache(root, mustHeadSHA(t))
	if !ok {
		t.Fatal("loadDriftCache = miss, want hit after priming")
	}
	saveDriftCache(root, mustHeadSHA(t), &DriftReport{Records: cached.Records, Count: 9999})

	fresh, err := DetectDriftFresh(root)
	if err != nil {
		t.Fatalf("DetectDriftFresh: %v", err)
	}
	if fresh.Count != 1 {
		t.Errorf("DetectDriftFresh Count = %d, want 1 (the poisoned cache entry must be ignored)", fresh.Count)
	}
}

// TestDriftCountFresh_IgnoresPoisonedCache is the count-only counterpart.
func TestDriftCountFresh_IgnoresPoisonedCache(t *testing.T) {
	root := setupEntryPointFixture(t)

	if _, err := DetectDrift(root); err != nil {
		t.Fatalf("DetectDrift (priming): %v", err)
	}

	cached, ok := loadDriftCache(root, mustHeadSHA(t))
	if !ok {
		t.Fatal("loadDriftCache = miss, want hit after priming")
	}
	saveDriftCache(root, mustHeadSHA(t), &DriftReport{Records: cached.Records, Count: 9999})

	// The cached path happily serves the poisoned count...
	poisoned, err := DriftCount(root)
	if err != nil {
		t.Fatalf("DriftCount: %v", err)
	}
	if poisoned != 9999 {
		t.Fatalf("DriftCount = %d, want the poisoned 9999 (this test's premise is that the cache IS consulted)", poisoned)
	}

	// ...while the fresh path recomputes and returns the truth.
	count, err := DriftCountFresh(root)
	if err != nil {
		t.Fatalf("DriftCountFresh: %v", err)
	}
	if count != 1 {
		t.Errorf("DriftCountFresh = %d, want 1 (must bypass the cache)", count)
	}
}

// mustHeadSHA resolves HEAD in the current working directory, failing the test on error.
func mustHeadSHA(t *testing.T) string {
	t.Helper()

	head, err := gitHeadSHA()
	if err != nil {
		t.Fatalf("gitHeadSHA: %v", err)
	}
	return head
}
